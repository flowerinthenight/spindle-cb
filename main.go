package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	clockboundclient "github.com/flowerinthenight/clockbound-client-go"
)

func main() {
	client, err := clockboundclient.New()
	if err != nil {
		log.Println("New failed:", err)
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

			now, err := client.Now()
			if err != nil {
				log.Println("Now failed:", err)
				continue
			}

			log.Printf("earliest: %v\n", now.Earliest.Format(time.RFC3339Nano))
			log.Printf("latest  : %v\n", now.Latest.Format(time.RFC3339Nano))
			log.Printf("range: %v\n", now.Latest.Sub(now.Earliest))
			log.Printf("status: %v\n", now.Status)
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
	client.Close()
}
