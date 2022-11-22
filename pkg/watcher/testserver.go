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

package watcher

import (
	"github.com/charstal/load-monitor/pkg/metricsprovider"
	"github.com/charstal/load-monitor/pkg/metricstype"
)

var FifteenMinutesMetricsMap = map[string][]metricstype.Metric{
	FirstNode: {
		{
			Name:  "test-cpu",
			Type:  metricstype.CPU,
			Value: 26,
		},
	},
	SecondNode: {
		{
			Name:  "test-cpu",
			Type:  metricstype.CPU,
			Value: 60,
		},
	},
}

var TenMinutesMetricsMap = map[string][]metricstype.Metric{
	FirstNode: {
		{
			Name:  "test-cpu",
			Type:  metricstype.CPU,
			Value: 22,
		},
	},
	SecondNode: {
		{
			Name:  "test-cpu",
			Type:  metricstype.CPU,
			Value: 65,
		},
	},
}

var FiveMinutesMetricsMap = map[string][]metricstype.Metric{
	FirstNode: {
		{
			Name:  "test-cpu",
			Type:  metricstype.CPU,
			Value: 21,
		},
	},
	SecondNode: {
		{
			Name:  "test-cpu",
			Type:  metricstype.CPU,
			Value: 50,
		},
	},
}

var _ metricsprovider.MetricsProviderClient = &testServerClient{}

const (
	FirstNode            = "worker-1"
	SecondNode           = "worker-2"
	TestServerClientName = "TestServerClient"
)

type testServerClient struct {
}

func (t testServerClient) Name() string {
	return TestServerClientName
}

func NewTestMetricsServerClient() metricsprovider.MetricsProviderClient {
	return testServerClient{}
}

func (t testServerClient) FetchHostMetrics(host string, window *metricstype.Window) ([]metricstype.Metric, error) {
	if _, ok := FifteenMinutesMetricsMap[host]; !ok {
		return nil, nil
	}
	if _, ok := TenMinutesMetricsMap[host]; !ok {
		return nil, nil
	}
	if _, ok := FiveMinutesMetricsMap[host]; !ok {
		return nil, nil
	}

	if window.Duration == metricstype.TenMinutes {
		return TenMinutesMetricsMap[host], nil
	} else if window.Duration == metricstype.FiveMinutes {
		return FiveMinutesMetricsMap[host], nil
	}

	return FifteenMinutesMetricsMap[host], nil
}

func (t testServerClient) FetchAllHostsMetrics(window *metricstype.Window) (map[string][]metricstype.Metric, error) {
	if window.Duration == metricstype.TenMinutes {
		return TenMinutesMetricsMap, nil
	} else if window.Duration == metricstype.FiveMinutes {
		return FiveMinutesMetricsMap, nil
	}

	return FifteenMinutesMetricsMap, nil
}

func (t testServerClient) Healthy() (int, error) {
	return 0, nil
}
