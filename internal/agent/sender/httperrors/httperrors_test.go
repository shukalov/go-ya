package httperrors

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/shukalov/go-ya/internal/retry"
)

type testNetError struct {
	timeout   bool
	temporary bool
}

func (e *testNetError) Error() string   { return "test net error" }
func (e *testNetError) Timeout() bool   { return e.timeout }
func (e *testNetError) Temporary() bool { return e.temporary }

func TestClassify_NetError(t *testing.T) {
	classifier := NewHTTPErrorClassifier()

	err := &testNetError{timeout: false}
	c := classifier.Classify(err)
	if !c.Retriable {
		t.Error("net.Error should be Retriable")
	}
	if c.Strategy != retry.Linear {
		t.Errorf("expected strategy Linear, got %d", c.Strategy)
	}
}

func TestClassify_NetTimeoutError(t *testing.T) {
	classifier := NewHTTPErrorClassifier()

	err := &testNetError{timeout: true}
	c := classifier.Classify(err)
	if !c.Retriable {
		t.Error("net.Error (timeout) should be Retriable")
	}
	if c.Strategy != retry.Linear {
		t.Errorf("expected strategy Linear, got %d", c.Strategy)
	}
}

func TestClassify_NonNetError(t *testing.T) {
	classifier := NewHTTPErrorClassifier()

	err := errors.New("some generic error")
	c := classifier.Classify(err)
	if c.Retriable {
		t.Error("non-net error should be NonRetriable")
	}
}

func TestClassify_NilError(t *testing.T) {
	classifier := NewHTTPErrorClassifier()
	c := classifier.Classify(nil)
	if c.Retriable {
		t.Error("nil error should be NonRetriable")
	}
}

func TestClassify_WrappedNetError(t *testing.T) {
	classifier := NewHTTPErrorClassifier()

	var netErr net.Error = &testNetError{timeout: false}
	err := fmt.Errorf("wrapped: %w", netErr)
	c := classifier.Classify(err)
	if !c.Retriable {
		t.Error("wrapped net.Error should be Retriable")
	}
}

func TestIsRetriable(t *testing.T) {
	classifier := NewHTTPErrorClassifier()

	netErr := &testNetError{timeout: false}
	if !classifier.IsRetriable(netErr) {
		t.Error("net.Error should be retriable")
	}

	err := errors.New("some error")
	if classifier.IsRetriable(err) {
		t.Error("non-net error should not be retriable")
	}

	if classifier.IsRetriable(nil) {
		t.Error("nil should not be retriable")
	}
}
