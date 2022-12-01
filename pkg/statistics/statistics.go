package statistics

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cfg "github.com/charstal/load-monitor/pkg/config"
	"github.com/charstal/load-monitor/pkg/metricstype"
	"github.com/charstal/load-monitor/pkg/util"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type OfflineReader struct {
	// etcd client
	client         clientv3.Client
	basePath       string
	remoteFilePath string
	remoteFileMD5  string
	localFilePath  string
	localFileMD5   string
	tmpFilePath    string
	tmpFileMD5     string
	statisData     *map[string][]metricstype.Metric
}

const (
	// podInfoUrl       = "/metric/pod_info_path"
	// podInfoMD5Url    = "/metric/pod_info_md5"
	statisticsUrl    = "/metric/statistics_path"
	statisticsMD5Url = "/metric/statistics_path_md5"
	EtcdUrlKey       = "ETCD_URL"
	EtcdUsernameKey  = "ETCD_USERNAME"
	EtcdPasswdKey    = "ETCD_PASSWD"
	BaseDirKey       = "NFS_BASE_DIR"

	tmpFilePrefix = "tmp-"
)

var (
	etcdUrl      string
	etcdUsername string
	etcdPasswd   string
	baseDir      string
)

func NewOfflineReader() (*OfflineReader, error) {
	var ok bool
	etcdUrl, ok = os.LookupEnv(EtcdUrlKey)
	if !ok {
		etcdUrl = cfg.DefaultETCDURL
	}
	etcdUsername, ok = os.LookupEnv(EtcdUsernameKey)
	if !ok {
		etcdUsername = cfg.DefaultETCDUsername
	}
	etcdPasswd, ok = os.LookupEnv(EtcdPasswdKey)
	if !ok {
		etcdPasswd = cfg.DefaultETCDPasswd
	}
	baseDir, ok = os.LookupEnv(BaseDirKey)
	if !ok {
		baseDir = cfg.DefaultSourceBaseDir
	}

	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdUrl},
		DialTimeout: 5 * time.Second,
		Username:    etcdUsername,
		Password:    etcdPasswd,
	})

	if err != nil {
		return nil, err
	}

	// fmt.Printf("%s", fileMd5)

	offlineReader := OfflineReader{
		client:   *etcdClient,
		basePath: baseDir,
	}

	// offlineReader.Update()

	return &offlineReader, nil
}

func (or *OfflineReader) Update() error {
	err := or.pullFromEtcd()
	if err != nil {
		return err
	}
	err = or.fetchStatisticsFile()
	if err != nil {
		return err
	}
	err = or.transferTmpFile2LocalFile()
	if err != nil {
		return err
	}
	err = or.readFromCsv()
	if err != nil {
		return err
	}

	return nil
}

func (or *OfflineReader) GetMetrics() *map[string][]metricstype.Metric {
	return or.statisData
}

func (or *OfflineReader) pullFromEtcd() error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := or.client.Get(ctx, statisticsUrl)
	filePath := string(res.Kvs[0].Value)
	if err != nil {
		return err
	}

	filePath = filepath.Join(baseDir, filePath)
	// fmt.Printf("%s", filePath)
	res, err = or.client.Get(ctx, statisticsMD5Url)
	fileMd5 := string(res.Kvs[0].Value)
	if err != nil {
		return err
	}

	or.remoteFileMD5 = fileMd5
	or.remoteFilePath = filePath

	return nil
}

func (or *OfflineReader) fetchStatisticsFile() error {
	sourceFile := or.remoteFilePath
	if len(sourceFile) == 0 {
		return errors.New("sourceFile empty")
	}
	fileName := filepath.Base(sourceFile)

	desFileName := tmpFilePrefix + fileName
	desPath := desFileName
	err := util.CopyFile(sourceFile, desPath)
	if err != nil {
		return err
	}

	or.tmpFilePath = desFileName
	or.tmpFileMD5 = or.remoteFileMD5

	return nil
}

func (or *OfflineReader) checkFileMd5(filePath string, fileMD5 string) (bool, error) {
	if len(filePath) == 0 {
		return false, errors.New("filePath empty")
	}
	calMd5, err := util.GetFileMd5(filePath)
	if err != nil {
		return false, err
	}
	if fileMD5 != calMd5 {
		return false, nil
	}
	return true, nil
}

func (or *OfflineReader) transferTmpFile2LocalFile() error {
	var err error = nil
	filePath := or.tmpFilePath
	fileMd5 := or.tmpFileMD5
	if len(filePath) == 0 {
		return nil
	}
	// same
	if or.localFileMD5 == fileMd5 {
		return nil
	}

	res, err := or.checkFileMd5(filePath, fileMd5)

	if err != nil {
		return err
	}

	if res {
		newPath := strings.TrimPrefix(filePath, tmpFilePrefix)
		util.RenameFile(filePath, newPath)
		or.localFileMD5 = fileMd5
		or.localFilePath = newPath
	} else {
		err = errors.New("file and md5 no matcher")
	}

	or.tmpFileMD5 = ""
	or.tmpFilePath = ""

	return err
}

func (or *OfflineReader) readFromCsv() error {
	filePath := or.localFilePath

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	// skip first row
	row1, err := bufio.NewReader(file).ReadSlice('\n')
	if err != nil {
		return err
	}
	_, err = file.Seek(int64(len(row1)), io.SeekStart)
	if err != nil {
		return err
	}

	reader := csv.NewReader(file)
	// reader.FieldsPerRecord = -1

	record, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	curMetrics := make(map[string][]metricstype.Metric)
	for _, item := range record {
		arr := make([]metricstype.Metric, 0)
		label := item[0]
		if len(label) == 0 {
			continue
		}
		// fmt.Println(label)
		cpuStd, _ := strconv.ParseFloat(item[1], 64)
		arr = append(arr, metricstype.Metric{Name: "statistic", Type: metricstype.CPU, Operator: metricstype.Std, Rollup: "", Unit: metricstype.M, Value: cpuStd})
		cpuAvg, _ := strconv.ParseFloat(item[2], 64)
		arr = append(arr, metricstype.Metric{Name: "statistic", Type: metricstype.CPU, Operator: metricstype.Average, Rollup: "", Unit: metricstype.M, Value: cpuAvg})
		memStd, _ := strconv.ParseFloat(item[3], 64)
		arr = append(arr, metricstype.Metric{Name: "statistic", Type: metricstype.Memory, Operator: metricstype.Std, Rollup: "", Unit: metricstype.MiB, Value: memStd})
		memAvg, _ := strconv.ParseFloat(item[4], 64)
		arr = append(arr, metricstype.Metric{Name: "statistic", Type: metricstype.Memory, Operator: metricstype.Average, Rollup: "", Unit: metricstype.MiB, Value: memAvg})
		curMetrics[label] = append(curMetrics[label], arr...)
	}

	or.statisData = &curMetrics
	return nil
}
