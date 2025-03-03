package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/flowerinthenight/spindle-cb"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbstr := flag.String("db", "", "database connection string")
	table := flag.String("table", "locktable", "table name")
	name := flag.String("name", "mylock", "lock name")
	flag.Parse()

	// To run, update the database name, table name, and, optionally, the lock name.
	db, err := sql.Open("pgx", *dbstr)
	if err != nil {
		log.Println("Open failed:", err)
		return
	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Println("Ping failed:", err)
	} else {
		log.Println("PING!")
	}

	quit, cancel := context.WithCancel(context.Background())
	lock := spindle.New(db,
		*table,
		*name,
		spindle.WithDuration(5000),
		spindle.WithLeaderCallback(nil, func(d any, m []byte) {
			log.Println("callback:", string(m))
		}),
	)

	done := make(chan error, 1)
	lock.Run(quit, done) // start main loop

	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		<-sigch
		cancel()
	}()

	<-done
}
