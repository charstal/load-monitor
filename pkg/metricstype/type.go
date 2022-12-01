package metricstype

import "time"

const (
	// time
	FifteenMinutes = "15m"
	TenMinutes     = "10m"
	FiveMinutes    = "5m"

	// operator
	Average         = "Avg"
	Std             = "Std"
	Latest          = "Latest"
	UnknownOperator = "Unknown"
	Capacity        = "Capacity"

	// Resource
	CPU               = "cpu"
	Memory            = "memory"
	Ephemeral_storage = "ephemeral_storage"
	Hugepages_2Mi     = "hugepages_2Mi"
	Pods              = "pods"
	Network           = "network"

	// unit
	Ratio   = "ratio"
	Core    = "core"
	Integer = "integer"
	Bytes   = "bytes"
	MiB     = "MiB"
	M       = "m"
)

type Window struct {
	Duration string `json:"duration"`
	Start    int64  `json:"start"`
	End      int64  `json:"end"`
}

type Metric struct {
	Name     string  `json:"name"`               // Name of metric at the provider
	Type     string  `json:"type,omitempty"`     // CPU or Memory
	Operator string  `json:"operator,omitempty"` // STD or AVE or SUM, etc.
	Rollup   string  `json:"rollup,omitempty"`   // Rollup used for metric calculation
	Unit     string  `json:"unit,omitempty"`     // Unit of Value
	Value    float64 `json:"value"`              // Value is expected to be in %
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
