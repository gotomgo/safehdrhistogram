# Concurrency Safe HDR Historgram

safehdrhistogram is a GO (Golang) package that wraps https://github.com/HdrHistogram/histogram-go with safe, 
concurrent read and write access to a hdrhistogram.Histogram, or a collection of named instances of hdrhistogram.Histogram
(see HistogramMap)

## Installation

```sh
$ go get -u github.com/gotomgo/safhdrhistogram
```

### import

```go
import "github.com/gotomgo/safehdrhistogram"
```

### optional import
When you need or want to access the underlying hdrhistogram package directly:

```go
import "github.com/HdrHistogram/hdrhistogram-go"
```

## Quick Start

### Create Histogram 
```go
// Create a concurrency safe version of HdrHistogram using the standard parameters 
hist := safehdrhistogram.NewHistogram(1, 30000000, 3)
```

### Create a Histogram using a configuration
```go
// Create a concurrency safe version of HdrHistogram using configuration values 
hist := safehdrhistogram.NewHistogramFromConfig(
		safehdrhistogram.HistogramConfig{
			LowestDiscernibleValue:         lowestDiscernibleValue,
			HighestTrackableValue:          highestTrackableValue,
			NumberOfSignificantValueDigits: numberOfSignificantValueDigits,
			CommandBufferSize:              32,
		})
```

### Record Value
```go
startTime := time.Now()
// ... do some work here
hist.RecordValue(time.Since(startTime).Microseconds())
```

### Go Routine to ship Percentiles
```go
func shipPercentiles(server string,hist *safehdrhistorgram,done <-chan bool) {
    for {
        select {
            case <-done:
                return
            case <-time.After(time.Duration(60) * time.Second):
                sendPercentiles(server,hist.Percentiles(false))
        }
    }
}
```

### Go Routine to ship Percentiles (fully parameterized)
```go
func shipPercentiles(
	server string,
	hist *safehdrhistorgram,
	interval time.Duration,
	resetHist bool,
	done <-chan bool) {
    for {
        select {
            case <-done:
                return
            case <-time.After(interval):
            	// send to server (code not shown)
                sendPercentiles(server,hist.Percentiles(resetHist))
        }
    }
}
```

### Go Routine to ship Snapshots
```go
func shipSnapshot(server string,hist *safehdrhistorgram,done <-chan bool) {
    for {
        select {
            case <-done:
                return
            case <-time.After(time.Duration(60) * time.Second):
                // send to server (code not shown)
                sendSnapshot(server,hist.Snapshot(false))
        }
    }
}
```

## Examples

## About HdrHistogram
A good summary of HdrHistogram can be found here: https://github.com/HdrHistogram/HdrHistogram

One of the many presentations by the creator of HdrHistogram, Gil Tene, called 'How NOT to Measure Latency'
can be found here: https://www.youtube.com/watch?v=lJ8ydIuPFeU&feature=youtu.be

A great blog post by Tyler Treat, 'Everything you know about latency is Wrong' which does an excellent job of 
summarizing a Gil Tene presentation, can be found here: 
https://bravenewgeek.com/everything-you-know-about-latency-is-wrong/

