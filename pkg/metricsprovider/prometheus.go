/*
Copyright 2020

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

package metricsprovider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"

	"github.com/charstal/load-monitor/pkg/metricstype"

	cfg "github.com/charstal/load-monitor/pkg/config"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type PromMethod = string
type PromResource = string
type PromSQL = string

const (
	baseHealthyURL      = "/-/healthy"
	EnableOpenShiftAuth = "ENABLE_OPENSHIFT_AUTH"
	KubeConfig          = "KUBE_CONFIG"
)

const (
	Std    PromMethod = "stddev_over_time"
	Avg    PromMethod = "avg_over_time"
	Latest PromMethod = "Latest"
)

const (
	KubeNodeStatusCapacity PromResource = "kube_node_status_capacity"
	NodeCpuRatio           PromResource = "instance:node_cpu:ratio"
	NodeMemoryUtilRatio    PromResource = "instance:node_memory_utilisation:ratio"
	NodeRunningPodCount    PromResource = "kubelet_running_pods"

	NodeCpuRateSum           PromResource = "instance:node_cpu:rate:sum"
	NodeCpuNum               PromResource = "instance:node_num_cpu:sum"
	NodeNetworkReceiveBytes  PromResource = "instance:node_network_receive_bytes:rate:sum"
	NodeNetworkTransmitBytes PromResource = "instance:node_network_transmit_bytes:rate:sum"

	NodeCpuUtilRate5m                         PromResource = "instance:node_cpu_utilisation:rate5m"
	NodeNetworkReceiveBytesExcludinglo5m      PromResource = "instance:node_network_receive_bytes_excluding_lo:rate5m"
	NodeNetworkReceiveDropBytesExcludinglo5m  PromResource = "instance:node_network_receive_drop_excluding_lo:rate5m"
	NodeNetworkTransmitBytesExcludinglo5m     PromResource = "instance:node_network_transmit_bytes_excluding_lo:rate5m"
	NodeNetworkTransmitDropBytesExcludinglo5m PromResource = "instance:node_network_transmit_drop_excluding_lo:rate5m"
	NodeNetworkTotalBytesExcludinglo5m        PromResource = `instance:node_network_receive_bytes_excluding_lo:rate5m
																+instance:node_network_transmit_bytes_excluding_lo:rate5m`
	NodeDiskIOTimeSecondsRate5m         PromResource = "instance_device:node_disk_io_time_seconds:rate5m"
	NodeDiskIOTimeWeightedSecondsRate5m PromResource = "instance_device:node_disk_io_time_weighted_seconds:rate5m"
)

const (
	PromSQLNodeDiskTotalUtilRate PromSQL = "sum by (instance) (rate(node_disk_reads_completed_total[%s]) + rate(node_disk_writes_completed_total[%s]))"
	PromSQLNodeDiskReadUtilRate  PromSQL = "sum by (instance) (rate(node_disk_reads_completed_total[%s]))"
	PromSQLNodeDiskWriteUtilRate PromSQL = "sum by (instance) (rate(node_disk_writes_completed_total[%s]))"

	PromSQLNodeDiskUtilRate5m PromSQL = `sum by (instance) (
		instance_device:node_disk_io_time_seconds:rate5m
	)`

	PromSQLNodeDiskSaturation5m PromSQL = `sum by (kubernetes_node) (
		instance_device:node_disk_io_time_weighted_seconds:rate5m 
		/ scalar(count(instance_device:node_disk_io_time_weighted_seconds:rate5m)))`

	// PromSQLNodeThreads PromSQL = "sum by (node)(container_threads)"
	// need a label to confirm
	// Todo(label)
	// courseLabel                        = metricstype.DEFAULT_COURSE_LABEL
	// nodeNameLabel                      = "kubernetes_node"
	// PromSQLNodePodCountOfLabel PromSQL = "sum by(" + courseLabel + "," + nodeNameLabel + ") (kube_pod_labels)"

	PodClassOfNodeName = "pod_class_of_node_count"
)

var (
	ratioSource = []PromResource{
		NodeCpuRatio,
		NodeMemoryUtilRatio,
	}
	sqlSet5m = []PromResource{
		PromSQLNodeDiskSaturation5m,
		// PromSQLNodeDiskUtilRate5m,
		// NodeCpuUtilRate5m,
		// NodeNetworkReceiveBytesExcludinglo5m,
		// NodeNetworkReceiveDropBytesExcludinglo5m,
		// NodeNetworkTransmitBytesExcludinglo5m,
		// NodeNetworkTransmitDropBytesExcludinglo5m,
		// NodeNetworkTotalBytesExcludinglo5m,
	}
	sqlSetTimes = []PromResource{
		// PromSQLNodeDiskTotalUtilRate,
		// PromSQLNodeDiskReadUtilRate,
		// PromSQLNodeDiskWriteUtilRate,
	}
	sqlNoTime = []PromResource{
		// NodeRunningPodCount,
	}

	sql2NameMap = map[string]string{
		PromSQLNodeDiskTotalUtilRate:              "node_disk_total_util_rate",
		PromSQLNodeDiskReadUtilRate:               "node_disk_read_util_rate",
		PromSQLNodeDiskWriteUtilRate:              "node_disk_write_util_rate",
		PromSQLNodeDiskSaturation5m:               metricstype.NODE_DISK_SATURATION,
		PromSQLNodeDiskUtilRate5m:                 "node_disk_util_rate",
		NodeCpuUtilRate5m:                         "node_cpu_util_rate",
		NodeNetworkReceiveBytesExcludinglo5m:      metricstype.NODE_NETWORK_RECEIVE_BYTES_EXCLUDING_LO,
		NodeNetworkReceiveDropBytesExcludinglo5m:  "node_network_receive_drop_bytes_excluding_lo",
		NodeNetworkTransmitBytesExcludinglo5m:     metricstype.NODE_NETWORK_TRANSMIT_BYTES_EXCLUDING_LO,
		NodeNetworkTransmitDropBytesExcludinglo5m: "node_network_transmit_drop_bytes_excluding_lo",
		NodeNetworkTotalBytesExcludinglo5m:        metricstype.NODE_NETWORK_TOTAL_BYTES_PERCENTAGE_EXCLUDING_LO,
		NodeRunningPodCount:                       "node_running_pod_count",
		// PromSQLNodePodCountOfLabel:                "node_pod_count_of_label",
		// PromSQLNodeThreads:                        "node_thread_count",
	}
)

var (
	promAddress    string
	promToken      string
	promHealthyUrl string
)

type promClient struct {
	client api.Client
}

func NewPromClient(opts MetricsProviderOpts) (MetricsProviderClient, error) {
	if opts.Name != PromClientName {
		return nil, fmt.Errorf("metric provider name should be %v, found %v", PromClientName, opts.Name)
	}

	var client api.Client
	var err error
	promToken, promAddress = "", cfg.DefaultPromAddress
	if opts.AuthToken != "" {
		promToken = opts.AuthToken
	}
	if opts.Address != "" {
		promAddress = opts.Address
	}

	promHealthyUrl = promAddress + baseHealthyURL

	// Ignore TLS verify errors if InsecureSkipVerify is set
	roundTripper := api.DefaultRoundTripper

	// Check if EnableOpenShiftAuth is set.
	_, enableOpenShiftAuth := os.LookupEnv(EnableOpenShiftAuth)
	if enableOpenShiftAuth {
		// Create the config for kubernetes client
		clusterConfig, err := rest.InClusterConfig()
		if err != nil {
			// Get the kubeconfig path
			kubeConfigPath, ok := os.LookupEnv(KubeConfig)
			if !ok {
				kubeConfigPath = cfg.DefaultKubeConfig
			}
			clusterConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
			if err != nil {
				return nil, fmt.Errorf("failed to get kubernetes config: %v", err)
			}
		}

		// Create the client for kubernetes
		kclient, err := kubernetes.NewForConfig(clusterConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
		}

		// Retrieve router CA cert
		routerCAConfigMap, err := kclient.CoreV1().ConfigMaps("openshift-config-managed").Get(context.TODO(), "default-ingress-cert", metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		bundlePEM := []byte(routerCAConfigMap.Data["ca-bundle.crt"])

		// make a client connection configured with the provided bundle.
		roots := x509.NewCertPool()
		roots.AppendCertsFromPEM(bundlePEM)

		// Get Prometheus Host
		u, _ := url.Parse(opts.Address)
		roundTripper = transport.NewBearerAuthRoundTripper(
			opts.AuthToken,
			&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 10 * time.Second,
				TLSClientConfig: &tls.Config{
					RootCAs:    roots,
					ServerName: u.Host,
				},
			},
		)
	} else if opts.InsecureSkipVerify {
		roundTripper = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		}
	}

	if promToken != "" {
		client, err = api.NewClient(api.Config{
			Address:      promAddress,
			RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(opts.AuthToken), roundTripper),
		})
	} else {
		client, err = api.NewClient(api.Config{
			Address: promAddress,
		})
	}

	if err != nil {
		log.Errorf("error creating prometheus client: %v", err)
		return nil, err
	}

	return promClient{client}, err
}

func (s promClient) Name() string {
	return PromClientName
}

func (s promClient) FetchAllHostsMetrics(window *metricstype.Window) (map[string][]metricstype.Metric, error) {
	allMetrics := make(map[string][]metricstype.Metric)
	var anyerr error

	if err := s.fetchCapacity(&allMetrics); err != nil {
		log.Errorf("cannot fetching capacity %v", err)
		anyerr = err
	}

	if err := s.fetchAllCpuMem(window, &allMetrics); err != nil {
		log.Errorf("cannot fetching all cpu and memory %v", err)
	}

	if err := s.fetchFromSql(window, &allMetrics); err != nil {
		log.Errorf("cannot fetching from sql %v", err)
	}
	s.fetchCustomizationMetrics(&allMetrics)

	return allMetrics, anyerr

}

func (s promClient) FetchHostMetrics(host string, window *metricstype.Window) ([]metricstype.Metric, error) {
	return nil, nil
}

func (s promClient) fetchCustomizationMetrics(m *map[string][]metricstype.Metric) {
	if err := s.fetchNodeClassNumOfPodMetrics(m); err != nil {
		log.Errorf("cannot generate node class num of pod %v", err)
	}
}

func (s promClient) fetchNodeClassNumOfPodMetrics(m *map[string][]metricstype.Metric) error {
	runningPodSql := `kube_pod_status_phase{phase="Running"}`
	allRunningPodMap := make(map[string]struct{})
	data, err := s.getPromResults(runningPodSql)
	if err != nil {
		log.Errorf("error querying Prometheus for query %v: %v\n", runningPodSql, err)
		return err
	}
	switch data.(type) {
	case model.Vector:
		for _, result := range data.(model.Vector) {
			if result.Metric["phase"] == "Running" {
				allRunningPodMap[string(result.Metric["pod"])] = struct{}{}
			}
		}
	default:
		log.Errorf("error: The capacity results should not be type: %v.\n", data.Type())
		return nil
	}

	allRunningPod2NodeMap := make(map[string]string)
	node2PodSql := "kube_pod_info"
	data, err = s.getPromResults(node2PodSql)
	if err != nil {
		log.Errorf("error querying Prometheus for query %v: %v\n", runningPodSql, err)
		return err
	}
	switch data.(type) {
	case model.Vector:
		for _, result := range data.(model.Vector) {
			if _, ok := allRunningPodMap[string(result.Metric["pod"])]; ok {
				allRunningPod2NodeMap[string(result.Metric["pod"])] = string(result.Metric["node"])
			}
		}
	default:
		log.Errorf("error: The capacity results should not be type: %v.\n", data.Type())
		return nil
	}

	podLaabelSql := "kube_pod_labels"
	nodePodClassNum := make(map[string]map[string]int)
	data, err = s.getPromResults(podLaabelSql)
	if err != nil {
		log.Errorf("error querying Prometheus for query %v: %v\n", runningPodSql, err)
		return err
	}
	switch data.(type) {
	case model.Vector:
		for _, result := range data.(model.Vector) {
			if label, ok := result.Metric["label_course_id"]; ok {
				label := string(label)
				if node, ok := allRunningPod2NodeMap[string(result.Metric["pod"])]; ok {
					if _, ok := nodePodClassNum[node][label]; !ok {
						mm := make(map[string]int)
						mm[label] = 0
						nodePodClassNum[node] = mm
					}
					nodePodClassNum[node][label] = nodePodClassNum[node][label] + 1
				}
			}
		}
	default:
		log.Errorf("error: The capacity results should not be type: %v.\n", data.Type())
		return nil
	}

	for host, mm := range nodePodClassNum {
		for t, v := range mm {
			curMetric := metricstype.Metric{Name: PodClassOfNodeName, Type: t, Operator: metricstype.Latest,
				Rollup: "", Unit: metricstype.Integer, Value: float64(v)}
			(*m)[host] = append((*m)[host], curMetric)
		}
	}

	return nil
}

func (s promClient) fetchFromSql(window *metricstype.Window, m *map[string][]metricstype.Metric) error {
	var anyerr error
	for _, sql := range sqlSetTimes {
		var newsql string
		if sql == PromSQLNodeDiskTotalUtilRate {
			newsql = fmt.Sprintf(sql, window.Duration, window.Duration)
		} else {
			newsql = fmt.Sprintf(sql, window.Duration)
		}

		results, err := s.getPromResults(newsql)
		if err != nil {
			log.Errorf("error querying Prometheus for query %v: %v\n", newsql, err)
			anyerr = err
			continue
		}

		curMetricMap := s.sqlWithTime2MetricMap(sql, results, "")
		for k, v := range curMetricMap {
			(*m)[k] = append((*m)[k], v...)
		}
	}

	for _, sql := range sqlNoTime {
		results, err := s.getPromResults(sql)
		if err != nil {
			log.Errorf("error querying Prometheus for query %v: %v\n", sql, err)
			anyerr = err
			continue
		}

		curMetricMap := s.sqlWithNoTime2MetricMap(sql, results)
		for k, v := range curMetricMap {
			(*m)[k] = append((*m)[k], v...)
		}
	}

	if window.Duration == metricstype.FiveMinutes {
		for _, sql := range sqlSet5m {
			results, err := s.getPromResults(sql)
			if err != nil {
				log.Errorf("error querying Prometheus for query %v: %v\n", sql, err)
				anyerr = err
				continue
			}

			curMetricMap := s.sqlWithTime2MetricMap(sql, results, "5m")
			for k, v := range curMetricMap {
				(*m)[k] = append((*m)[k], v...)
			}
		}
	}

	return anyerr
}

// fetchAllCpuMem Fetch all host metrics with different operators (avg_over_time, stddev_over_time, realtime) and
// different resource types (CPU, Memory)
func (s promClient) fetchAllCpuMem(window *metricstype.Window, m *map[string][]metricstype.Metric) error {
	var anyerr error

	for _, metric := range ratioSource {
		for _, method := range []PromMethod{Avg, Std, Latest} {
			promQuery := s.buildCpuMemPromQuery(metric, method, window.Duration)
			promResults, err := s.getPromResults(promQuery)

			if err != nil {
				log.Errorf("error querying Prometheus for query %v: %v\n", promQuery, err)
				anyerr = err
				continue
			}

			curMetricMap := s.nodeCpuMem2MetricMap(promResults, metric, method, window.Duration)

			for k, v := range curMetricMap {
				(*m)[k] = append((*m)[k], v...)
			}
		}
	}

	return anyerr
}

func (s promClient) fetchCapacity(m *map[string][]metricstype.Metric) error {
	log.Info("fetching capacity")
	promNodeCapaicty, err := s.getPromResults(KubeNodeStatusCapacity)
	if err != nil {
		log.Errorf("fetching capacity error")
		return err
	}

	result := s.nodeCapacity2MetricMap(promNodeCapaicty)

	for k, v := range result {
		(*m)[k] = append((*m)[k], v...)
	}

	return nil
}

func (s promClient) buildCpuMemPromQuery(metric PromResource, method PromMethod, rollup string) string {
	var promQuery string

	if method == Latest {
		promQuery = metric
	} else {
		promQuery = fmt.Sprintf("%s(%s[%s])", method, metric, rollup)
	}

	return promQuery
}

func (s promClient) getPromResults(promQuery string) (model.Value, error) {
	v1api := v1.NewAPI(s.client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, warnings, err := v1api.Query(ctx, promQuery, time.Now())
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Warnf("Warnings: %v\n", warnings)
	}
	// log.Debugf("result:\n%v\n", results)

	return results, nil
}

func (s promClient) sqlWithNoTime2MetricMap(sql string, data model.Value) map[string][]metricstype.Metric {
	curMetrics := make(map[string][]metricstype.Metric)

	switch data.(type) {
	case model.Vector:
		for _, result := range data.(model.Vector) {
			var host, label string

			host = string(result.Metric["node"])

			value := float64(result.Value)
			curMetric := metricstype.Metric{Name: sql2NameMap[sql], Type: label, Operator: metricstype.Latest, Rollup: "", Unit: metricstype.Integer, Value: value}

			curMetrics[host] = append(curMetrics[host], curMetric)
		}
	default:
		log.Errorf("error: The capacity results should not be type: %v.\n", data.Type())
	}
	return curMetrics
}

func (s promClient) sqlWithTime2MetricMap(sql string, data model.Value, rollup string) map[string][]metricstype.Metric {
	curMetrics := make(map[string][]metricstype.Metric)

	switch data.(type) {
	case model.Vector:
		for _, result := range data.(model.Vector) {
			host := string(result.Metric["kubernetes_node"])
			value := float64(result.Value)
			unit := ""
			if strings.Contains(sql2NameMap[sql], "bytes") {
				unit = metricstype.Byte
			}
			if sql == NodeNetworkTotalBytesExcludinglo5m {
				value = value / cfg.DiskBandwidthByte * 100
				unit = ""
			}
			if sql == PromSQLNodeDiskSaturation5m {
				value = value * 100
			}
			curMetric := metricstype.Metric{Name: sql2NameMap[sql], Type: "", Operator: metricstype.UnknownOperator, Rollup: rollup, Unit: unit, Value: value}
			curMetrics[host] = append(curMetrics[host], curMetric)
		}
	default:
		log.Errorf("error: The capacity results should not be type: %v.\n", data.Type())
	}
	return curMetrics
}

func (s promClient) nodeCapacity2MetricMap(data model.Value) map[string][]metricstype.Metric {
	curMetrics := make(map[string][]metricstype.Metric)

	operator := metricstype.Capacity

	switch data.(type) {
	case model.Vector:
		for _, result := range data.(model.Vector) {
			host := string(result.Metric["node"])
			resource := strings.ToLower(string(result.Metric["resource"]))
			unit := strings.ToLower(string(result.Metric["unit"]))
			value := float64(result.Value)
			curMetric := metricstype.Metric{Name: KubeNodeStatusCapacity, Type: resource, Operator: operator, Rollup: metricstype.Latest, Unit: unit, Value: value}
			curMetrics[host] = append(curMetrics[host], curMetric)

		}
	default:
		log.Errorf("error: The capacity results should not be type: %v.\n", data.Type())
	}

	for host := range curMetrics {
		// networkcapacity
		networkcapcaity := metricstype.Metric{Name: "kube_node_status_capacity", Type: metricstype.Network, Operator: operator, Rollup: metricstype.Latest, Unit: metricstype.Byte, Value: cfg.DiskBandwidthByte}
		curMetrics[host] = append(curMetrics[host], networkcapcaity)
	}

	return curMetrics
}

func (s promClient) nodeCpuMem2MetricMap(data model.Value, metric PromResource, method PromMethod, rollup string) map[string][]metricstype.Metric {
	var metricType string
	var operator string
	var unit string

	curMetrics := make(map[string][]metricstype.Metric)

	if metric == NodeCpuRatio {
		metricType = metricstype.CPU
	} else {
		metricType = metricstype.Memory
	}

	if method == Avg {
		operator = metricstype.Average
	} else if method == Std {
		operator = metricstype.Std
	} else if method == Latest {
		operator = metricstype.Latest
	} else {
		operator = metricstype.UnknownOperator
	}

	unit = metricstype.Ratio

	switch data.(type) {
	case model.Vector:
		for _, result := range data.(model.Vector) {
			value := float64(result.Value * 100)
			curHost := string(result.Metric["kubernetes_node"])
			curMetric := metricstype.Metric{Name: metric, Type: metricType, Operator: operator, Rollup: rollup, Unit: unit, Value: value}
			curMetrics[curHost] = append(curMetrics[curHost], curMetric)
		}
	default:
		log.Errorf("error: The Prometheus results should not be type: %v.\n", data.Type())
	}

	return curMetrics
}

func (s promClient) Healthy() (int, error) {
	req, err := http.NewRequest("GET", promHealthyUrl, nil)
	if err != nil {
		return -1, err
	}
	resp, _, err := s.client.Do(context.Background(), req)
	if err != nil {
		return -1, err
	}
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("received response status code: %v", resp.StatusCode)
	}
	return 0, nil
}
