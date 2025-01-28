package clockbound

import (
	"log"
	"sync/atomic"
	"time"
)

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -lclockbound
// #include "cb_ffi.h"
import "C"

type cbtime struct {
	earliest_s  int
	earliest_ns int
	latest_s    int
	latest_ns   int
	status      int
}

type NowT struct {
	Earliest time.Time
	Latest   time.Time
	Status   int
}

type ClockBound struct {
	active atomic.Int32
	get    chan struct{}
	data   chan cbtime
	close  chan error
	done   chan error
}

func (cb *ClockBound) Now() NowT {
	cb.get <- struct{}{}
	d := <-cb.data
	return NowT{
		Earliest: time.Unix(int64(d.earliest_s), int64(d.earliest_ns)),
		Latest:   time.Unix(int64(d.latest_s), int64(d.latest_ns)),
		Status:   d.status,
	}
}

func (cb *ClockBound) Close() {
	if cb.active.Load() == 0 {
		return
	}

	cb.close <- nil
	<-cb.done
}

func New() *ClockBound {
	cb := &ClockBound{}
	cb.get = make(chan struct{}, 1)
	cb.data = make(chan cbtime, 1)
	cb.close = make(chan error, 1)
	cb.done = make(chan error, 1)

	go func() {
		e_open := C.cb_open()
		log.Println("e_open:", e_open)
		cb.active.Store(1)

		for {
			select {
			case <-cb.close:
				C.cb_close()
				cb.active.Store(0)
				cb.done <- nil
				return
			case <-cb.get:
			}

			var earliest_s, latest_s, status C.int
			var earliest_ns, latest_ns C.int

			_ = C.cb_now(
				&earliest_s,
				&earliest_ns,
				&latest_s,
				&latest_ns,
				&status,
			)

			cb.data <- cbtime{
				earliest_s:  int(earliest_s),
				earliest_ns: int(earliest_ns),
				latest_s:    int(latest_s),
				latest_ns:   int(latest_ns),
				status:      int(status),
			}
		}
	}()

	return cb
}
