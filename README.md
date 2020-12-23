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

### Get Snapshot
```go
// get a snapshot of the histogram (no reset)
snapshot := hist.Snapshot(false)

// create an hdrhistogram.Histogram from a snapshot
histCopy := hist.Snapshot(false).ToHistogram()
```


### Go Routine to ship Percentiles
```go
func shipPercentiles(server string,hist *safehdrhistorgram,done <-chan bool) {
    for {
        select {
            case <-done:
                return
            case <-time.After(time.Duration(60) * time.Second):
                // send to server (code not shown)
                sendPercentilesToServer(server,hist.Percentiles(false))
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
	autoReset bool,
	done <-chan bool) {
    for {
        select {
            case <-done:
                return
            case <-time.After(interval):
            	// send to server (code not shown)
                sendPercentilesToServer(server,hist.Percentiles(autoReset))
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
                sendSnapshotToServer(server,hist.Snapshot(false))
        }
    }
}
```

### Reset a Histogram
```go
// reset the state of a histogram
hist.Reset()

// reset the state of a histogram after a snapshot
snapshot := hist.Snapshot(true)

// reset the state of a histogram after a Percentiles snapshot
snapshot := hist.Percentiles(true)
```

## HistogramMap
A HistogramMap manages a collection of histograms that are referenced by name. It allows for dynamic creation of
histograms, based on usage, and allows a large number of histograms to be managed by a single command channel that is
processed by a single go routine.

The functionality of HistogramMap is more or less the same as Histogram but for operations on individual histograms
the name of the histogram must be provided.

```go
// recording to a histogram directly
hist.RecordValue(latency)

// recording to a histograms via a HistogramMap
hist.RecordValue(latency, "get-user")

// we can also record a single value to multiple histograms with one call
hist.RecordValue(latency, "get-user","user-api","api")
```

Because HistogramMap is a collection of histograms, bulk operations (such as the record example shown above) are
supported.

* Record
* RequestSnapshot
* SnapshotAll
* RequestPercentiles
* PercentilesAll
* RequestReset
* ResetAll

The Request* variants allow multiple names to be specified and are non-blocking. The *All variants operate on all
histogram instances in the collection and wait for the operation to complete before returning.

In addition, HistogramMap provides the Names() function to return the names of all the histograms being managed

## Examples

## About HdrHistogram
* [HdrHistogram](https://github.com/HdrHistogram/HdrHistogram): a good summary of HdrHistogram packages
* [How NOT to Measure Latency](https://www.youtube.com/watch?v=lJ8ydIuPFeU&feature=youtu.be): one of the many 
  presentations by the creator of HdrHistogram, Gil Tene
* [Everything you know about latency is Wrong](https://bravenewgeek.com/everything-you-know-about-latency-is-wrong/):
  a great blog post by Tyler Treat, which does an excellent job of summarizing a Gil Tene presentation 


