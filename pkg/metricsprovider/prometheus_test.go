package metricsprovider

import (
	"testing"

	"github.com/charstal/load-monitor/pkg/metricstype"
	"github.com/stretchr/testify/assert"
)

var (
	opt = MetricsProviderOpts{
		Name:               PromClientName,
		Address:            "http://10.214.241.226:39090/",
		InsecureSkipVerify: true,
		AuthToken:          "",
	}
)

func TestHealthy(t *testing.T) {
	client, err := NewPromClient(opt)
	assert.Nil(t, err)
	code, err := client.Healthy()
	assert.Nil(t, err)
	assert.Equal(t, 0, code)
}

// avg_over_time(instance:node_cpu:ratio[15m])
// avg_over_time(instance:node_cpu:ratio{instance="k8s-master"}[15m])
func TestBuildPromQuery(t *testing.T) {
	client, err := NewPromClient(opt)
	assert.Nil(t, err)

	sql := client.(promClient).buildCpuMemPromQuery(NodeCpuRatio, Avg, metricstype.CurrentFifteenMinuteWindow().Duration)
	t.Logf(sql)

	sql = client.(promClient).buildCpuMemPromQuery(NodeCpuRatio, Std, metricstype.CurrentFifteenMinuteWindow().Duration)
	t.Logf(sql)
}

// {container="kube-rbac-proxy", endpoint="https", instance="k8s-master", job="node-exporter", namespace="monitoring", pod="node-exporter-7g7fx", service="node-exporter"} => 0.024736842105263043 @[1669188942.163]
// {container="kube-rbac-proxy", endpoint="https", instance="k8s-node01", job="node-exporter", namespace="monitoring", pod="node-exporter-xlrt7", service="node-exporter"} => 0.02419162605588522 @[1669188942.163]
// {container="kube-rbac-proxy", endpoint="https", instance="k8s-node02", job="node-exporter", namespace="monitoring", pod="node-exporter-mhzvp", service="node-exporter"} => 0.024206725146193073 @[1669188942.163]
func TestGetPromResults(t *testing.T) {
	client, err := NewPromClient(opt)
	assert.Nil(t, err)

	sql := client.(promClient).buildCpuMemPromQuery(NodeCpuRatio, Avg, metricstype.CurrentFifteenMinuteWindow().Duration)
	res, err := client.(promClient).getPromResults(sql)
	assert.Nil(t, err)

	t.Log(res)
}

func TestGetCapacity(t *testing.T) {
	client, err := NewPromClient(opt)
	assert.Nil(t, err)

	allMetrics := make(map[string][]metricstype.Metric)
	err = client.(promClient).fetchCapacity(&allMetrics)
	assert.Nil(t, err)

	t.Log(allMetrics)
}

func TestGetAllMetrics(t *testing.T) {
	client, err := NewPromClient(opt)
	assert.Nil(t, err)

	res, err := client.FetchAllHostsMetrics(metricstype.CurrentFifteenMinuteWindow())
	assert.Nil(t, err)

	t.Log(res)

}
