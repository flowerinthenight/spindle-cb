package main

import (
	"context"
	"encoding/binary"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mmap "github.com/edsrzf/mmap-go"
)

func main() {
	f, err := os.OpenFile("/var/run/clockbound/shm", os.O_RDONLY, 0755)
	if err != nil {
		log.Println("OpenFile failed:", err)
		return
	}

	defer f.Close()

	m, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		log.Println("Map failed:", err)
		return
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

			magic := binary.BigEndian.Uint64(m[:8])
			log.Printf("magic: %X\n", magic)
			log.Printf("%X\n", m)

			// now, err := c.Now()
			// if err != nil {
			// 	log.Println(err)
			// 	continue
			// }

			// if now.Header.Unsynchronized {
			// 	log.Println("Unsynchronized")
			// } else {
			// 	log.Println("Synchronized")
			// }

			// log.Println("Current: ", now.Time)
			// log.Println("Earliest:", now.Bound.Earliest)
			// log.Println("Latest:  ", now.Bound.Latest)
			// log.Println("Range:   ", now.Bound.Latest.Sub(now.Bound.Earliest))
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

	if err := m.Unmap(); err != nil {
		log.Println("Unmap failed:", err)
		return
	}
}
