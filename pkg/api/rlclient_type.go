package api

import (
	"github.com/charstal/load-monitor/pkg/metricstype"
	"github.com/francoispqt/gojay"
)

type RLClientPredictRequest struct {
	PodName  string   `json:"pod_name"`
	PodLabel string   `json:"pod_label"`
	Nodes    []string `json:"nodes"`
}

type PredictResponceResult struct {
	PodName string `json:"pod"`
	Node    string `json:"node"`
}

type RLClientPredictResponce struct {
	State   string                `json:"state"`
	Message string                `json:"message"`
	Result  PredictResponceResult `json:"result"`
}

type RLClientUpdateRequest struct {
	Timestamp int64                      `json:"tiemstamp"`
	PodName   string                     `json:"pod_name"`
	Metrics   metricstype.WatcherMetrics `json:"metrics"`
}

func (rl *PredictResponceResult) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "pod":
		return dec.String(&rl.PodName)
	case "node":
		return dec.String(&rl.Node)
	}
	return nil
}

func (rl *RLClientPredictResponce) NKeys() int {
	return 2
}

func (rl *RLClientPredictResponce) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "state":
		return dec.String(&rl.State)
	case "message":
		return dec.String(&rl.Message)
	case "result":
		err := dec.Object(&rl.Result)
		return err
	}
	return nil
}

func (rl *PredictResponceResult) NKeys() int {
	return 3
}
