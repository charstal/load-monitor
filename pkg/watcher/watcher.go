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
	"os"
	"sync"
	"time"

	"github.com/charstal/load-monitor/pkg/metricsprovider"
	"github.com/charstal/load-monitor/pkg/metricstype"
	"github.com/charstal/load-monitor/pkg/statistics"
	"github.com/charstal/load-monitor/pkg/storage"
	log "github.com/sirupsen/logrus"
)

type Watcher struct {
	mutex            sync.RWMutex // For thread safe access to cache
	fifteenMinute    []metricstype.WatcherMetrics
	tenMinute        []metricstype.WatcherMetrics
	fiveMinute       []metricstype.WatcherMetrics
	cacheSize        int
	client           metricsprovider.MetricsProviderClient
	isStarted        bool // Indicates if the Watcher is started by calling StartWatching()
	shutdown         chan os.Signal
	statisticsReader *statistics.OfflineReader
	storage          *storage.Storage
}

// NewWatcher Returns a new initialised Watcher
func NewWatcher(client metricsprovider.MetricsProviderClient) *Watcher {
	sizePerWindow := 5

	// init statistics offline reader
	statistics, err := statistics.NewOfflineReader()
	if err != nil {
		panic("statistic init error")
	}

	storage, err := storage.NewStorage()
	if err != nil {
		panic("storage init error")
	}

	return &Watcher{
		mutex:            sync.RWMutex{},
		fifteenMinute:    make([]metricstype.WatcherMetrics, 0, sizePerWindow),
		tenMinute:        make([]metricstype.WatcherMetrics, 0, sizePerWindow),
		fiveMinute:       make([]metricstype.WatcherMetrics, 0, sizePerWindow),
		cacheSize:        sizePerWindow,
		client:           client,
		shutdown:         make(chan os.Signal, 1),
		statisticsReader: statistics,
		storage:          storage,
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

		watcherMetrics := metricMapToWatcherMetrics(hostMetrics, w.statisticsReader.GetMetrics(), w.client.Name(), *curWindow)
		w.appendWatcherMetrics(metric, &watcherMetrics)
	}

	windowWatcher := func(duration string) {
		for {
			fetchOnce(duration)
			// This is assuming fetching of metrics won't exceed more than 1 minute. If it happens we need to throttle rate of fetches
			time.Sleep(time.Minute)
		}
	}

	// fetch statistic
	w.statisticsReader.Update()

	// w.storage.Test()

	durations := [3]string{metricstype.FifteenMinutes, metricstype.TenMinutes, metricstype.FiveMinutes}
	for _, duration := range durations {
		// Populate cache initially before returning
		fetchOnce(duration)
		go windowWatcher(duration)
	}
	// start http server
	w.startHttpServer(shutdown)

	w.mutex.Lock()
	w.isStarted = true
	w.mutex.Unlock()
	log.Info("Started watching metrics")
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
		curWindow = metricstype.CurrentFiveMinuteWindow()
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

	statisticMetricsMap := make(map[string]metricstype.NodeMetrics)
	for host, fetchedMetric := range src.Statistics.NodeMetricsMap {
		nodeMetric := metricstype.NodeMetrics{
			Metrics: make([]metricstype.Metric, len(fetchedMetric.Metrics)),
			Tags:    fetchedMetric.Tags,
		}
		copy(nodeMetric.Metrics, fetchedMetric.Metrics)
		nodeMetric.Metadata = fetchedMetric.Metadata
		statisticMetricsMap[host] = nodeMetric
	}

	return &metricstype.WatcherMetrics{
		Timestamp: src.Timestamp,
		Window:    src.Window,
		Source:    src.Source,
		Data: metricstype.Data{
			NodeMetricsMap: nodeMetricsMap,
		},
		Statistics: metricstype.Data{
			NodeMetricsMap: statisticMetricsMap,
		},
	}
}

// Utility functions

func metricMapToWatcherMetrics(metricMap map[string][]metricstype.Metric, statistics *map[string][]metricstype.Metric, clientName string, window metricstype.Window) metricstype.WatcherMetrics {
	metricsMap := make(map[string]metricstype.NodeMetrics)
	for host, metricList := range metricMap {
		nodeMetric := metricstype.NodeMetrics{
			Metrics: make([]metricstype.Metric, len(metricList)),
		}
		copy(nodeMetric.Metrics, metricList)
		metricsMap[host] = nodeMetric
	}

	statisticsMap := make(map[string]metricstype.NodeMetrics)
	if statistics != nil {
		for host, metricList := range *statistics {
			nodeMetric := metricstype.NodeMetrics{
				Metrics: make([]metricstype.Metric, len(metricList)),
			}
			copy(nodeMetric.Metrics, metricList)
			statisticsMap[host] = nodeMetric
		}
	}

	watcherMetrics := metricstype.WatcherMetrics{
		Timestamp:  time.Now().Unix(),
		Data:       metricstype.Data{NodeMetricsMap: metricsMap},
		Source:     clientName,
		Window:     window,
		Statistics: metricstype.Data{NodeMetricsMap: statisticsMap},
	}
	// fmt.Printf("%v", watcherMetrics)
	return watcherMetrics
}
