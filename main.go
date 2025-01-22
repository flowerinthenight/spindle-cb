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

			defer func(s time.Time) {
				log.Println("clockbound took", time.Since(s))
			}(time.Now())

			log.Printf("len: %d\n", len(m))
			log.Printf("%X\n", m)

			size := binary.LittleEndian.Uint32(m[8:12])
			log.Printf("size: 0x%X %d\n", size, size)

			ver := binary.LittleEndian.Uint16(m[12:14])
			log.Printf("version: 0x%X %d\n", ver, ver)

			gen := binary.LittleEndian.Uint16(m[14:16])
			log.Printf("generation: 0x%X %d\n", gen, gen)

			// As-of-timestamp
			asof_s := binary.LittleEndian.Uint64(m[16:24])
			log.Printf("as-of-ts (s): 0x%X %d\n", asof_s, asof_s)
			asof_ns := binary.LittleEndian.Uint64(m[24:32])
			log.Printf("as-of-ts (ns): 0x%X %d\n", asof_ns, asof_ns)
			ts := time.Unix(int64(asof_s), int64(asof_ns))

			// Void-after-timestamp
			va_s := binary.LittleEndian.Uint64(m[32:40])
			log.Printf("void-after-ts (s): 0x%X %d\n", va_s, va_s)
			va_ns := binary.LittleEndian.Uint64(m[40:48])
			log.Printf("void-after-ts (ns): 0x%X %d\n", va_ns, va_ns)
			vts := time.Unix(int64(va_s), int64(va_ns))

			log.Printf("as_of_ts  : %v\n", ts.Format(time.RFC3339Nano))
			log.Printf("void_after: %v\n", vts.Format(time.RFC3339Nano))

			bound := binary.LittleEndian.Uint64(m[48:56])
			log.Printf("bound_ns: 0x%X %d\n", bound, bound)

			drift := binary.LittleEndian.Uint32(m[56:60])
			log.Printf("drift: 0x%X %d\n", drift, drift)

			reserved := binary.LittleEndian.Uint32(m[60:64])
			log.Printf("reserved: 0x%X %d\n", reserved, reserved)

			status := binary.LittleEndian.Uint32(m[64:68])
			log.Printf("clock_status: 0x%X %d\n", status, status)

			earliest := ts.Add(-1 * (time.Nanosecond * time.Duration(bound)))
			latest := ts.Add(time.Nanosecond * time.Duration(bound))

			unix_ns := latest.UnixNano() - (latest.UnixNano()-earliest.UnixNano())/2
			log.Printf("up : %v\n", latest)
			log.Printf("now: %v\n", fromUnixNano(uint64(unix_ns)))
			log.Printf("low: %v\n", earliest)
			log.Printf("range: %v\n", latest.Sub(earliest))

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
