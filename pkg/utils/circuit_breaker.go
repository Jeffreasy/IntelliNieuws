package utils

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitOpen                         // Too many failures, blocking calls
	CircuitHalfOpen                     // Testing if service recovered
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name        string
	failures    int
	successes   int
	lastFailure time.Time
	lastSuccess time.Time
	state       CircuitState
	threshold   int           // Number of failures before opening
	timeout     time.Duration // How long to stay open
	mu          sync.Mutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:      name,
		threshold: threshold,
		timeout:   timeout,
		state:     CircuitClosed,
	}
}

// Call executes the given function through the circuit breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Check if circuit is open
	if cb.state == CircuitOpen {
		// Check if timeout has elapsed
		if time.Since(cb.lastFailure) < cb.timeout {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker '%s' is open (failures: %d, timeout: %v remaining)",
				cb.name, cb.failures, cb.timeout-time.Since(cb.lastFailure))
		}
		// Try to close circuit - move to half-open
		cb.state = CircuitHalfOpen
		cb.successes = 0
	}

	state := cb.state
	cb.mu.Unlock()

	// Execute function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// Failure
		cb.failures++
		cb.lastFailure = time.Now()

		if state == CircuitHalfOpen {
			// Failed during recovery - open circuit again
			cb.state = CircuitOpen
		} else if cb.failures >= cb.threshold {
			// Too many failures - open circuit
			cb.state = CircuitOpen
		}

		return err
	}

	// Success
	cb.successes++
	cb.lastSuccess = time.Now()

	if state == CircuitHalfOpen {
		// Successful recovery - close circuit
		cb.state = CircuitClosed
		cb.failures = 0
	} else if cb.state == CircuitClosed && cb.failures > 0 {
		// Partial recovery
		cb.failures--
	}

	return nil
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return map[string]interface{}{
		"name":         cb.name,
		"state":        cb.state.String(),
		"failures":     cb.failures,
		"successes":    cb.successes,
		"last_failure": cb.lastFailure,
		"last_success": cb.lastSuccess,
		"threshold":    cb.threshold,
		"timeout":      cb.timeout.String(),
	}
}

// Reset manually resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.successes = 0
	cb.state = CircuitClosed
}

// IsOpen returns whether the circuit is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state == CircuitOpen
}

// IsClosed returns whether the circuit is closed
func (cb *CircuitBreaker) IsClosed() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state == CircuitClosed
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
}

// NewCircuitBreakerManager creates a new manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetOrCreate returns an existing circuit breaker or creates a new one
func (m *CircuitBreakerManager) GetOrCreate(name string, threshold int, timeout time.Duration) *CircuitBreaker {
	m.mu.RLock()
	cb, exists := m.breakers[name]
	m.mu.RUnlock()

	if exists {
		return cb
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists := m.breakers[name]; exists {
		return cb
	}

	cb = NewCircuitBreaker(name, threshold, timeout)
	m.breakers[name] = cb
	return cb
}

// Get returns a circuit breaker by name
func (m *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cb, exists := m.breakers[name]
	return cb, exists
}

// GetAllStats returns stats for all circuit breakers
func (m *CircuitBreakerManager) GetAllStats() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]map[string]interface{}, 0, len(m.breakers))
	for _, cb := range m.breakers {
		stats = append(stats, cb.GetStats())
	}
	return stats
}

// ResetAll resets all circuit breakers
func (m *CircuitBreakerManager) ResetAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cb := range m.breakers {
		cb.Reset()
	}
}
