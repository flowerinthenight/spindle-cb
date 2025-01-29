package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	clockbound "github.com/flowerinthenight/clockbound-ffi-go"
	"github.com/jackc/pgx/v5"
)

func main() {
	log.Println("args:", os.Args)

	if false {
		pgctx := context.Background()
		conn, err := pgx.Connect(pgctx, os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Println("Connect failed:", err)
			return
		}

		err = conn.Ping(pgctx)
		if err != nil {
			log.Println("Ping failed:", err)
		}

		defer conn.Close(pgctx)
	}

	client := clockbound.New()
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
