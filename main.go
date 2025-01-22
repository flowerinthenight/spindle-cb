package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shogo82148/go-clockboundc"
)

func main() {
	c, err := clockboundc.NewWithPath(clockboundc.DefaultSocketPath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	ticker := time.NewTicker(time.Second * 3)

	go func() {
		for {
			select {
			case <-ctx.Done():
				done <- nil
				return
			case <-ticker.C:
			}

			now, err := c.Now()
			if err != nil {
				log.Println(err)
				continue
			}

			if now.Header.Unsynchronized {
				log.Println("Unsynchronized")
			} else {
				log.Println("Synchronized")
			}

			log.Println("Current: ", now.Time)
			log.Println("Earliest:", now.Bound.Earliest)
			log.Println("Latest:  ", now.Bound.Latest)
			log.Println("Range:   ", now.Bound.Latest.Sub(now.Bound.Earliest))
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
	c.Close()
}
