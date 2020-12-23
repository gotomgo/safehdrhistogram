package safehdrhistogram

import (
	"fmt"
	"io"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// PercentilesChannel is used for non-blocking percentiles requests
type PercentilesChannel chan *Percentiles

// Percentile represents a percentile, it's value, and cumulative count
type Percentile struct {
	Value      int64   `json:"value"`
	Percentile float64 `json:"percentile"`
	Count      int64   `json:"count"`
}

// Percentiles represents a percentiles snapshot of a hdrhistogram.Histogram
//
//	Notes
//		Percentiles are ordered lowest percentile to highest
//
type Percentiles struct {
	MinValue    int64        `json:"minValue"`
	MaxValue    int64        `json:"maxValue"`
	TotalCount  int64        `json:"totalCount"`
	Percentiles []Percentile `json:"percentiles"`
	StartTime   int64        `json:"startTime"`
	EndTime     int64        `json:"endTime"`
	Tag         string       `json:"tag"`
}

// Write produces reasonably well formatted output for Percentiles
func (p *Percentiles) Write(writer io.Writer) (err error) {
	_, err = writer.Write([]byte(fmt.Sprintf("%12s %12s %12s\n", "Value", "Percentile", "TotalCount")))
	if err != nil {
		return
	}

	for _, perc := range p.Percentiles {
		_, err = writer.Write([]byte(fmt.Sprintf("%12d %12f %12d\n", perc.Value, perc.Percentile, perc.Count)))
		if err != nil {
			return
		}
	}

	footer := fmt.Sprintf("  [Min = %d, Max = %d, Total count = %d]\n",
		p.MinValue,
		p.MaxValue,
		p.TotalCount,
	)
	_, err = writer.Write([]byte(footer))

	return
}

// CreatePercentiles creates an instance of Percentiles from a
// hdrhistogram.Histogram
func CreatePercentiles(hist *hdrhistogram.Histogram) (result *Percentiles) {
	result = &Percentiles{
		MinValue:   hist.Min(),
		MaxValue:   hist.Max(),
		TotalCount: hist.TotalCount(),
		StartTime:  hist.StartTimeMs(),
		EndTime:    time.Now().UTC().UnixNano() / 1e6,
		Tag:        hist.Tag(),
	}

	dist := hist.CumulativeDistributionWithTicks(1)
	for _, slice := range dist {
		result.Percentiles = append(
			result.Percentiles,
			Percentile{
				Percentile: slice.Quantile / 100.0,
				Value:      slice.ValueAt,
				Count:      slice.Count,
			})
	}

	return
}
