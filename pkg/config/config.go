package config

import (
	"os"
	"strconv"
)

const (
	// 	// iperf3 ex. 79.5 Mbits/sec
	// 	DiskBandwidthMB    float64 = 79.5 / 8
	// 	DiskBandwidthKB    float64 = DiskBandwidthMB * 1024
	// 	DiskBandwidthBytes float64 = DiskBandwidthKB * 1024
	DefaultBandWidthMB = 79.5 / 8

	DiskBytesKey = "DISK_BAND_WIDTH_Bytes"
	DiskKBKey    = "DISK_BAND_WIDTH_KB"
	DiskMBKey    = "DISK_BAND_WIDTH_MB"
	DiskGBKey    = "DISK_BAND_WIDTH_GB"
)

var (
	DiskBandwidthGB   float64
	DiskBandwidthMB   float64
	DiskBandwidthKB   float64
	DiskBandwidthByte float64
)

func init() {
	var ok bool
	var s string
	s, ok = os.LookupEnv(DiskBytesKey)
	if ok {
		DiskBandwidthBytes, err := strconv.ParseFloat(s, 64)
		if err != nil {
			{
				panic(err)
			}
		}
		DiskBandwidthKB = DiskBandwidthBytes * 1024
		DiskBandwidthMB = DiskBandwidthKB * 1024
		DiskBandwidthGB = DiskBandwidthMB * 1024
		return
	}

	s, ok = os.LookupEnv(DiskKBKey)
	if ok {
		DiskBandwidthKB, err := strconv.ParseFloat(s, 64)
		if err != nil {
			{
				panic(err)
			}
		}
		DiskBandwidthByte = DiskBandwidthKB / 1024
		DiskBandwidthMB = DiskBandwidthKB * 1024
		DiskBandwidthGB = DiskBandwidthMB * 1024
		return
	}

	s, ok = os.LookupEnv(DiskMBKey)
	if ok {
		DiskBandwidthMB, err := strconv.ParseFloat(s, 64)
		if err != nil {
			{
				panic(err)
			}
		}
		DiskBandwidthKB = DiskBandwidthMB / 1024
		DiskBandwidthByte = DiskBandwidthKB / 1024
		DiskBandwidthGB = DiskBandwidthMB * 1024
		return
	}
	s, ok = os.LookupEnv(DiskGBKey)
	if ok {
		DiskBandwidthGB, err := strconv.ParseFloat(s, 64)
		if err != nil {
			{
				panic(err)
			}
		}
		DiskBandwidthMB = DiskBandwidthGB / 1024
		DiskBandwidthKB = DiskBandwidthMB / 1024
		DiskBandwidthByte = DiskBandwidthKB / 1024
		return
	}

	// default
	DiskBandwidthMB = DefaultBandWidthMB
	DiskBandwidthGB = DefaultBandWidthMB / 1024
	DiskBandwidthKB = DefaultBandWidthMB * 1024
	DiskBandwidthByte = DefaultBandWidthMB * 1024 * 1024
}

const (
	// env variable that provides path to kube config file, if deploying from outside K8s cluster

	DefaultKubeConfig     = "~/.kube/config"
	DefaultPromAddress    = "http://prometheus-server.svc.cluster.local:80"
	DefaultPromhealthyURL = "-/healthy"

	// DefaultETCDURL      = "etcd-dev.default.svc.cluster.local:2379"
	// DefaultETCDUsername = "root"
	// DefaultETCDPasswd   = "CrHkL98Ryr"

	// DefaultRemoteBaseDir = "offline"
	// DefaultLocalBaseDir  = "statistics"
	DefaultMongoURL                 = "mongodb://192.168.30.154:27017"
	DefaultMongoDatabase            = "load_monitor"
	DefaultMongoStatisticCollection = "statistics"

	DefaultInfluxURL   = "http://192.168.30.154:8086"
	DefaultInfluxToken = "K4cPAKvAFgmu0Kdfo0LZH6qS3jq1umX0zTOyV_oVyWM6p_nGq_R55MIE8RXDEXiopTJKpj2OiPJMRKVbx1P0Dg=="
	DefaultInfluxOrg   = "mo"
	DefaultInfluxDB    = "k8s-info"

	DefaultRLServerAddress = "http://localhost:8000"
)
