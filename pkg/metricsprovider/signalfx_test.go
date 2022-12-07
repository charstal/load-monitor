package metricsprovider

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/charstal/load-monitor/pkg/metricstype"
	"github.com/stretchr/testify/assert"
)

func TestNewSignalFxClient(t *testing.T) {
	opts := MetricsProviderOpts{
		Name:      SignalFxClientName,
		Address:   "",
		AuthToken: "Test",
	}
	_, err := NewSignalFxClient(opts)
	assert.Nil(t, err)

	opts.Name = "invalid"
	_, err = NewSignalFxClient(opts)
	assert.NotNil(t, err)
}

func TestFetchAllHostMetrics(t *testing.T) {
	metricData := `{
  "data": {
    "Ehql_bxBgAc": [
      [
        1600213380000,
        84.64246793530153
      ]
    ],
    "EuXgJm7BkAA": [
	  [
		1614634260000,
		5.450946379084264
     ]
    ]
  },
  "errors": []
}`
	metaData := `{
   "count":2,
   "partialCount":false,
   "results":[
      {
         "active":true,
         "created":1614534848000,
         "creator":null,
         "dimensions":{
            "host":"test1.dev.com",
            "sf_metric":null
         },
         "id":"Ehql_bxBgAc",
         "lastUpdated":0,
         "lastUpdatedBy":null,
         "metric":"cpu.utilization"
      },
      {
         "active":true,
         "created":1614534848000,
         "creator":null,
         "dimensions":{
            "host":"test2.dev.com",
            "sf_metric":null
         },
         "id":"EuXgJm7BkAA",
         "lastUpdated":0,
         "lastUpdatedBy":null,
         "metric":"cpu.utilization"
      }
   ]
}`
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.Path, signalFxMetdataAPI) {
			resp.Write([]byte(metaData))
		} else {
			resp.Write([]byte(metricData))
		}
	}))
	opts := MetricsProviderOpts{
		Name:      SignalFxClientName,
		Address:   server.URL,
		AuthToken: "PWNED",
	}

	client, err := NewSignalFxClient(opts)
	assert.Nil(t, err)

	curTime := time.Now().Unix()
	windows := &metricstype.Window{Duration: metricstype.FifteenMinutes, Start: curTime - 15*60, End: curTime}

	metrics, err := client.FetchAllHostsMetrics(windows)
	assert.Nil(t, err)
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics["test1"])
	assert.NotNil(t, metrics["test2"])

	defer server.Close()
}
