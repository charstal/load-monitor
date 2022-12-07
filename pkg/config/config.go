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
	DiskBandwidthGB    float64
	DiskBandwidthMB    float64
	DiskBandwidthKB    float64
	DiskBandwidthBytes float64
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
		DiskBandwidthBytes = DiskBandwidthKB / 1024
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
		DiskBandwidthBytes = DiskBandwidthKB / 1024
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
		DiskBandwidthBytes = DiskBandwidthKB / 1024
		return
	}

	// default
	DiskBandwidthMB = DefaultBandWidthMB
	DiskBandwidthGB = DefaultBandWidthMB / 1024
	DiskBandwidthKB = DefaultBandWidthMB * 1024
	DiskBandwidthBytes = DefaultBandWidthMB * 1024 * 1024
}

const (
	// env variable that provides path to kube config file, if deploying from outside K8s cluster

	DefaultKubeConfig     = "~/.kube/config"
	DefaultPromAddress    = "http://prometheus-k8s.monitoring.svc.cluster.local:9090"
	DefaultPromhealthyURL = "-/healthy"

	// DefaultETCDURL      = "etcd-dev.default.svc.cluster.local:2379"
	// DefaultETCDUsername = "root"
	// DefaultETCDPasswd   = "CrHkL98Ryr"

	// DefaultRemoteBaseDir = "offline"
	// DefaultLocalBaseDir  = "statistics"
	DefaultMongoURL                 = "mongodb://mo:momodel@192.168.122.67:27017/?authSource=mo"
	DefaultMongoDatabase            = "mo"
	DefaultMongoStatisticCollection = "statistics"

	DefaultInfluxURL   = "http://192.168.122.67:8086"
	DefaultInfluxToken = "YEkyUh-YUJ6pfV3Tf996_uHQan_szSihhTxdgxGf9HDMxQ2AXin5UqXN7EoKeHDaM9p12yKOngeD-OrbKf0zTA=="
	DefaultInfluxOrg   = "mo"
)
