package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gotomgo/safehdrhistogram"
)

func dumpPercentiles(hist *safehdrhistogram.Histogram, done <-chan bool) {
	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second):
			var buf bytes.Buffer
			hist.Percentiles(false).Write(&buf)
			fmt.Println(buf.String())
		}
	}
}

func main() {
	// Create a concurrency safe version of HdrHistogram using the standard parameters
	hist := safehdrhistogram.NewHistogram(1, 30000000, 3)
	// close the histogram when the program exits
	defer hist.Close()

	// create a WaitGroup to block the main thread until the goroutines complete
	wg := sync.WaitGroup{}

	done := make(chan bool)
	go dumpPercentiles(hist, done)

	// run 5 goroutines that run for an avg of 5 seconds each and record
	// their random wait time over 100 iterations
	for i := 0; i < 5; i++ {
		wg.Add(1)

		// simulate concurrent activity by recording 100 random latency values
		go func(hist *safehdrhistogram.Histogram) {
			for i := 0; i < 100; i++ {
				// calculate a wait time
				waitTime := time.Duration(rand.Float64()*100) * time.Millisecond
				// wait
				<-time.After(waitTime)
				// record wait time
				hist.Record(waitTime.Microseconds())
			}

			wg.Done()
		}(hist)
	}

	// wait for all the go routines to complete
	wg.Wait()
	// stop the dumpPercentiles coroutine
	done <- true
}
