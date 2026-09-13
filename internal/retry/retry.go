package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Strategy определяет способ вычисления задержки между повторными попытками.
type Strategy int

const (
	// Fixed - повтор с фиксированной задержкой, предсказуем при кратковременных сбоях.
	Fixed Strategy = iota
	// Linear - линейное увеличение задержки: base + step*attempt.
	Linear
	// Exponential - экспоненциальное увеличение задержки: base * multiplier^attempt.
	Exponential
	// Random - случайная задержка в диапазоне [base, base+step], предотвращает thundering herd.
	Random
)

// StrategyConfig параметры стратегии retry.
type StrategyConfig struct {
	Base        time.Duration // базовая задержка
	MaxAttempts int           // количество дополнительных попыток
	Step        time.Duration // шаг для Linear и Random
	Multiplier  int           // множитель для Exponential (0 = используется 2)
}

// Classification описывает результат классификации ошибки.
type Classification struct {
	Retriable bool     // true — операцию можно повторить
	Strategy  Strategy // стратегия вычисления задержки для данного типа ошибки
}

// Config конфигурация retry-механизма.
type Config struct {
	StrategyCfg StrategyConfig
	Classify    func(err error) Classification
}

func (c *Config) delay(attempt int, strategy Strategy) time.Duration {
	base := c.StrategyCfg.Base

	switch strategy {
	case Linear:
		return base + c.StrategyCfg.Step*time.Duration(attempt)
	case Exponential:
		multiplier := c.StrategyCfg.Multiplier
		if multiplier <= 0 {
			multiplier = 2
		}
		return time.Duration(float64(base) * math.Pow(float64(multiplier), float64(attempt)))
	case Random:
		step := c.StrategyCfg.Step
		if step <= 0 {
			step = base
		}
		return base + time.Duration(rand.Int63n(int64(step)))
	default:
		return base
	}
}

// DefaultConfig - конфигурация по умолчанию: 3 попытки, Linear 1s, 3s, 5s.
var DefaultConfig = Config{
	StrategyCfg: StrategyConfig{
		Base:        1 * time.Second,
		MaxAttempts: 3,
		Step:        2 * time.Second,
		Multiplier:  2,
	},
}

// Execute выполняет fn с повторными попытками согласно cfg.
// Количество попыток = MaxAttempts + 1.
func Execute(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.StrategyCfg.MaxAttempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err

			classification := Classification{Retriable: true, Strategy: Fixed}
			if cfg.Classify != nil {
				classification = cfg.Classify(err)
			}

			if !classification.Retriable {
				return err
			}

			if attempt < cfg.StrategyCfg.MaxAttempts {
				delay := cfg.delay(attempt, classification.Strategy)
				select {
				case <-ctx.Done():
					return fmt.Errorf("operation cancelled after %d attempts: %w", attempt+1, ctx.Err())
				case <-time.After(delay):
				}
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", cfg.StrategyCfg.MaxAttempts+1, lastErr)
}
