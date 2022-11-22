/*
Copyright 2020 PayPal

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package Watcher is responsible for watching latest metrics from metrics provider via a fetcher client.
It exposes an HTTP REST endpoint to get these metrics, in addition to application API via clients
This also uses a fast json parser
*/
package watcher

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/charstal/load-monitor/pkg/metricsprovider"
	"github.com/charstal/load-monitor/pkg/metricstype"
	"github.com/francoispqt/gojay"
	log "github.com/sirupsen/logrus"
)

const (
	BaseUrl        = "/watcher"
	HealthCheckUrl = "/watcher/healthy"
)

type Watcher struct {
	mutex         sync.RWMutex // For thread safe access to cache
	fifteenMinute []metricstype.WatcherMetrics
	tenMinute     []metricstype.WatcherMetrics
	fiveMinute    []metricstype.WatcherMetrics
	cacheSize     int
	client        metricsprovider.MetricsProviderClient
	isStarted     bool // Indicates if the Watcher is started by calling StartWatching()
	shutdown      chan os.Signal
}

// type Window struct {
// 	Duration string `json:"duration"`
// 	Start    int64  `json:"start"`
// 	End      int64  `json:"end"`
// }

// type Metric struct {
// 	Name     string  `json:"name"`             // Name of metric at the provider
// 	Type     string  `json:"type"`             // CPU or Memory
// 	Operator string  `json:"operator"`         // STD or AVE or SUM, etc.
// 	Rollup   string  `json:"rollup,omitempty"` // Rollup used for metric calculation
// 	Value    float64 `json:"value"`            // Value is expected to be in %
// }

// type NodeMetricsMap map[string]NodeMetrics

// type Data struct {
// 	NodeMetricsMap NodeMetricsMap
// }

// type WatcherMetrics struct {
// 	Timestamp int64  `json:"timestamp"`
// 	Window    Window `json:"window"`
// 	Source    string `json:"source"`
// 	Data      Data   `json:"data"`
// }

// type Tags struct {
// }

// type Metadata struct {
// 	DataCenter string `json:"dataCenter,omitempty"`
// }

// type NodeMetrics struct {
// 	Metrics  []Metric `json:"metrics,omitempty"`
// 	Tags     Tags     `json:"tags,omitempty"`
// 	Metadata Metadata `json:"metadata,omitempty"`
// }

// NewWatcher Returns a new initialised Watcher
func NewWatcher(client metricsprovider.MetricsProviderClient) *Watcher {
	sizePerWindow := 5
	return &Watcher{
		mutex:         sync.RWMutex{},
		fifteenMinute: make([]metricstype.WatcherMetrics, 0, sizePerWindow),
		tenMinute:     make([]metricstype.WatcherMetrics, 0, sizePerWindow),
		fiveMinute:    make([]metricstype.WatcherMetrics, 0, sizePerWindow),
		cacheSize:     sizePerWindow,
		client:        client,
		shutdown:      make(chan os.Signal, 1),
	}
}

// StartWatching This function needs to be called to begin actual watching
func (w *Watcher) StartWatching(shutdown chan struct{}) {
	w.mutex.RLock()
	if w.isStarted {
		w.mutex.RUnlock()
		return
	}
	w.mutex.RUnlock()

	fetchOnce := func(duration string) {
		curWindow, metric := w.getCurrentWindow(duration)
		hostMetrics, err := w.client.FetchAllHostsMetrics(curWindow)

		if err != nil {
			log.Errorf("received error while fetching metrics: %v", err)
			return
		}
		log.Debugf("fetched metrics for window: %v", curWindow)

		// TODO： add tags, etc.
		watcherMetrics := metricMapToWatcherMetrics(hostMetrics, w.client.Name(), *curWindow)
		w.appendWatcherMetrics(metric, &watcherMetrics)
	}

	windowWatcher := func(duration string) {
		for {
			fetchOnce(duration)
			// This is assuming fetching of metrics won't exceed more than 1 minute. If it happens we need to throttle rate of fetches
			time.Sleep(time.Minute)
		}
	}

	durations := [3]string{metricstype.FifteenMinutes, metricstype.TenMinutes, metricstype.FiveMinutes}
	for _, duration := range durations {
		// Populate cache initially before returning
		fetchOnce(duration)
		go windowWatcher(duration)
	}

	http.HandleFunc(BaseUrl, w.handler)
	http.HandleFunc(HealthCheckUrl, w.healthCheckHandler)
	server := &http.Server{
		Addr:    ":2020",
		Handler: http.DefaultServeMux,
	}

	go func() {
		log.Warn(server.ListenAndServe())
	}()

	signal.Notify(w.shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-w.shutdown
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Errorf("Unable to shutdown server: %v", err)
		}
		shutdown <- struct{}{}
	}()

	w.mutex.Lock()
	w.isStarted = true
	w.mutex.Unlock()
	log.Info("Started watching metrics")
}

// GetLatestWatcherMetrics It starts from 15 minute window, and falls back to 10 min, 5 min windows subsequently
// if metrics are not present. StartWatching() should be called before calling this.
func (w *Watcher) GetLatestWatcherMetrics(duration string) (*metricstype.WatcherMetrics, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	if !w.isStarted {
		return nil, errors.New("need to call StartWatching() first")
	}

	switch {
	case duration == metricstype.FifteenMinutes && len(w.fifteenMinute) > 0:
		return w.deepCopyWatcherMetrics(&w.fifteenMinute[len(w.fifteenMinute)-1]), nil
	case (duration == metricstype.FifteenMinutes || duration == metricstype.TenMinutes) && len(w.tenMinute) > 0:
		return w.deepCopyWatcherMetrics(&w.tenMinute[len(w.tenMinute)-1]), nil
	case (duration == metricstype.TenMinutes || duration == metricstype.FiveMinutes) && len(w.fiveMinute) > 0:
		return w.deepCopyWatcherMetrics(&w.fiveMinute[len(w.fiveMinute)-1]), nil
	default:
		return nil, errors.New("unable to get any latest metrics")
	}
}

func (w *Watcher) getCurrentWindow(duration string) (*metricstype.Window, *[]metricstype.WatcherMetrics) {
	var curWindow *metricstype.Window
	var watcherMetrics *[]metricstype.WatcherMetrics
	switch duration {
	case metricstype.FifteenMinutes:
		curWindow = metricstype.CurrentFifteenMinuteWindow()
		watcherMetrics = &w.fifteenMinute
	case metricstype.TenMinutes:
		curWindow = metricstype.CurrentTenMinuteWindow()
		watcherMetrics = &w.tenMinute
	case metricstype.FiveMinutes:
		curWindow = metricstype.CurrentFiveMinuteWindow()
		watcherMetrics = &w.fiveMinute
	default:
		log.Error("received unexpected window duration, defaulting to 15m")
		curWindow = metricstype.CurrentFifteenMinuteWindow()
	}
	return curWindow, watcherMetrics
}

func (w *Watcher) appendWatcherMetrics(recentMetrics *[]metricstype.WatcherMetrics, metric *metricstype.WatcherMetrics) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if len(*recentMetrics) == w.cacheSize {
		*recentMetrics = (*recentMetrics)[1:]
	}
	*recentMetrics = append(*recentMetrics, *metric)
}

func (w *Watcher) deepCopyWatcherMetrics(src *metricstype.WatcherMetrics) *metricstype.WatcherMetrics {
	nodeMetricsMap := make(map[string]metricstype.NodeMetrics)
	for host, fetchedMetric := range src.Data.NodeMetricsMap {
		nodeMetric := metricstype.NodeMetrics{
			Metrics: make([]metricstype.Metric, len(fetchedMetric.Metrics)),
			Tags:    fetchedMetric.Tags,
		}
		copy(nodeMetric.Metrics, fetchedMetric.Metrics)
		nodeMetric.Metadata = fetchedMetric.Metadata
		nodeMetricsMap[host] = nodeMetric
	}
	return &metricstype.WatcherMetrics{
		Timestamp: src.Timestamp,
		Window:    src.Window,
		Source:    src.Source,
		Data: metricstype.Data{
			NodeMetricsMap: nodeMetricsMap,
		},
	}
}

// HTTP Handler for BaseUrl endpoint
func (w *Watcher) handler(resp http.ResponseWriter, r *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	metrics, err := w.GetLatestWatcherMetrics(metricstype.FifteenMinutes)
	if metrics == nil {
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			log.Error(err)
			return
		}
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	host := r.URL.Query().Get("host")
	var bytes []byte
	if host != "" {
		if _, ok := metrics.Data.NodeMetricsMap[host]; ok {
			hostMetricsData := make(map[string]metricstype.NodeMetrics)
			hostMetricsData[host] = metrics.Data.NodeMetricsMap[host]
			hostMetrics := metricstype.WatcherMetrics{Timestamp: metrics.Timestamp,
				Window: metrics.Window,
				Source: metrics.Source,
				Data:   metricstype.Data{NodeMetricsMap: hostMetricsData},
			}
			bytes, err = gojay.MarshalJSONObject(&hostMetrics)
		} else {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		bytes, err = gojay.MarshalJSONObject(metrics)
	}

	if err != nil {
		log.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = resp.Write(bytes)
	if err != nil {
		log.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
	}
}

// Simple server status handler
func (w *Watcher) healthCheckHandler(resp http.ResponseWriter, r *http.Request) {
	if status, err := w.client.Healthy(); status != 0 {
		log.Warnf("health check failed with: %v", err)
		resp.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	resp.WriteHeader(http.StatusOK)
}

// Utility functions

func metricMapToWatcherMetrics(metricMap map[string][]metricstype.Metric, clientName string, window metricstype.Window) metricstype.WatcherMetrics {
	metricsMap := make(map[string]metricstype.NodeMetrics)
	for host, metricList := range metricMap {
		nodeMetric := metricstype.NodeMetrics{
			Metrics: make([]metricstype.Metric, len(metricList)),
		}
		copy(nodeMetric.Metrics, metricList)
		metricsMap[host] = nodeMetric
	}

	watcherMetrics := metricstype.WatcherMetrics{Timestamp: time.Now().Unix(),
		Data:   metricstype.Data{NodeMetricsMap: metricsMap},
		Source: clientName,
		Window: window,
	}
	return watcherMetrics
}
