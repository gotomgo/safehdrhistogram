package safehdrhistogram

import (
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// commandType is an alias for int
type commandType int

const (
	// cmdStart initializes the Histogram for use
	cmdStart commandType = iota
	// cmdRecord records a value to the histogram
	cmdRecord
	// cmdSnapshot requests a snapshot of the histogram
	cmdSnapshot
	// cmdPercentiles requests a percentile distribution summary
	cmdPercentiles
	// cmdReset requests a reset of the histogram
	cmdReset
	// cmdSync allows for waiting for the command to be processed
	cmdSync
	// cmdStop indicates that the histogram should be finalized
	cmdStop
)

// command represents a command to be processed. Commands operate on a
// specific hdrhistogram.Histogram and generally include an argument (such as
// the value to be recorded, or an acknowledgement channel)
type command struct {
	hist    *hdrhistogram.Histogram
	command commandType
	arg     interface{}
}

// process starts a go routine to process commands on the channel
//
//	Notes
//		Processing terminates when the channel is closed. If the caller
//		provides a done channel, that channel will be signaled after
//		processing stops
//
func process(commands <-chan command, done chan bool) {
	go func(commands <-chan command, done chan bool) {
		for cmd := range commands {
			if err := processCommand(cmd); err != nil {
				// TODO: do something with OutOfRange errors that can occur
				// when recording a value. Not much more than a warning IMO
			}
		}

		if done != nil {
			done <- true
		}
	}(commands, done)
}

// processCommand executes the actions related to a command
func processCommand(cmd command) (err error) {
	switch cmd.command {
	case cmdStart:
		cmd.hist.SetStartTimeMs(time.Now().UTC().UnixNano() / 1e6)

		if cmd.arg != nil {
			cmd.arg.(chan bool) <- true
		}
	case cmdStop:
		cmd.hist.SetEndTimeMs(time.Now().UTC().UnixNano() / 1e6)

		if cmd.arg != nil {
			cmd.arg.(chan bool) <- true
		}
	case cmdRecord:
		err = cmd.hist.RecordValue(cmd.arg.(int64))
	case cmdSnapshot:
		cmd.arg.(SnapshotChannel) <- getSnapshot(cmd.hist)
	case cmdPercentiles:
		cmd.arg.(PercentilesChannel) <- getPercentiles(cmd.hist)
	case cmdSync:
		cmd.arg.(chan bool) <- true
	case cmdReset:
		cmd.hist.Reset()
		cmd.hist.SetStartTimeMs(time.Now().UTC().UnixNano() / 1e6)

		if cmd.arg != nil {
			cmd.arg.(chan bool) <- true
		}
	}

	return
}
