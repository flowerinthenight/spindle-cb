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

			log.Printf("len: %d\n", len(m))
			log.Printf("%X\n", m)

			size := binary.LittleEndian.Uint32(m[8:12])
			log.Printf("size: 0x%X %d\n", size, size)

			ver := binary.LittleEndian.Uint16(m[12:14])
			log.Printf("version: 0x%X %d\n", ver, ver)

			gen := binary.LittleEndian.Uint16(m[14:16])
			log.Printf("generation: 0x%X %d\n", gen, gen)

			asof_s := binary.LittleEndian.Uint64(m[16:24])
			log.Printf("as-of-ts (s): 0x%X %d\n", asof_s, asof_s)
			asof_ns := binary.LittleEndian.Uint64(m[24:32])
			log.Printf("as-of-ts (ns): 0x%X %d\n", asof_ns, asof_ns)
			ts := time.Unix(int64(asof_s), int64(asof_ns))

			va_s := binary.LittleEndian.Uint64(m[32:40])
			log.Printf("void-after-ts (s): 0x%X %d\n", va_s, va_s)
			va_ns := binary.LittleEndian.Uint64(m[40:48])
			log.Printf("void-after-ts (ns): 0x%X %d\n", va_ns, va_ns)
			vts := time.Unix(int64(va_s), int64(va_ns))

			log.Printf("ts1: %v %v\n", ts.Format(time.RFC3339), ts.Format(time.RFC3339Nano))
			log.Printf("ts2: %v %v\n", vts.Format(time.RFC3339), vts.Format(time.RFC3339Nano))

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

func fromUnixNano(nano uint64) time.Time {
	return time.Unix(int64(nano/1e9), int64(nano%1e9))
}
