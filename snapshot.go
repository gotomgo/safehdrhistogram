package safehdrhistogram

import (
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// SnapshotChannel is used for non-blocking snapshot requests
type SnapshotChannel chan *Snapshot

// Snapshot represents a snapshot of a hdrhistogram.Histogram
type Snapshot struct {
	Snapshot  *hdrhistogram.Snapshot
	StartTime int64
	EndTime   int64
	Tag       string
}

// ToHistogram converts an Snapshot to a hdrhistogram.Histogram
func (snapshot *Snapshot) ToHistogram() (result *hdrhistogram.Histogram) {
	result = hdrhistogram.Import(snapshot.Snapshot)
	result.SetStartTimeMs(snapshot.StartTime)
	result.SetEndTimeMs(snapshot.EndTime)
	result.SetTag(snapshot.Tag)
	return
}

func getSnapshot(hist *hdrhistogram.Histogram) *Snapshot {
	return &Snapshot{
		Snapshot:  hist.Export(),
		StartTime: hist.StartTimeMs(),
		EndTime:   time.Now().UTC().UnixNano() / 1e6,
		Tag:       hist.Tag(),
	}
}
