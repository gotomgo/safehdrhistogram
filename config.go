package safehdrhistogram

// DefaultCommandBufferSize is the default size of the command buffer used to
// process commands, such as recording a value to a histogram
const DefaultCommandBufferSize = 256

// HistogramConfig represents the values used to construct a
// Histogram and is designed for use in yaml or JSON configuration files
type HistogramConfig struct {
	LowestDiscernibleValue         int64 `yaml:"lowestDiscernibleValue" json:"lowestDiscernibleValue"`
	HighestTrackableValue          int64 `yaml:"highestTrackableValue" json:"highestTrackableValue"`
	NumberOfSignificantValueDigits int   `yaml:"numberOfSignificantValueDigits" json:"numberOfSignificantValueDigits"`
	CommandBufferSize              int   `yaml:"commandBufferSize" json:"commandBufferSize"`
}
