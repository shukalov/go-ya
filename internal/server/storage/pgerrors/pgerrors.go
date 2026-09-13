package pgerrors

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shukalov/go-ya/internal/retry"
)

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) IsRetriable(err error) bool {
	return c.Classify(err).Retriable
}

func (c *PostgresErrorClassifier) Classify(err error) retry.Classification {
	if err == nil {
		return retry.Classification{Retriable: false}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return classifyPgError(pgErr)
	}

	return retry.Classification{Retriable: false}
}

func classifyPgError(pgErr *pgconn.PgError) retry.Classification {
	switch pgErr.Code {
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure,
		pgerrcode.TransactionRollback,
		pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected,
		pgerrcode.CannotConnectNow:
		return retry.Classification{Retriable: true, Strategy: retry.Linear}
	}

	return retry.Classification{Retriable: false}
}
