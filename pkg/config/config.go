package config

const (
	// iperf3 ex. 79.5 Mbits/sec
	DiskBandwidthMB    float64 = 79.5 / 8
	DiskBandwidthKB    float64 = DiskBandwidthMB * 1024
	DiskBandwidthBytes float64 = DiskBandwidthKB * 1024
)

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
