package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flowerinthenight/spindle-cb/clockbound"
)

// // #cgo CFLAGS: -g -Wall
// // #cgo LDFLAGS: -lclockbound
// // #include "hello.h"
// import "C"

func main() {
	// var earliest_s, latest_s, status C.int
	// var earliest_ns, latest_ns C.int
	// _ = C.cb_open()

	// _ = C.cb_now(
	// 	&earliest_s,
	// 	&earliest_ns,
	// 	&latest_s,
	// 	&latest_ns,
	// 	&status,
	// )

	// log.Println("from C:", earliest_s, earliest_ns, latest_s, latest_ns, status)

	// _ = C.cb_close()

	cb := clockbound.New()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	ticker := time.NewTicker(time.Second * 3)
	first := make(chan struct{}, 1)
	first <- struct{}{}

	go func() {
		for {
			select {
			case <-ctx.Done():
				done <- nil
				return
			case <-first:
			case <-ticker.C:
			}

			now, err := cb.Now()
			log.Printf("earliest: %v\n", now.Earliest.Format(time.RFC3339Nano))
			log.Printf("latest  : %v\n", now.Latest.Format(time.RFC3339Nano))
			log.Printf("range: %v\n", now.Latest.Sub(now.Earliest))
			log.Printf("status: %v\n", now.Status)
			log.Printf("err: %v\n", err)
			log.Println("")
		}
	}()

	// Interrupt handler.
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		log.Println("signal:", <-sigch)
		cancel()
	}()

	<-done

	ticker.Stop()
	cb.Close()
}
