package storage

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"

	serverlogger "github.com/shukalov/go-ya/internal/server/logger"
)

type DBStorage struct {
	MemStorage
	db            *sql.DB
	storeInterval time.Duration
}

func NewDBStorage(db *sql.DB, storeInterval time.Duration) *DBStorage {
	return &DBStorage{
		MemStorage:    *NewMemStorage(),
		db:            db,
		storeInterval: storeInterval,
	}
}

func (s *DBStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	s.MemStorage.UpdateGauge(ctx, name, value)
	return nil
}

func (s *DBStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	s.MemStorage.UpdateCounter(ctx, name, value)
	return nil
}

func (s *DBStorage) Save() error {
	ctx := context.Background()

	metrics, err := s.MemStorage.GetAllMetrics(ctx)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for name, value := range metrics.Gauges {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO gauges (id, value) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value`,
			name, value,
		)
		if err != nil {
			return err
		}
	}

	for name, delta := range metrics.Counters {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO counters (id, delta) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET delta = EXCLUDED.delta`,
			name, delta,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	serverlogger.L.Info("saved metrics to database")
	return nil
}

func (s *DBStorage) Load() error {
	ctx := context.Background()

	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.QueryContext(ctx, `SELECT id, value FROM gauges`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var value float64
		if err := rows.Scan(&id, &value); err != nil {
			return err
		}
		s.gauges[id] = value
	}
	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = s.db.QueryContext(ctx, `SELECT id, delta FROM counters`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var delta int64
		if err := rows.Scan(&id, &delta); err != nil {
			return err
		}
		s.counters[id] = delta
	}
	if err := rows.Err(); err != nil {
		return err
	}

	serverlogger.L.Info("loaded metrics from database")
	return nil
}

func (s *DBStorage) Run() {
	if s.storeInterval <= 0 {
		return
	}

	ticker := time.NewTicker(s.storeInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.Save(); err != nil {
			serverlogger.L.Error("periodic save to database failed", zap.Error(err))
		}
	}
}
