package httperrors

import (
	"errors"
	"net"

	"github.com/shukalov/go-ya/internal/retry"
)

type HTTPErrorClassifier struct{}

func NewHTTPErrorClassifier() *HTTPErrorClassifier {
	return &HTTPErrorClassifier{}
}

func (c *HTTPErrorClassifier) IsRetriable(err error) bool {
	return c.Classify(err).Retriable
}

func (c *HTTPErrorClassifier) Classify(err error) retry.Classification {
	if err == nil {
		return retry.Classification{Retriable: false}
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return retry.Classification{Retriable: true, Strategy: retry.Linear}
	}

	return retry.Classification{Retriable: false}
}
