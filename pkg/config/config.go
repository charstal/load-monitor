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
	DefaultPromAddress    = "http://prometheus-k8s:9090"
	DefaultPromhealthyURL = "-/healthy"

	DefaultETCDURL      = "10.214.241.226:32379"
	DefaultETCDUsername = "root"
	DefaultETCDPasswd   = "Y4b5EAwMlQ"

	DefaultRemoteBaseDir = "../offline"
	DefaultLocalBaseDir  = "statistics"

	DefaultInfluxURL   = "http://10.72.100.7:8086/"
	DefaultInfluxToken = "lVnogqXxfgpSOeQ7EledUAF6_luxl5NtvVv1YJg5CaTkgcUB-2dvPK_Royv1E_g37y81HhjehudrCR64Ri4nBw=="
	DefaultInfluxOrg   = "mo"
)
