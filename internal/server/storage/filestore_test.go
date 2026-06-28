package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileBackedStorage_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	fs := NewFileBackedStorage(path, 0)
	fs.UpdateGauge("gauge1", 1.1)
	fs.UpdateGauge("gauge2", 2.2)
	fs.UpdateCounter("counter1", 10)

	if err := fs.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	fs2 := NewFileBackedStorage(path, 0)
	if err := fs2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	val, ok, _ := fs2.GetGauge("gauge1")
	if !ok || val != 1.1 {
		t.Errorf("expected gauge1=1.1, got %v, %v", val, ok)
	}

	val2, ok2, _ := fs2.GetCounter("counter1")
	if !ok2 || val2 != 10 {
		t.Errorf("expected counter1=10, got %v, %v", val2, ok2)
	}

	all, _ := fs2.GetAllMetrics()
	if len(all.Gauges) != 2 {
		t.Errorf("expected 2 gauges, got %d", len(all.Gauges))
	}
	if len(all.Counters) != 1 {
		t.Errorf("expected 1 counter, got %d", len(all.Counters))
	}
}

func TestFileBackedStorage_Load_NonExistent(t *testing.T) {
	fs := NewFileBackedStorage("/nonexistent/path.json", 0)
	if err := fs.Load(); err != nil {
		t.Errorf("expected nil for missing file, got %v", err)
	}
}

func TestFileBackedStorage_Load_Corrupted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupted.json")

	os.WriteFile(path, []byte("not valid json"), 0644)

	fs := NewFileBackedStorage(path, 0)
	if err := fs.Load(); err == nil {
		t.Error("expected error for corrupted file, got nil")
	}
}

func TestFileBackedStorage_Save_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")

	fs := NewFileBackedStorage(path, 0)
	if err := fs.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Error("expected non-empty file")
	}
}

func TestFileBackedStorage_Load_RestoresGaugeOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gauge_only.json")

	fs := NewFileBackedStorage(path, 0)
	fs.UpdateGauge("only_gauge", 42.5)
	fs.Save()

	fs2 := NewFileBackedStorage(path, 0)
	fs2.Load()

	val, ok, _ := fs2.GetGauge("only_gauge")
	if !ok || val != 42.5 {
		t.Errorf("expected only_gauge=42.5, got %v, %v", val, ok)
	}

	_, ok, _ = fs2.GetCounter("only_gauge")
	if ok {
		t.Error("expected only_gauge to not be a counter")
	}
}
