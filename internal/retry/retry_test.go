package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecute_Success(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 10 * time.Millisecond, MaxAttempts: 3},
	}
	calls := 0
	err := Execute(context.Background(), cfg, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestExecute_SuccessAfterRetries(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 10 * time.Millisecond, MaxAttempts: 3},
	}
	calls := 0
	err := Execute(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return errors.New("temporary error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestExecute_FailsAfterAllRetries(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 10 * time.Millisecond, MaxAttempts: 2},
	}
	calls := 0
	err := Execute(context.Background(), cfg, func() error {
		calls++
		return errors.New("persistent error")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestExecute_NonRetriableStopsImmediately(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 10 * time.Millisecond, MaxAttempts: 2},
		Classify: func(err error) Classification {
			return Classification{Retriable: false}
		},
	}
	calls := 0
	err := Execute(context.Background(), cfg, func() error {
		calls++
		return errors.New("non-retriable error")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestExecute_ContextCancelled(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 1 * time.Second, MaxAttempts: 2},
	}
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Execute(ctx, cfg, func() error {
		calls++
		return errors.New("temporary error")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled in error chain, got %v", err)
	}
}

func TestExecute_FixedDelay(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 10 * time.Millisecond, MaxAttempts: 2},
		Classify: func(err error) Classification {
			return Classification{Retriable: true, Strategy: Fixed}
		},
	}
	calls := 0
	err := Execute(context.Background(), cfg, func() error {
		calls++
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestExecute_LinearDelay(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 10 * time.Millisecond, MaxAttempts: 2, Step: 5 * time.Millisecond},
		Classify: func(err error) Classification {
			return Classification{Retriable: true, Strategy: Linear}
		},
	}
	calls := 0
	err := Execute(context.Background(), cfg, func() error {
		calls++
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestExecute_ExponentialDelay(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 10 * time.Millisecond, MaxAttempts: 2, Multiplier: 2},
		Classify: func(err error) Classification {
			return Classification{Retriable: true, Strategy: Exponential}
		},
	}
	calls := 0
	err := Execute(context.Background(), cfg, func() error {
		calls++
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestExecute_RandomDelay(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 10 * time.Millisecond, MaxAttempts: 2, Step: 5 * time.Millisecond},
		Classify: func(err error) Classification {
			return Classification{Retriable: true, Strategy: Random}
		},
	}
	calls := 0
	err := Execute(context.Background(), cfg, func() error {
		calls++
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestExecute_NilClassify_DefaultRetriable(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 10 * time.Millisecond, MaxAttempts: 1},
	}
	calls := 0
	err := Execute(context.Background(), cfg, func() error {
		calls++
		return errors.New("error")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestExecute_LinearFormula(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 1 * time.Second, MaxAttempts: 2, Step: 2 * time.Second},
	}
	d0 := cfg.delay(0, Linear)
	d1 := cfg.delay(1, Linear)
	d2 := cfg.delay(2, Linear)

	if d0 != 1*time.Second {
		t.Errorf("expected 1s, got %v", d0)
	}
	if d1 != 3*time.Second {
		t.Errorf("expected 3s, got %v", d1)
	}
	if d2 != 5*time.Second {
		t.Errorf("expected 5s, got %v", d2)
	}
}

func TestExecute_ExponentialFormula(t *testing.T) {
	cfg := Config{
		StrategyCfg: StrategyConfig{Base: 1 * time.Second, MaxAttempts: 2, Multiplier: 2},
	}
	d0 := cfg.delay(0, Exponential)
	d1 := cfg.delay(1, Exponential)
	d2 := cfg.delay(2, Exponential)

	if d0 != 1*time.Second {
		t.Errorf("expected 1s, got %v", d0)
	}
	if d1 != 2*time.Second {
		t.Errorf("expected 2s, got %v", d1)
	}
	if d2 != 4*time.Second {
		t.Errorf("expected 4s, got %v", d2)
	}
}
