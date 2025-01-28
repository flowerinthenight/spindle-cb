package clockbound

import (
	"fmt"
	"sync/atomic"
	"time"
)

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -lclockbound
// #include "cb_ffi.h"
import "C"

type ClockStatus int

const (
	ClockStatusUnknown ClockStatus = iota
	ClockStatusSynchronized
	ClockStatusFreeRunning
)

var ClockStatusName = map[ClockStatus]string{
	ClockStatusUnknown:      "UNKNOWN",
	ClockStatusSynchronized: "SYNCHRONIZED",
	ClockStatusFreeRunning:  "FREE_RUNNING",
}

func (cs ClockStatus) String() string { return ClockStatusName[cs] }

type ClockBoundErrorKind int

const (
	ClockBoundErrorKindNone ClockBoundErrorKind = iota
	ClockBoundErrorKindSyscall
	ClockBoundErrorKindSegmentNotInitialized
	ClockBoundErrorKindSegmentMalformed
	ClockBoundErrorKindCausalityBreach
)

var ClockBoundErrorKindName = map[ClockBoundErrorKind]string{
	ClockBoundErrorKindNone:                  "NONE",
	ClockBoundErrorKindSyscall:               "ERR_SYSCALL",
	ClockBoundErrorKindSegmentNotInitialized: "ERR_SEGMENT_NOT_INITIALIZED",
	ClockBoundErrorKindSegmentMalformed:      "ERR_SEGMENT_MALFORMED",
	ClockBoundErrorKindCausalityBreach:       "ERR_CAUSALITY_BREACH",
}

func (e ClockBoundErrorKind) String() string { return ClockBoundErrorKindName[e] }

type cbtime struct {
	earliest_s  int
	earliest_ns int
	latest_s    int
	latest_ns   int
	status      int
}

// Now represents a range of bounded timestamp from ClockBound.
// The "true" time is somewhere within the range.
type Now struct {
	Earliest time.Time
	Latest   time.Time
	Status   ClockStatus
}

// Client represents a connection to ClockBound's shared memory file via FFI.
type Client struct {
	active atomic.Int32
	error  atomic.Int32
	get    chan struct{}
	data   chan cbtime
	close  chan error
	done   chan error
}

// Now gets a set range of bounded timestamps from ClockBound.
func (c *Client) Now() (Now, error) {
	code := c.error.Load()
	if code != 0 {
		err := fmt.Errorf("Now failed: %d", code)
		i := ClockBoundErrorKind(code)
		if _, ok := ClockBoundErrorKindName[i]; ok {
			err = fmt.Errorf("%v", ClockBoundErrorKindName[i])
		}

		return Now{}, err
	}

	c.get <- struct{}{}
	d := <-c.data
	return Now{
		Earliest: time.Unix(int64(d.earliest_s), int64(d.earliest_ns)),
		Latest:   time.Unix(int64(d.latest_s), int64(d.latest_ns)),
		Status:   ClockStatus(d.status),
	}, nil
}

// Close releases client resources.
func (c *Client) Close() {
	if c.active.Load() == 0 {
		return
	}

	if c.error.Load() != 0 {
		return
	}

	c.close <- nil
	<-c.done
}

// New creates an instance of Client.
func New() *Client {
	c := &Client{}
	c.get = make(chan struct{}, 1)
	c.data = make(chan cbtime, 1)
	c.close = make(chan error, 1)
	c.done = make(chan error, 1)

	go func() {
		c.error.Store(int32(C.cb_open()))
		if c.error.Load() != 0 {
			return
		}

		c.active.Store(1)

		for {
			select {
			case <-c.close:
				c.error.Store(int32(C.cb_close()))
				c.active.Store(0)
				c.done <- nil
				return
			case <-c.get:
			}

			var earliest_s, latest_s, status C.int
			var earliest_ns, latest_ns C.int

			c.error.Store(int32(C.cb_now(
				&earliest_s,
				&earliest_ns,
				&latest_s,
				&latest_ns,
				&status,
			)))

			if c.error.Load() != 0 {
				continue
			}

			c.data <- cbtime{
				earliest_s:  int(earliest_s),
				earliest_ns: int(earliest_ns),
				latest_s:    int(latest_s),
				latest_ns:   int(latest_ns),
				status:      int(status),
			}
		}
	}()

	return c
}
