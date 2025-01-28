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

type NowT struct {
	Earliest time.Time
	Latest   time.Time
	Status   int
}

type ClockBound struct {
	active atomic.Int32
	get    chan struct{}
	data   chan int
	close  chan error
}

func (cb *ClockBound) Now() (NowT, error) {
	cb.get <- struct{}{}
	data := <-cb.data
	log.Println("now:", data)
	return NowT{}, nil
}

func (cb *ClockBound) Close() {
	if cb.active.Load() == 0 {
		return
	}

	cb.close <- nil
}

func New() *ClockBound {
	cb := &ClockBound{}
	cb.get = make(chan struct{}, 1)
	cb.data = make(chan int, 1)
	cb.close = make(chan error, 1)

	go func() {
		_ = C.cb_open()
		cb.active.Store(1)
		log.Println("starting cb loop...")

		for {
			select {
			case <-cb.close:
				_ = C.cb_close()
				cb.active.Store(0)
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

			log.Println("from C:", earliest_s, earliest_ns, latest_s, latest_ns, status)
			cb.data <- int(earliest_s)
		}
	}()

	return cb
}

// func main() {
// 	var earliest_s, latest_s, status C.int
// 	var earliest_ns, latest_ns C.int
// 	_ = C.cb_open()

// 	_ = C.cb_now(
// 		&earliest_s,
// 		&earliest_ns,
// 		&latest_s,
// 		&latest_ns,
// 		&status,
// 	)

// 	log.Println("from C:", earliest_s, earliest_ns, latest_s, latest_ns, status)

// 	_ = C.cb_close()

// 	client, err := clockboundclient.New()
// 	if err != nil {
// 		log.Println("New failed:", err)
// 		return
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	done := make(chan error, 1)
// 	ticker := time.NewTicker(time.Second * 3)
// 	first := make(chan struct{}, 1)
// 	first <- struct{}{}

// 	go func() {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				done <- nil
// 				return
// 			case <-first:
// 			case <-ticker.C:
// 			}

// 			now, err := client.Now()
// 			if err != nil {
// 				log.Println("Now failed:", err)
// 				continue
// 			}

// 			log.Printf("earliest: %v\n", now.Earliest.Format(time.RFC3339Nano))
// 			log.Printf("latest  : %v\n", now.Latest.Format(time.RFC3339Nano))
// 			log.Printf("range: %v\n", now.Latest.Sub(now.Earliest))
// 			log.Printf("status: %v\n", now.Status)
// 			log.Println("")
// 		}
// 	}()

// 	// Interrupt handler.
// 	go func() {
// 		sigch := make(chan os.Signal, 1)
// 		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
// 		log.Println("signal:", <-sigch)
// 		cancel()
// 	}()

// 	<-done

// 	ticker.Stop()
// 	client.Close()
// }
