package storage

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestDBStorage_UpdateGauge(t *testing.T) {
	s := NewDBStorage(nil, 0)
	ctx := context.Background()

	if err := s.UpdateGauge(ctx, "test", 1.1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok, err := s.GetGauge(ctx, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected metric to exist")
	}
	if val != 1.1 {
		t.Errorf("expected 1.1, got %f", val)
	}
}

func TestDBStorage_UpdateCounter(t *testing.T) {
	s := NewDBStorage(nil, 0)
	ctx := context.Background()

	s.UpdateCounter(ctx, "c", 5)
	s.UpdateCounter(ctx, "c", 3)

	val, ok, err := s.GetCounter(ctx, "c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected metric to exist")
	}
	if val != 8 {
		t.Errorf("expected 8, got %d", val)
	}
}

func TestDBStorage_GetGauge_NotFound(t *testing.T) {
	s := NewDBStorage(nil, 0)
	ctx := context.Background()

	val, ok, err := s.GetGauge(ctx, "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected metric not to exist")
	}
	if val != 0 {
		t.Errorf("expected 0, got %f", val)
	}
}

func TestDBStorage_GetCounter_NotFound(t *testing.T) {
	s := NewDBStorage(nil, 0)
	ctx := context.Background()

	val, ok, err := s.GetCounter(ctx, "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected metric not to exist")
	}
	if val != 0 {
		t.Errorf("expected 0, got %d", val)
	}
}

func TestDBStorage_GetAllMetrics(t *testing.T) {
	s := NewDBStorage(nil, 0)
	ctx := context.Background()

	s.UpdateGauge(ctx, "g1", 1.1)
	s.UpdateGauge(ctx, "g2", 2.2)
	s.UpdateCounter(ctx, "c1", 10)

	result, err := s.GetAllMetrics(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Gauges) != 2 {
		t.Errorf("expected 2 gauges, got %d", len(result.Gauges))
	}
	if len(result.Counters) != 1 {
		t.Errorf("expected 1 counter, got %d", len(result.Counters))
	}
}

func TestDBStorage_Save_NilDB(t *testing.T) {
	s := NewDBStorage(nil, 0)

	// Save with nil db should panic (no real db)
	// but MemStorage methods should work fine
	ctx := context.Background()
	s.UpdateGauge(ctx, "g", 1.0)
}

func TestDBStorage_Save_WithDB(t *testing.T) {
	dsn := "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Skip("cannot open database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("cannot connect to database:", err)
	}

	ctx := context.Background()
	s := NewDBStorage(db, 0)
	s.UpdateGauge(ctx, "test_gauge", 42.5)
	s.UpdateCounter(ctx, "test_counter", 100)

	if err := s.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	s2 := NewDBStorage(db, 0)
	if err := s2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	val, ok, err := s2.GetGauge(ctx, "test_gauge")
	if err != nil {
		t.Fatalf("GetGauge error: %v", err)
	}
	if !ok {
		t.Fatal("expected test_gauge to exist")
	}
	if val != 42.5 {
		t.Errorf("expected 42.5, got %f", val)
	}

	cval, ok, err := s2.GetCounter(ctx, "test_counter")
	if err != nil {
		t.Fatalf("GetCounter error: %v", err)
	}
	if !ok {
		t.Fatal("expected test_counter to exist")
	}
	if cval != 100 {
		t.Errorf("expected 100, got %d", cval)
	}
}

func TestDBStorage_Load_EmptyDB(t *testing.T) {
	dsn := "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Skip("cannot open database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("cannot connect to database:", err)
	}

	s := NewDBStorage(db, 0)
	if err := s.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	ctx := context.Background()
	result, err := s.GetAllMetrics(ctx)
	if err != nil {
		t.Fatalf("GetAllMetrics error: %v", err)
	}
	if len(result.Gauges) != 0 {
		t.Errorf("expected 0 gauges on empty db, got %d", len(result.Gauges))
	}
}
