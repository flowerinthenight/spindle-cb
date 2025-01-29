package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flowerinthenight/clockbound-ffi-go"
	_ "github.com/jackc/pgx/v5/stdlib"
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

			db, err := sql.Open("pgx", os.Args[1])
			if err != nil {
				log.Println("Connect failed:", err)
				return
			}

			defer db.Close()
			err = db.Ping()
			if err != nil {
				log.Println("Ping failed:", err)
			} else {
				log.Println("PING!")
			}

			var q strings.Builder

			// fmt.Fprintf(&q, "insert into locktable (name, heartbeat, token, writer) ")
			// fmt.Fprintf(&q, "values ('spindle', $1, $2, 'writer_me');")
			// tag, err := conn.Exec(pgctx, q.String(), time.Now(), time.Now())
			// if err != nil {
			// 	log.Println("Exec failed:", err)
			// } else {
			// 	log.Println(tag.String())
			// }

			fmt.Fprintf(&q, "select * from locktable where name = 'spindle';")
			var name, writer string
			var hb, token time.Time
			err = db.QueryRow(q.String()).Scan(&name, &hb, &token, &writer)
			if err != nil {
				log.Println("QueryRow failed:", err)
			} else {
				log.Println(name, hb, token, writer)
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
