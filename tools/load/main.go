// A tiny load generator for demos (autoscaling, rolling updates). Standard library only.
//
//	go run ./tools/load/main.go                          # 20 workers, 60s, http://localhost:8081/api/message
//	go run ./tools/load/main.go -url http://<ip>:8081/api/message -c 40 -d 2m
//
// Prints one line per second; stop early with Ctrl+C.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8081/api/message", "URL to call")
	workers := flag.Int("c", 20, "concurrent workers")
	duration := flag.Duration("d", 60*time.Second, "how long to run")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, *duration)
	defer cancel()

	// A new connection per request: a Kubernetes Service picks a pod per connection, so reused
	// connections would keep hitting the same pods and new pods would get no traffic.
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}

	var ok, failed atomic.Int64
	var wg sync.WaitGroup
	for range *workers {
		wg.Go(func() {
			for ctx.Err() == nil {
				resp, err := client.Get(*url)
				if err != nil || resp.StatusCode != http.StatusOK {
					if ctx.Err() == nil { // don't count requests cut off by the end of the run
						failed.Add(1)
					}
				} else {
					ok.Add(1)
				}
				if resp != nil {
					resp.Body.Close()
				}
			}
		})
	}

	fmt.Printf("%d workers → %s for %s\n", *workers, *url, *duration)
	start := time.Now()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	var lastOK, lastFailed int64
	for done := false; !done; {
		select {
		case <-ctx.Done():
			done = true
		case <-tick.C:
			o, f := ok.Load(), failed.Load()
			fmt.Printf("%5.0fs  %5d req/s  %3d errors/s\n", time.Since(start).Seconds(), o-lastOK, f-lastFailed)
			lastOK, lastFailed = o, f
		}
	}
	wg.Wait()
	fmt.Printf("done: %d ok, %d errors\n", ok.Load(), failed.Load())
}
