package rss

import (
	"math/rand"
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	mu          sync.Mutex
	state       CircuitState
	failures    int
	maxFailures int
	openSince   time.Time
	timeout     time.Duration
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.openSince) > cb.timeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = StateClosed
}

func (cb *CircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
		cb.openSince = time.Now()
	}
}

func (cb *CircuitBreaker) BackoffDuration() time.Duration {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Exponential: 30s, 1m, 5m, 30m
	stages := []time.Duration{
		30 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
		30 * time.Minute,
	}
	idx := cb.failures - cb.maxFailures
	if idx < 0 {
		idx = 0
	}
	if idx >= len(stages) {
		idx = len(stages) - 1
	}
	d := stages[idx]
	// Jitter ±20%
	jitter := time.Duration(float64(d) * 0.2 * (rand.Float64()*2 - 1))
	return d + jitter
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) Failures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failures
}
