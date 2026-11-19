package handler

import (
	"errors"
	"sync/atomic"
	"time"
)

type BackpressureStrategy int8

const (
	BackpressureStrategyUnset BackpressureStrategy = iota
	BackpressureStrategyBlock
	BackpressureStrategyDrop
	BackpressureStrategyTimeout
	BackpressureStrategySample
)

var (
	ErrBackpressureDropped = errors.New("backpressure dropped message")
	ErrBackpressureTimeout = errors.New("backpressure enqueue timeout")
)

type BackpressureConfig struct {
	Strategy   BackpressureStrategy
	Timeout    time.Duration
	SampleRate uint64
}

func NewBackpressureConfig(strategy BackpressureStrategy) BackpressureConfig {
	return BackpressureConfig{Strategy: strategy}
}

func (c BackpressureConfig) WithTimeout(timeout time.Duration) BackpressureConfig {
	c.Timeout = timeout
	return c
}

func (c BackpressureConfig) WithSampleRate(rate uint64) BackpressureConfig {
	c.SampleRate = rate
	return c
}

func (c BackpressureConfig) Normalize(defaultStrategy BackpressureStrategy) BackpressureConfig {
	if c.Strategy == BackpressureStrategyUnset {
		c.Strategy = defaultStrategy
	}
	if c.Strategy == BackpressureStrategyTimeout && c.Timeout <= 0 {
		c.Timeout = 100 * time.Millisecond
	}
	if c.Strategy == BackpressureStrategySample && c.SampleRate == 0 {
		c.SampleRate = 10
	}
	return c
}

type BackpressureStats struct {
	Enqueued uint64
	Dropped  uint64
	TimedOut uint64
	Sampled  uint64
}

type BackpressureCounter struct {
	enqueued  uint64
	dropped   uint64
	timedOut  uint64
	sampled   uint64
	sampleSeq uint64
}

func (c *BackpressureCounter) AddEnqueued() {
	atomic.AddUint64(&c.enqueued, 1)
}

func (c *BackpressureCounter) AddDropped() {
	atomic.AddUint64(&c.dropped, 1)
}

func (c *BackpressureCounter) AddTimedOut() {
	atomic.AddUint64(&c.timedOut, 1)
}

func (c *BackpressureCounter) AddSampled() {
	atomic.AddUint64(&c.sampled, 1)
}

// ShouldSample reports whether this overloaded entry should be retained.
// Call AddSampled only after the entry has actually been enqueued.
func (c *BackpressureCounter) ShouldSample(rate uint64) bool {
	if rate <= 1 {
		return true
	}
	n := atomic.AddUint64(&c.sampleSeq, 1)
	return n%rate == 0
}

// AllowSample preserves the previous convenience API for callers that only
// need to make a sampling decision.
func (c *BackpressureCounter) AllowSample(rate uint64) bool {
	if !c.ShouldSample(rate) {
		return false
	}
	c.AddSampled()
	return true
}

func (c *BackpressureCounter) Snapshot() BackpressureStats {
	return BackpressureStats{
		Enqueued: atomic.LoadUint64(&c.enqueued),
		Dropped:  atomic.LoadUint64(&c.dropped),
		TimedOut: atomic.LoadUint64(&c.timedOut),
		Sampled:  atomic.LoadUint64(&c.sampled),
	}
}

type BackpressureStatsProvider interface {
	BackpressureStats() BackpressureStats
}
