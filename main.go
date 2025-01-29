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
	if len(os.Args) > 1 {
		func() {
			// CREATE DATABASE spindle;
			// CREATE TABLE locktable (
			// 	name TEXT PRIMARY KEY,
			// 	heartbeat TIMESTAMP,
			// 	token TIMESTAMP,
			// 	writer TEXT
			// );

			pgctx := context.Background()
			conn, err := pgx.Connect(pgctx, os.Args[1])
			if err != nil {
				log.Println("Connect failed:", err)
				return
			}

			defer conn.Close(pgctx)
			err = conn.Ping(pgctx)
			if err != nil {
				log.Println("Ping failed:", err)
			} else {
				log.Println("PING!")
			}

			tag, err := conn.Exec(pgctx, "insert into locktable (name, heartbeat, token, writer) values ('spindle_name', '2021-10-10 01:01:01', '2021-10-10 01:01:01', 'writer_me');")
			if err != nil {
				log.Println("Exec failed:", err)
			} else {
				log.Println(tag.String())
			}
		}()
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
