package safehdrhistogram

import (
	"github.com/HdrHistogram/hdrhistogram-go"
)

// Histogram is a wrapper around hdrhistogram.Histogram that uses a
// command channel to safely manipulate the histogram when there are multiple
// writers
type Histogram struct {
	hist *hdrhistogram.Histogram
	cmds chan command
	done chan bool
}

// NewHistogram creates an instance of hdrhistogram.Histogram that is
// safe for concurrent use
func NewHistogram(
	lowestDiscernibleValue,
	highestTrackableValue int64,
	numberOfSignificantValueDigits int) *Histogram {

	return NewHistogramFromConfig(
		HistogramConfig{
			LowestDiscernibleValue:         lowestDiscernibleValue,
			HighestTrackableValue:          highestTrackableValue,
			NumberOfSignificantValueDigits: numberOfSignificantValueDigits,
			CommandBufferSize:              DefaultCommandBufferSize,
		})
}

// NewHistogramFromConfig creates an instance of hdrhistogram.Histogram that is
// safe for concurrent use based on values from a HistogramConfig
func NewHistogramFromConfig(config HistogramConfig) *Histogram {
	hdr := &Histogram{
		hist: hdrhistogram.New(
			config.LowestDiscernibleValue,
			config.HighestTrackableValue,
			config.NumberOfSignificantValueDigits),
		done: make(chan bool),
		cmds: make(chan command, config.CommandBufferSize),
	}

	// start the cmd processor using the done channel associated with the
	// histogram, which is signalled when all processing completes
	process(hdr.cmds, hdr.done)

	// create a acknowledgment channel for the start command
	done := make(chan bool)
	defer close(done)

	// request a start
	hdr.cmds <- command{
		hist:    hdr.hist,
		command: cmdStart,
		arg:     done,
	}

	// wait for acknowledgement of the start command
	<-done

	return hdr
}

func (hdr *Histogram) WithTag(tag string) *Histogram {
	hdr.hist.SetTag(tag)
	return hdr
}

// Record requests that a value be recorded but will not block if the
// channel is full
//
//	Notes
//		RequestRecord will not block, so if the buffer is full the value is
//		**dropped**
//
func (hdr *Histogram) Record(value int64) {
	select {
	case hdr.cmds <- command{
		hist:    hdr.hist,
		command: cmdRecord,
		arg:     value,
	}:
	default:
		// value is not recorded
	}
}

// RequestSnapshot requests a snapshot of the Histogram but doesn't wait
// for the snapshot
//
//	Notes
//		RequestSnapshot will block if the command buffer is full, but
//		otherwise, the request is made and returns to the caller
//
func (hdr *Histogram) RequestSnapshot(snap SnapshotChannel, reset bool) {
	// request a snapshot. The snap channel will be signalled with the
	// snapshot data when the command is processed
	hdr.cmds <- command{
		hist:    hdr.hist,
		command: cmdSnapshot,
		arg:     snap,
	}

	if reset {
		hdr.cmds <- command{
			hist:    hdr.hist,
			command: cmdReset,
		}
	}
}

// Snapshot blocks until a snapshot request completes
func (hdr *Histogram) Snapshot(reset bool) Snapshot {
	// create a channel for the snapshot
	snap := make(SnapshotChannel)
	defer close(snap)

	// request a snapshot. The snap channel will be signalled with the
	//  snapshot data when the command is processed
	hdr.cmds <- command{
		hist:    hdr.hist,
		command: cmdSnapshot,
		arg:     snap,
	}

	if reset {
		hdr.cmds <- command{
			hist:    hdr.hist,
			command: cmdReset,
		}
	}

	// return the Snapshot
	return <-snap
}

// RequestPercentiles requests a percentiles snapshot of the Histogram but
// doesn't wait for the snapshot
//
//	Notes
//		RequestPercentiles will block if the command buffer is full, but
//		otherwise, the request is made and returns to the caller
//
func (hdr *Histogram) RequestPercentiles(perc PercentilesChannel, reset bool) {
	// request a snapshot. The snap channel will be signalled with the
	// snapshot data when the command is processed
	hdr.cmds <- command{
		hist:    hdr.hist,
		command: cmdPercentiles,
		arg:     perc,
	}

	if reset {
		hdr.cmds <- command{
			hist:    hdr.hist,
			command: cmdReset,
		}
	}
}

// Snapshot blocks until a snapshot request completes
func (hdr *Histogram) Percentiles(reset bool) Percentiles {
	// create a channel for the percentiles
	perc := make(PercentilesChannel)
	defer close(perc)

	// request a snapshot. The snap channel will be signalled with the
	//  snapshot data when the command is processed
	hdr.cmds <- command{
		hist:    hdr.hist,
		command: cmdPercentiles,
		arg:     perc,
	}

	if reset {
		hdr.cmds <- command{
			hist:    hdr.hist,
			command: cmdReset,
		}
	}

	// return the Percentiles
	return <-perc
}

// Reset resets the histogram
//
//	Notes
//		Reset waits for acknowledgement of the reset
//
func (hdr *Histogram) Reset() {
	// use a channel to wait for confirmation of reset
	done := make(chan bool)
	defer close(done)

	// request a reset
	hdr.cmds <- command{
		hist:    hdr.hist,
		command: cmdReset,
		arg:     done,
	}

	// wait for acknowledgement
	<-done
}

// Close closes the command channel, waits for all commands to be processed,
// and returns the final histogram
func (hdr *Histogram) Close() *hdrhistogram.Histogram {
	// send a stop command
	hdr.cmds <- command{
		hist:    hdr.hist,
		command: cmdStop,
	}

	// close the channel to terminate processing once queued commands are
	// consumed. Leave the channel as non-nil so any callers will panic.
	// the alternative is to set it to nil and callers will deadlock
	// (RequestRecord is an exception)
	close(hdr.cmds)

	// done is unbuffered so process will block until we read
	<-hdr.done
	close(hdr.done)

	// return the final histogram
	return hdr.hist
}
