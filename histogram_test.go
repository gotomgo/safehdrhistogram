package safehdrhistogram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Histogram_New(t *testing.T) {
	t.Run("New Histogram", func(t *testing.T) {
		t.Parallel()

		shdr := NewHistogram(1, 30000000, 3)
		if !assert.NotNil(t, shdr, "Histogram should not be nil") {
			return
		}
		if !assert.NotNil(t, shdr.hist, "Histogram.hist should not be nil") {
			return
		}
		if !assert.NotNil(t, shdr.done, "Histogram.done should not be nil") {
			return
		}
		if !assert.Equal(t, int64(1), shdr.hist.LowestTrackableValue(), "Histogram.hist has wrong LowestTrackableValue") {
			return
		}
		if !assert.Equal(t, int64(30000000), shdr.hist.HighestTrackableValue(), "Histogram.hist has wrong HighestTrackableValue") {
			return
		}
		if !assert.Equal(t, int64(3), shdr.hist.SignificantFigures(), "Histogram.hist has wrong SignificantFigures") {
			return
		}
		if !assert.NotNil(t, shdr.cmds, "Histogram.cmds should be non-nil") {
			return
		}
	})
}

func Test_Histogram_Cycle(t *testing.T) {
	t.Run("Stop Histogram", func(t *testing.T) {
		t.Parallel()

		shdr := NewHistogram(1, 30000000, 3)

		hist := shdr.Close()
		if !assert.NotNil(t, hist, "Close() should return a non-nil histogram") {
			return
		}
		if !assert.True(t, hist.StartTimeMs() != 0, "histogram start time should not be zero") {
			return
		}
		if !assert.True(t, hist.EndTimeMs() != 0, "histogram end time should not be zero") {
			return
		}
		if !assert.GreaterOrEqual(t, hist.EndTimeMs(), hist.StartTimeMs(), "histogram end time should be > start time") {
			return
		}

		if !assert.Panics(t, func() { shdr.Record(32) }, "RequestRecord should panic when stopped") {
			return
		}
		if !assert.Panics(t, func() { shdr.Record(32) }, "Record should panic when stopped") {
			return
		}
		if !assert.Panics(t, func() { shdr.Reset() }, "Record should panic when stopped") {
			return
		}
	})
}

func Test_Histogram_Record(t *testing.T) {
	t.Run("Record Histogram", func(t *testing.T) {
		t.Parallel()

		shdr := NewHistogram(1, 30000000, 3)

		// these are the values take from the hdr example
		input := []int64{
			459876, 669187, 711612, 816326, 931423, 1033197, 1131895, 2477317,
			3964974, 12718782,
		}

		for _, sample := range input {
			shdr.Record(sample)
		}

		hist := shdr.Close()
		if !assert.NotNil(t, hist, "Close() should return a non-nil histogram") {
			return
		}

		// this value is from the hdr example that has the same config as used above
		if !assert.Equal(t, int64(931839), hist.ValueAtQuantile(50.0), "histogram value for quartile 50.0 is incorrect") {
			return
		}
	})

	t.Run("Request Record Histogram", func(t *testing.T) {
		t.Parallel()

		shdr := NewHistogram(1, 30000000, 3)

		// these are the values take from the hdr example
		input := []int64{
			459876, 669187, 711612, 816326, 931423, 1033197, 1131895, 2477317,
			3964974, 12718782,
		}

		for _, sample := range input {
			shdr.Record(sample)
		}

		hist := shdr.Close()
		if !assert.NotNil(t, hist, "Close() should return a non-nil histogram") {
			return
		}

		// this value is from the hdr example that has the same config as used above
		if !assert.Equal(t, int64(931839), hist.ValueAtQuantile(50.0), "histogram value for quartile 50.0 is incorrect") {
			return
		}
	})
}

func Test_Histogram_Snapshot(t *testing.T) {
	t.Run("Snapshot Histogram", func(t *testing.T) {
		t.Parallel()

		shdr := NewHistogram(1, 30000000, 3)

		// these are the values take from the hdr example
		input := []int64{
			459876, 669187, 711612, 816326, 931423, 1033197, 1131895, 2477317,
			3964974, 12718782,
		}

		for _, sample := range input {
			shdr.Record(sample)
		}

		snapshot := shdr.Snapshot(false)

		if !assert.NotNil(t, snapshot.Snapshot, "Snapshot.Snapshot should not be nil") {
			return
		}
		if !assert.True(t, snapshot.StartTime != 0, "Snapshot start time should not be zero") {
			return
		}
		if !assert.True(t, snapshot.EndTime != 0, "Snapshot end time should not be zero") {
			return
		}
		if !assert.GreaterOrEqual(t, snapshot.EndTime, snapshot.StartTime, "Snapshot end time should be > start time") {
			return
		}

		// import and verify 50.0 quartile
		hist := snapshot.ToHistogram()
		// this value is from the hdr example that has the same config as used above
		if !assert.Equal(t, int64(931839), hist.ValueAtQuantile(50.0), "histogram value for quartile 50.0 is incorrect") {
			return
		}
		if !assert.True(t, hist.StartTimeMs() != 0, "histogram start time should not be zero") {
			return
		}
		if !assert.True(t, hist.EndTimeMs() != 0, "histogram end time should not be zero") {
			return
		}
		if !assert.GreaterOrEqual(t, hist.EndTimeMs(), hist.StartTimeMs(), "histogram end time should be > start time") {
			return
		}

		hist = shdr.Close()
		if !assert.NotNil(t, hist, "Close() should return a non-nil histogram") {
			return
		}

		// this value is from the hdr example that has the same config as used above
		if !assert.Equal(t, int64(931839), hist.ValueAtQuantile(50.0), "histogram value for quartile 50.0 is incorrect") {
			return
		}
	})

	t.Run("Request Snapshot Histogram", func(t *testing.T) {
		t.Parallel()

		shdr := NewHistogram(1, 30000000, 3)

		// these are the values take from the hdr example
		input := []int64{
			459876, 669187, 711612, 816326, 931423, 1033197, 1131895, 2477317,
			3964974, 12718782,
		}

		for _, sample := range input {
			shdr.Record(sample)
		}

		snap := make(SnapshotChannel)
		shdr.RequestSnapshot(snap, false)
		snapshot := <-snap

		if !assert.NotNil(t, snapshot.Snapshot, "Snapshot.Snapshot should not be nil") {
			return
		}
		if !assert.True(t, snapshot.StartTime != 0, "Snapshot start time should not be zero") {
			return
		}
		if !assert.True(t, snapshot.EndTime != 0, "Snapshot end time should not be zero") {
			return
		}
		if !assert.GreaterOrEqual(t, snapshot.EndTime, snapshot.StartTime, "Snapshot end time should be > start time") {
			return
		}

		// import and verify 50.0 quartile
		hist := snapshot.ToHistogram()
		// this value is from the hdr example that has the same config as used above
		if !assert.Equal(t, int64(931839), hist.ValueAtQuantile(50.0), "histogram value for quartile 50.0 is incorrect") {
			return
		}
		if !assert.True(t, hist.StartTimeMs() != 0, "histogram start time should not be zero") {
			return
		}
		if !assert.True(t, hist.EndTimeMs() != 0, "histogram end time should not be zero") {
			return
		}
		if !assert.GreaterOrEqual(t, hist.EndTimeMs(), hist.StartTimeMs(), "histogram end time should be > start time") {
			return
		}

		hist = shdr.Close()
		if !assert.NotNil(t, hist, "Close() should return a non-nil histogram") {
			return
		}

		// this value is from the hdr example that has the same config as used above
		if !assert.Equal(t, int64(931839), hist.ValueAtQuantile(50.0), "histogram value for quartile 50.0 is incorrect") {
			return
		}
	})
}

func Test_Histogram_Percentiles(t *testing.T) {
	t.Run("Percentiles Snapshot Histogram", func(t *testing.T) {
		t.Parallel()

		shdr := NewHistogram(1, 30000000, 3)

		// these are the values take from the hdr example
		input := []int64{
			459876, 669187, 711612, 816326, 931423, 1033197, 1131895, 2477317,
			3964974, 12718782,
		}

		for _, sample := range input {
			shdr.Record(sample)
		}

		percentiles := shdr.Percentiles(false)

		if !assert.NotNil(t, percentiles, "Snapshot.Snapshot should not be nil") {
			return
		}
		if !assert.True(t, percentiles.StartTime != 0, "Snapshot start time should not be zero") {
			return
		}
		if !assert.True(t, percentiles.EndTime != 0, "Snapshot end time should not be zero") {
			return
		}
		if !assert.GreaterOrEqual(t, percentiles.EndTime, percentiles.StartTime, "Snapshot end time should be > start time") {
			return
		}
		if !assert.Equal(t, int64(len(input)), percentiles.TotalCount, "Percentiles TotalCount should be %d", len(input)) {
			return
		}
		if !assert.Equal(t, int64(459776), percentiles.MinValue, "Percentiles Min should be %d", 459776) {
			return
		}
		if !assert.Equal(t, int64(12722175), percentiles.MaxValue, "Percentiles Max should be %d", 12722175) {
			return
		}

		hist := shdr.Close()
		if !assert.NotNil(t, hist, "Close() should return a non-nil histogram") {
			return
		}

		// this value is from the hdr example that has the same config as used above
		if !assert.Equal(t, int64(931839), hist.ValueAtQuantile(50.0), "histogram value for quartile 50.0 is incorrect") {
			return
		}
	})

	t.Run("Request Percentiles Snapshot Histogram", func(t *testing.T) {
		t.Parallel()

		shdr := NewHistogram(1, 30000000, 3)

		// these are the values take from the hdr example
		input := []int64{
			459876, 669187, 711612, 816326, 931423, 1033197, 1131895, 2477317,
			3964974, 12718782,
		}

		for _, sample := range input {
			shdr.Record(sample)
		}

		perc := make(PercentilesChannel)
		shdr.RequestPercentiles(perc, false)
		percentiles := <-perc

		if !assert.NotNil(t, percentiles, "Snapshot.Snapshot should not be nil") {
			return
		}
		if !assert.True(t, percentiles.StartTime != 0, "Snapshot start time should not be zero") {
			return
		}
		if !assert.True(t, percentiles.EndTime != 0, "Snapshot end time should not be zero") {
			return
		}
		if !assert.GreaterOrEqual(t, percentiles.EndTime, percentiles.StartTime, "Snapshot end time should be > start time") {
			return
		}
		if !assert.Equal(t, int64(len(input)), percentiles.TotalCount, "Percentiles TotalCount should be %d", len(input)) {
			return
		}
		if !assert.Equal(t, int64(459776), percentiles.MinValue, "Percentiles Min should be %d", 459776) {
			return
		}
		if !assert.Equal(t, int64(12722175), percentiles.MaxValue, "Percentiles Max should be %d", 12722175) {
			return
		}

		hist := shdr.Close()
		if !assert.NotNil(t, hist, "Close() should return a non-nil histogram") {
			return
		}

		// this value is from the hdr example that has the same config as used above
		if !assert.Equal(t, int64(931839), hist.ValueAtQuantile(50.0), "histogram value for quartile 50.0 is incorrect") {
			return
		}
	})
}

func Test_Histogram_Reset(t *testing.T) {
	t.Run("Record Histogram", func(t *testing.T) {
		t.Parallel()

		shdr := NewHistogram(1, 30000000, 3)

		// these are the values take from the hdr example
		input := []int64{
			459876, 669187, 711612, 816326, 931423, 1033197, 1131895, 2477317,
			3964974, 12718782,
		}

		for _, sample := range input {
			shdr.Record(sample)
		}

		shdr.Reset()

		hist := shdr.Close()
		if !assert.NotNil(t, hist, "Close() should return a non-nil histogram") {
			return
		}

		// presume that an empty histogram returns 0 for every quartile?
		if !assert.Equal(t, int64(0), hist.ValueAtQuantile(50.0), "histogram value for quartile 50.0 is incorrect after reset") {
			return
		}
	})
}
