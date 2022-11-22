package metricstype

import "time"

const (
	FifteenMinutes  = "15m"
	TenMinutes      = "10m"
	FiveMinutes     = "5m"
	CPU             = "CPU"
	Memory          = "Memory"
	Average         = "AVG"
	Std             = "STD"
	Latest          = "Latest"
	UnknownOperator = "Unknown"
)

type Window struct {
	Duration string `json:"duration"`
	Start    int64  `json:"start"`
	End      int64  `json:"end"`
}

type Metric struct {
	Name     string  `json:"name"`             // Name of metric at the provider
	Type     string  `json:"type"`             // CPU or Memory
	Operator string  `json:"operator"`         // STD or AVE or SUM, etc.
	Rollup   string  `json:"rollup,omitempty"` // Rollup used for metric calculation
	Value    float64 `json:"value"`            // Value is expected to be in %
}

type NodeMetricsMap map[string]NodeMetrics

type Data struct {
	NodeMetricsMap NodeMetricsMap
}

type WatcherMetrics struct {
	Timestamp int64  `json:"timestamp"`
	Window    Window `json:"window"`
	Source    string `json:"source"`
	Data      Data   `json:"data"`
}

type Tags struct {
}

type Metadata struct {
	DataCenter string `json:"dataCenter,omitempty"`
}

type NodeMetrics struct {
	Metrics  []Metric `json:"metrics,omitempty"`
	Tags     Tags     `json:"tags,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

func CurrentFifteenMinuteWindow() *Window {
	curTime := time.Now().Unix()
	return &Window{FifteenMinutes, curTime - 15*60, curTime}
}

func CurrentTenMinuteWindow() *Window {
	curTime := time.Now().Unix()
	return &Window{TenMinutes, curTime - 10*60, curTime}
}

func CurrentFiveMinuteWindow() *Window {
	curTime := time.Now().Unix()
	return &Window{FiveMinutes, curTime - 5*60, curTime}
}
