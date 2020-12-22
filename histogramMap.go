package safehdrhistogram

import (
	"sync"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// HistogramMap is a collection of named hdrhistogram.Histogram
// instances.
//
//	Notes
//		Each histogram is created on demand when first referenced, and every
//		histogram has the same configuration (see the config field.)
//
type HistogramMap struct {
	config HistogramConfig
	cmds   chan command
	done   chan bool

	// the collection of histograms, protected by a mutex
	lock  sync.RWMutex
	hists map[string]*hdrhistogram.Histogram
}

// NewHistogramMap creates a collection to manage named instances of
// hdrhistogram.Histogram, that are safe for concurrent use
//
//	Notes
//		Named histograms are created dynamically, on first reference.
//
func NewHistogramMap(
	lowestDiscernibleValue,
	highestTrackableValue int64,
	numberOfSignificantValueDigits int) *HistogramMap {

	return NewHistogramMapFromConfig(
		HistogramConfig{
			LowestDiscernibleValue:         lowestDiscernibleValue,
			HighestTrackableValue:          highestTrackableValue,
			NumberOfSignificantValueDigits: numberOfSignificantValueDigits,
			CommandBufferSize:              DefaultCommandBufferSize,
		})
}

// NewHistogramMapFromConfig creates an instance of hdrhistogram.Histogram that is
// safe for concurrent use based on values from a HistogramConfig
//
//	Notes
//		The histogram cannot record values (or anything else) until Start()
//		is called. Eventually, Stop() should be called to terminate processing
//		and close the cmd channel
//
func NewHistogramMapFromConfig(config HistogramConfig) *HistogramMap {
	hdr := &HistogramMap{
		done:   make(chan bool),
		config: config,
		cmds:   make(chan command, config.CommandBufferSize),
		hists:  map[string]*hdrhistogram.Histogram{},
	}

	// start the cmd processor
	process(hdr.cmds, hdr.done)

	return hdr
}

// resolveHistogram looks up a histogram by name, creating and initializing
// it if it doesn't exist
func (hdr *HistogramMap) resolveHistogram(name string) *hdrhistogram.Histogram {
	hdr.lock.Lock()
	defer hdr.lock.Unlock()

	var hist *hdrhistogram.Histogram
	var ok bool

	if hist, ok = hdr.hists[name]; !ok {
		// create a new histogram for this name
		hist = hdrhistogram.New(
			hdr.config.LowestDiscernibleValue,
			hdr.config.HighestTrackableValue,
			hdr.config.NumberOfSignificantValueDigits)

		hist.SetTag(name)

		// remember it
		hdr.hists[name] = hist

		// issue a start command to initialize it
		//
		// Note that unlike the Histogram we do not block until the
		// start command is processed
		hdr.cmds <- command{
			hist:    hist,
			command: cmdStart,
			arg:     nil,
		}
	}

	return hist
}

// RequestRecord requests that a value be recorded but will not block if the
// channel is full
//
//	Notes
//		RequestRecord will not block, so the value is **dropped** if the buffer
//		is full
//
func (hdr *HistogramMap) Record(value int64, names ...string) {
	for _, name := range names {
		// get/create a histogram for name
		hist := hdr.resolveHistogram(name)

		// send the record command without blocking. If the buffer is full, the
		// value is dropped
		select {
		case hdr.cmds <- command{
			hist:    hist,
			command: cmdRecord,
			arg:     value,
		}:
		default:
		}
	}
}

// RequestSnapshot requests a snapshot for one or more histograms and is non-blocking
//
//	Notes
//		Consider buffering for snap if multiple snapshots are requested
//
func (hdr *HistogramMap) RequestSnapshot(snap SnapshotChannel, reset bool, names ...string) {
	for _, name := range names {
		// get/create a histogram for name
		hist := hdr.resolveHistogram(name)

		// request a snapshot
		hdr.cmds <- command{
			hist:    hist,
			command: cmdSnapshot,
			arg:     snap,
		}

		if reset {
			// request a reset
			hdr.cmds <- command{
				hist:    hist,
				command: cmdReset,
			}
		}
	}
}

// Snapshot blocks until a snapshot request completes
func (hdr *HistogramMap) Snapshot(name string, reset bool) *Snapshot {
	// get/create a histogram for name
	hist := hdr.resolveHistogram(name)

	// create a channel for the snapshot
	snap := make(SnapshotChannel)
	defer close(snap)

	// request a snapshot
	hdr.cmds <- command{
		hist:    hist,
		command: cmdSnapshot,
		arg:     snap,
	}

	if reset {
		// request a reset
		hdr.cmds <- command{
			hist:    hist,
			command: cmdReset,
		}
	}

	// block until the snapshot is available, then return it
	return <-snap
}

// SnapshotAll performs a snapshot for every named histogram
//
//	Notes
//		SnapshotAll waits for acknowledgement of the snapshots, and
//		effectively blocks all other activity as the lock for the
//		histogram map is held for the duration
//
func (hdr *HistogramMap) SnapshotAll(snap SnapshotChannel, reset bool) {
	// take the lock as we need to iterate the map of histograms
	hdr.lock.Lock()
	// we may have this lock for a while but nothing in the cmd processing can
	// access it and by the time we release it (and any waiters), the resets
	// have already happened
	defer hdr.lock.Unlock()

	for _, hist := range hdr.hists {
		// send a snapshot command
		hdr.cmds <- command{
			hist:    hist,
			command: cmdSnapshot,
			arg:     snap,
		}

		if reset {
			// request a reset
			hdr.cmds <- command{
				hist:    hist,
				command: cmdReset,
			}
		}
	}

	// use a channel to wait for confirmation of snapshots completing
	done := make(chan bool)

	// request a sync
	hdr.cmds <- command{
		hist:    nil,
		command: cmdSync,
		arg:     done,
	}

	<-done
	close(done)
}

// RequestPercentiles requests a percentiles snapshot for one or more
// histograms and is non-blocking
//
//	Notes
//		Consider buffering for perc if multiple percentiles are requested
//
func (hdr *HistogramMap) RequestPercentiles(perc PercentilesChannel, reset bool, names ...string) {
	for _, name := range names {
		// get/create a histogram for name
		hist := hdr.resolveHistogram(name)

		// request a snapshot
		hdr.cmds <- command{
			hist:    hist,
			command: cmdPercentiles,
			arg:     perc,
		}

		if reset {
			// request a reset
			hdr.cmds <- command{
				hist:    hist,
				command: cmdReset,
			}
		}
	}
}

// Percentiles blocks until a percentiles snapshot request completes
func (hdr *HistogramMap) Percentiles(name string, reset bool) *Percentiles {
	// get/create a histogram for name
	hist := hdr.resolveHistogram(name)

	// create a channel for the snapshot
	perc := make(PercentilesChannel)
	defer close(perc)

	// request a snapshot
	hdr.cmds <- command{
		hist:    hist,
		command: cmdPercentiles,
		arg:     perc,
	}

	if reset {
		// request a reset
		hdr.cmds <- command{
			hist:    hist,
			command: cmdReset,
		}
	}

	// block until the snapshot is available, then return it
	return <-perc
}

// PercentilesAll performs a percentiles snapshot for every named histogram
//
//	Notes
//		PercentilesAll waits for acknowledgement of the percentiles snapshots,
//		and effectively blocks all other activity as the lock for the
//		histogram map is held for the duration
//
func (hdr *HistogramMap) PercentilesAll(perc PercentilesChannel, reset bool) {
	// take the lock as we need to iterate the map of histograms
	hdr.lock.Lock()
	// we may have this lock for a while but nothing in the cmd processing can
	// access it and by the time we release it (and any waiters), the resets
	// have already happened
	defer hdr.lock.Unlock()

	for _, hist := range hdr.hists {
		// send a snapshot command
		hdr.cmds <- command{
			hist:    hist,
			command: cmdPercentiles,
			arg:     perc,
		}

		if reset {
			// request a reset
			hdr.cmds <- command{
				hist:    hist,
				command: cmdReset,
			}
		}
	}

	// use a channel to wait for confirmation of percentiles snapshots
	// completing
	done := make(chan bool)

	// request a sync
	hdr.cmds <- command{
		hist:    nil,
		command: cmdSync,
		arg:     done,
	}

	<-done
	close(done)
}

// RequestReset requests a snapshot of one or more histograms and is non-blocking
func (hdr *HistogramMap) RequestReset(snap SnapshotChannel, names ...string) {
	for _, name := range names {
		// get/create a histogram for name
		hist := hdr.resolveHistogram(name)

		// request a snapshot
		hdr.cmds <- command{
			hist:    hist,
			command: cmdReset,
		}
	}
}

// Reset resets a named histogram
//
//	Notes
//		Reset waits for acknowledgement of the reset
//
func (hdr *HistogramMap) Reset(name string) {
	// get/create a histogram for name
	hist := hdr.resolveHistogram(name)

	// use a channel to wait for confirmation of reset
	done := make(chan bool)
	defer close(done)

	// send a reset command
	hdr.cmds <- command{
		hist:    hist,
		command: cmdReset,
		arg:     done,
	}

	<-done
}

// ResetAll resets all named histograms
//
//	Notes
//		ResetAll waits for acknowledgement of the resets, and
//		effectively blocks all other activity as the lock for the
//		histogram map is held for the duration
//
func (hdr *HistogramMap) ResetAll() {
	// take the lock as we need to iterate the map of histograms
	hdr.lock.Lock()
	// we may have this lock for a while but nothing in the cmd processing can
	// access it and by the time we release it (and any waiters), the resets
	// have already happened
	defer hdr.lock.Unlock()

	for _, hist := range hdr.hists {
		// send a reset command
		hdr.cmds <- command{
			hist:    hist,
			command: cmdReset,
		}
	}

	// use a channel to wait for confirmation of percentiles snapshots
	// completing
	done := make(chan bool)

	// request a sync
	hdr.cmds <- command{
		hist:    nil,
		command: cmdSync,
		arg:     done,
	}

	<-done
	close(done)
}

func (hdr *HistogramMap) Close() map[string]*hdrhistogram.Histogram {
	// take the lock as we need to iterate the map of histograms
	hdr.lock.Lock()
	// we may have this lock for a while but nothing in the cmd processing can
	// access it and by the time we release it (and any waiters), the cmd
	// channel is already closed
	defer hdr.lock.Unlock()

	for _, hist := range hdr.hists {
		// send a stop command to each histogram
		hdr.cmds <- command{
			hist:    hist,
			command: cmdStop,
		}
	}

	// close the channel to terminate processing once queued commands are
	// consumed. Leave the channel as non-nil so any callers will panic.
	// the alternative is to set it to nil and callers will deadlock
	// (RequestRecord is an exception)
	close(hdr.cmds)

	// wait for process to signal (channel is unbuffered so process is blocked
	// until we read)
	<-hdr.done
	close(hdr.done)

	// return the map of histograms by name
	return hdr.hists
}
