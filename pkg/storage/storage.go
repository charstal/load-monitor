package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	cfg "github.com/charstal/load-monitor/pkg/config"
	"github.com/charstal/load-monitor/pkg/metricstype"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Storage struct {
	client influxdb2.Client
}

const (
	InfluxDBUrlKey   = "INFLUXDB_URL"
	InfluxDBTokenKey = "INFLUXDB_TOKEN"
	InfluxDBOrgKey   = "INFLUXDB_ORG"
	InfluxDBName     = "INFLUXDB_DB"
)

var (
	influxUrl   string
	influxToken string
	influxOrg   string
	influxDB    string
)

func NewStorage() (*Storage, error) {
	var ok bool
	influxUrl, ok = os.LookupEnv(InfluxDBUrlKey)
	if !ok {
		influxUrl = cfg.DefaultInfluxURL
	}
	influxToken, ok = os.LookupEnv(InfluxDBTokenKey)
	if !ok {
		influxToken = cfg.DefaultInfluxToken
	}
	influxOrg, ok = os.LookupEnv(InfluxDBOrgKey)
	if !ok {
		influxOrg = cfg.DefaultInfluxOrg
	}
	influxDB, ok = os.LookupEnv(InfluxDBName)
	if !ok {
		influxDB = cfg.DefaultInfluxDB
	}

	client := influxdb2.NewClient(influxUrl, influxToken)

	return &Storage{client: client}, nil
}

func (s *Storage) StoreMetrics(metric *metricstype.WatcherMetrics) error {
	if metric == nil {
		return nil
	}
	data := metric.Data.NodeMetricsMap
	window := metric.Window
	timestamp := window.End
	duration_type := window.Duration

	writeAPI := s.client.WriteAPIBlocking(influxOrg, influxDB)

	for key, values := range data {
		for _, v := range values.Metrics {
			p := influxdb2.NewPointWithMeasurement("node").AddTag("node", key).AddTag("duration", duration_type).SetTime(time.Unix(timestamp, 0))
			name := createName(v.Name, v.Type, v.Operator, v.Unit)
			// fmt.Println(name)
			p.AddField(name, v.Value)
			// p.AddField("operator", v.Operator)
			// p.AddField("type", v.Type)
			// p.AddField("unit", v.Unit)
			// p.AddField()

			writeAPI.WritePoint(context.Background(), p)
		}
	}

	s.client.Close()
	return nil
}

func createName(name, resourceType, operator, unit string) string {
	allName := name

	comactFunc := func(name, s string) string {
		name = name + ":" + s

		return name
	}

	if resourceType != "" {
		allName = comactFunc(allName, resourceType)
	}
	if operator != "" {
		allName = comactFunc(allName, operator)
	}
	if unit != "" {
		allName = comactFunc(allName, unit)
	}

	return allName
}

func (s *Storage) test() {
	// fmt.Println(influxOrg)
	fmt.Printf("%v", s.client)
	writeAPI := s.client.WriteAPIBlocking(influxOrg, "test")
	// Create point using full params constructor
	p := influxdb2.NewPoint("stat",
		map[string]string{"unit": "temperature"},
		map[string]interface{}{"avg": 24.5, "max": 45.0},
		time.Now())
	// write point immediately
	writeAPI.WritePoint(context.Background(), p)
	// Create point using fluent style
	p = influxdb2.NewPointWithMeasurement("stat").
		AddTag("unit", "temperature").
		AddField("avg", 23.2).
		AddField("max", 45.0).
		SetTime(time.Now())
	writeAPI.WritePoint(context.Background(), p)

	// Or write directly line protocol
	line := fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 23.5, 45.0)
	writeAPI.WriteRecord(context.Background(), line)

	// Get query client
	queryAPI := s.client.QueryAPI(influxOrg)
	// Get parser flux query result
	result, err := queryAPI.Query(context.Background(), `from(bucket:"test")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
	if err == nil {
		// Use Next() to iterate over query result lines
		for result.Next() {
			// Observe when there is new grouping key producing new table
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			// read result
			fmt.Printf("row: %s\n", result.Record().String())
		}
		if result.Err() != nil {
			fmt.Printf("Query error: %s\n", result.Err().Error())
		}
	}
	// Ensures background processes finishes
	s.client.Close()
}
