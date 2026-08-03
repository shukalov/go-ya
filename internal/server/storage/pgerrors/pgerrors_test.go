package pgerrors

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shukalov/go-ya/internal/retry"
)

func TestClassify_ConnectionException(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	tests := []struct {
		name     string
		code     string
		strategy retry.Strategy
	}{
		{"ConnectionException", pgerrcode.ConnectionException, retry.Linear},
		{"ConnectionDoesNotExist", pgerrcode.ConnectionDoesNotExist, retry.Linear},
		{"ConnectionFailure", pgerrcode.ConnectionFailure, retry.Linear},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgErr := &pgconn.PgError{Code: tt.code}
			c := classifier.Classify(pgErr)
			if !c.Retriable {
				t.Errorf("%s should be Retriable", tt.name)
			}
			if c.Strategy != tt.strategy {
				t.Errorf("%s: expected strategy %d, got %d", tt.name, tt.strategy, c.Strategy)
			}
		})
	}
}

func TestClassify_TransactionRollback(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	tests := []struct {
		name     string
		code     string
		strategy retry.Strategy
	}{
		{"TransactionRollback", pgerrcode.TransactionRollback, retry.Linear},
		{"SerializationFailure", pgerrcode.SerializationFailure, retry.Linear},
		{"DeadlockDetected", pgerrcode.DeadlockDetected, retry.Linear},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgErr := &pgconn.PgError{Code: tt.code}
			c := classifier.Classify(pgErr)
			if !c.Retriable {
				t.Errorf("%s should be Retriable", tt.name)
			}
			if c.Strategy != tt.strategy {
				t.Errorf("%s: expected strategy %d, got %d", tt.name, tt.strategy, c.Strategy)
			}
		})
	}
}

func TestClassify_CannotConnectNow(t *testing.T) {
	classifier := NewPostgresErrorClassifier()
	pgErr := &pgconn.PgError{Code: pgerrcode.CannotConnectNow}
	c := classifier.Classify(pgErr)
	if !c.Retriable {
		t.Error("CannotConnectNow should be Retriable")
	}
	if c.Strategy != retry.Linear {
		t.Errorf("expected strategy Linear, got %d", c.Strategy)
	}
}

func TestClassify_NonRetriable(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	tests := []struct {
		name string
		code string
	}{
		{"UniqueViolation", pgerrcode.UniqueViolation},
		{"NotNullViolation", pgerrcode.NotNullViolation},
		{"ForeignKeyViolation", pgerrcode.ForeignKeyViolation},
		{"SyntaxError", pgerrcode.SyntaxError},
		{"UndefinedTable", pgerrcode.UndefinedTable},
		{"DataException", pgerrcode.DataException},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgErr := &pgconn.PgError{Code: tt.code}
			c := classifier.Classify(pgErr)
			if c.Retriable {
				t.Errorf("%s should be NonRetriable", tt.name)
			}
		})
	}
}

func TestClassify_NilError(t *testing.T) {
	classifier := NewPostgresErrorClassifier()
	c := classifier.Classify(nil)
	if c.Retriable {
		t.Error("nil error should be NonRetriable")
	}
}

func TestClassify_NonPgError(t *testing.T) {
	classifier := NewPostgresErrorClassifier()
	err := errors.New("some generic error")
	c := classifier.Classify(err)
	if c.Retriable {
		t.Error("non-PG error should be NonRetriable")
	}
}

func TestIsRetriable(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	if !classifier.IsRetriable(pgErr) {
		t.Error("ConnectionException should be retriable")
	}

	pgErr = &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	if classifier.IsRetriable(pgErr) {
		t.Error("UniqueViolation should not be retriable")
	}

	if classifier.IsRetriable(nil) {
		t.Error("nil should not be retriable")
	}
}
