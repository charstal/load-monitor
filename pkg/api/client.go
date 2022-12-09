/*
Copyright 2021 PayPal

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

package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/charstal/load-monitor/pkg/metricsprovider"
	"github.com/charstal/load-monitor/pkg/metricstype"
	"github.com/charstal/load-monitor/pkg/watcher"
	"github.com/francoispqt/gojay"

	"k8s.io/klog/v2"
)

const (
	httpClientTimeoutSecond = 55 * time.Second
)

// Client for Watcher APIs as a library
type libraryClient struct {
	fetcherClient metricsprovider.MetricsProviderClient
	watcher       *watcher.Watcher
}

// Client for Watcher APIs as a service
type serviceClient struct {
	httpClient     http.Client
	watcherAddress string
}

// Creates a new watcher client when using watcher as a library
func NewLibraryClient(opts metricsprovider.MetricsProviderOpts) (Client, error) {
	var err error
	client := libraryClient{}
	switch opts.Name {
	case metricsprovider.PromClientName:
		client.fetcherClient, err = metricsprovider.NewPromClient(opts)
	case metricsprovider.SignalFxClientName:
		client.fetcherClient, err = metricsprovider.NewSignalFxClient(opts)
	default:
		client.fetcherClient, err = metricsprovider.NewMetricsServerClient()
	}
	if err != nil {
		return client, err
	}
	client.watcher = watcher.NewWatcher(client.fetcherClient)
	ch := make(chan struct{})
	client.watcher.StartWatching(ch)
	return client, nil
}

// Creates a new watcher client when using watcher as a service
func NewServiceClient(watcherAddress string) (Client, error) {
	return serviceClient{
		httpClient: http.Client{
			Timeout: httpClientTimeoutSecond,
		},
		watcherAddress: watcherAddress,
	}, nil
}

func (c libraryClient) Healthy() error {
	return c.watcher.Healthy()
}

func (c libraryClient) GetLatestWatcherMetrics() (*metricstype.WatcherMetrics, error) {
	return c.watcher.GetLatestWatcherMetrics(metricstype.FifteenMinutes)
}

func (c libraryClient) GetCompactWatcherMetrics() (*metricstype.WatcherMetrics, error) {
	// Todo
	panic("unimplement")
	return nil, nil
}

func (c serviceClient) Healthy() error {
	req, err := http.NewRequest(http.MethodGet, c.watcherAddress+watcher.HealthCheckUrl, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	klog.V(6).Infof("received status code %v from watcher", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		err = fmt.Errorf("received status code %v from watcher", resp.StatusCode)
		klog.Error(err)
		return err
	}
}

func (c serviceClient) GetCompactWatcherMetrics() (*metricstype.WatcherMetrics, error) {
	// Todo
	panic("unimplement")
	return nil, nil
}

func (c serviceClient) GetLatestWatcherMetrics() (*metricstype.WatcherMetrics, error) {
	req, err := http.NewRequest(http.MethodGet, c.watcherAddress+watcher.BaseUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	//TODO(aqadeer): Add a couple of retries for transient errors
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	klog.V(6).Infof("received status code %v from watcher", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		data := metricstype.Data{NodeMetricsMap: make(map[string]metricstype.NodeMetrics)}
		statisticsDate := metricstype.StatisticsData{StatisticsMap: make(map[string]metricstype.NodeMetrics)}
		metrics := metricstype.WatcherMetrics{Data: data, Statistics: statisticsDate}
		dec := gojay.BorrowDecoder(resp.Body)
		defer dec.Release()
		err = dec.Decode(&metrics)
		if err != nil {
			klog.Errorf("unable to decode watcher metrics: %v", err)
			return nil, err
		} else {
			return &metrics, nil
		}
	} else {
		err = fmt.Errorf("received status code %v from watcher", resp.StatusCode)
		klog.Error(err)
		return nil, err
	}

}
