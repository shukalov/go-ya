package storage

import (
	"context"
	"database/sql"

	"github.com/shukalov/go-ya/internal/retry"
	"github.com/shukalov/go-ya/internal/server/storage/pgerrors"
	"github.com/shukalov/go-ya/pkg/models"
)

type dbPersist struct {
	db         *sql.DB
	classifier *pgerrors.PostgresErrorClassifier
}

func newDBPersist(db *sql.DB) *dbPersist {
	return &dbPersist{
		db:         db,
		classifier: pgerrors.NewPostgresErrorClassifier(),
	}
}

func (p *dbPersist) retryCfg() retry.Config {
	return retry.Config{
		StrategyCfg: retry.DefaultConfig.StrategyCfg,
		Classify:  p.classifier.Classify,
	}
}

func (p *dbPersist) Save(metrics models.RuntimeMetrics) error {
	return retry.Execute(context.Background(), p.retryCfg(), func() error {
		ctx := context.Background()

		tx, err := p.db.BeginTx(ctx, nil)
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

		return tx.Commit()
	})
}

func (p *dbPersist) Load(metrics *models.RuntimeMetrics) error {
	return retry.Execute(context.Background(), p.retryCfg(), func() error {
		ctx := context.Background()

		metrics.Gauges = make(map[string]float64)
		metrics.Counters = make(map[string]int64)

		rows, err := p.db.QueryContext(ctx, `SELECT id, value FROM gauges`)
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
			metrics.Gauges[id] = value
		}
		if err := rows.Err(); err != nil {
			return err
		}

		rows, err = p.db.QueryContext(ctx, `SELECT id, delta FROM counters`)
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
			metrics.Counters[id] = delta
		}
		if err := rows.Err(); err != nil {
			return err
		}

		return nil
	})
}
