package pkg

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Controller struct {
	mu                   sync.Mutex
	stater               stater
	failureCount         int
	lastFailure          time.Time
	halfOpenSuccessCount int
	failureThreshold     int
	recoveryTime         time.Duration
	halfOpenMaxRequest   int
	timeout              time.Duration
}

func NewController(
	failureThreshold int,
	recoveryTime time.Duration,
	halfOpenMaxRequest int,
	timeout time.Duration,
) *Controller {
	return &Controller{
		stater:             Close,
		failureThreshold:   failureThreshold,
		recoveryTime:       recoveryTime,
		halfOpenMaxRequest: halfOpenMaxRequest,
		timeout:            timeout,
	}
}

func (c *Controller) Call(fn func() (any, error)) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.stater {
	case Close:
		return c.handleCloseState(fn)
	case Open:
		return c.handleOpenState()
	case HalfOpen:
		return c.handleHalfOpenState(fn)
	default:
		return nil, errors.New("unknown controller stater")
	}
}

func (c *Controller) handleCloseState(fn func() (any, error)) (any, error) {
	result, err := c.runWithTimeout(fn)
	if err != nil {
		c.failureCount++
		c.lastFailure = time.Now()

		if c.failureCount >= c.failureThreshold {
			c.stater = Open
		}

		return nil, err
	}

	c.resetController()
	return result, nil
}

func (c *Controller) handleOpenState() (any, error) {
	if time.Since(c.lastFailure) > c.recoveryTime {
		c.stater = HalfOpen
		c.halfOpenSuccessCount = 0
		c.failureCount = 0
		return nil, nil
	}

	return nil, errors.New("request blocked")
}

func (c *Controller) handleHalfOpenState(fn func() (any, error)) (any, error) {
	result, err := c.runWithTimeout(fn)
	if err != nil {
		c.stater = Open
		c.lastFailure = time.Now()
		return nil, err
	}

	c.halfOpenSuccessCount++

	if c.halfOpenSuccessCount >= c.halfOpenMaxRequest {
		c.resetController()
	}

	return result, nil
}

func (c *Controller) resetController() {
	c.failureCount = 0
	c.stater = Close
}

func (c *Controller) runWithTimeout(fn func() (any, error)) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	resultChan := make(chan struct {
		result any
		err    error
	}, 1)

	go func() {
		result, err := fn()
		resultChan <- struct {
			result any
			err    error
		}{result: result, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, errors.New("requested timed out")
	case result := <-resultChan:
		return result.result, result.err
	}
}
