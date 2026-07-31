package logger

import "go.uber.org/zap"

var L = zap.NewNop()

func New() *zap.Logger {
	L, _ = zap.NewProduction()
	return L
}
