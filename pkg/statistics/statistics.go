package statistics

import (
	"context"
	"os"
	"time"

	cfg "github.com/charstal/load-monitor/pkg/config"
	"github.com/charstal/load-monitor/pkg/metricstype"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type statisticsData struct {
	CpuStd float64 `bson:"cpu_std/m"`
	CpuAvg float64 `bson:"cpu_avg/m"`
	MemStd float64 `bson:"mem_std/MiB"`
	MemAvg float64 `bson:"mem_avg/MiB"`
}

type Statistics struct {
	ID   primitive.ObjectID        `bson:"_id"`
	Time time.Time                 `bson:"time"`
	Data map[string]statisticsData `bson:"data"`
}

type OfflineReader struct {
	// mongodb client
	client              *mongo.Client
	sourceStatistics    *Statistics
	timeReceived        time.Time
	generatedStatistics *map[string][]metricstype.Metric
	// mtx
}

const (
	// podInfoUrl       = "/metric/pod_info_path"
	// podInfoMD5Url    = "/metric/pod_info_md5"
	// statisticsUrlKey    = "/metric/statistics_path"
	// statisticsMD5UrlKey = "/metric/statistics_path_md5"
	// EtcdUrlKey          = "ETCD_URL"
	// EtcdUsernameKey     = "ETCD_USERNAME"
	// EtcdPasswdKey       = "ETCD_PASSWD"
	// RemoteBaseDirKey    = "REMOTE_BASE_DIR"
	// LocalBaseDirKey     = "LOCAL_BASE_DIR"

	// tmpFilePrefix = "tmp-"
	MongoDBUrlKey        = "MONGODB_URI"
	MongodbDatabaseKey   = "MONGODB_DATABASE"
	MongodbCollectionKey = "MONGODB_COLLETION"
)

var (
	// etcdUrl       string
	// etcdUsername  string
	// etcdPasswd    string
	// remoteBaseDir string
	// localBaseDir  string
	mongodbUrl                 string
	mongodbDatabase            string
	mongodbStatisticCollection string
)

func NewOfflineReader() (*OfflineReader, error) {
	var ok bool
	mongodbUrl, ok = os.LookupEnv(MongoDBUrlKey)
	if !ok {
		mongodbUrl = cfg.DefaultMongoURL
	}
	mongodbDatabase, ok = os.LookupEnv(MongodbDatabaseKey)
	if !ok {
		mongodbDatabase = cfg.DefaultMongoDatabase
	}
	mongodbStatisticCollection, ok = os.LookupEnv(MongodbCollectionKey)
	if !ok {
		mongodbStatisticCollection = cfg.DefaultMongoStatisticCollection
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongodbUrl))
	if err != nil {
		panic(err)
	}

	// var ok bool
	// etcdUrl, ok = os.LookupEnv(EtcdUrlKey)
	// if !ok {
	// 	etcdUrl = cfg.DefaultETCDURL
	// }
	// etcdUsername, ok = os.LookupEnv(EtcdUsernameKey)
	// if !ok {
	// 	etcdUsername = cfg.DefaultETCDUsername
	// }
	// etcdPasswd, ok = os.LookupEnv(EtcdPasswdKey)
	// if !ok {
	// 	etcdPasswd = cfg.DefaultETCDPasswd
	// }
	// remoteBaseDir, ok = os.LookupEnv(RemoteBaseDirKey)
	// if !ok {
	// 	remoteBaseDir = cfg.DefaultRemoteBaseDir
	// }
	// localBaseDir, ok = os.LookupEnv(LocalBaseDirKey)
	// if !ok {
	// 	localBaseDir = cfg.DefaultLocalBaseDir
	// }

	// _, err := os.Stat(localBaseDir)
	// if os.IsNotExist(err) {
	// 	err := os.Mkdir(localBaseDir, 0777)
	// 	if err != nil {
	// 		log.Error("cannot create %v", err)
	// 	}

	// }

	// if err != nil {
	// 	return nil, err
	// }

	// fmt.Printf("%s", fileMd5)

	offlineReader := OfflineReader{
		client: client,
	}

	// offlineReader.Update()

	return &offlineReader, nil
}

func (or *OfflineReader) Close() error {
	if err := or.client.Disconnect(context.TODO()); err != nil {
		return err
	}
	return nil
}

func (or *OfflineReader) Update() error {
	// err := or.pullFromEtcd()
	// if err != nil {
	// 	return err
	// }
	// err = or.fetchStatisticsFile()
	// if err != nil {
	// 	return err
	// }
	// err = or.transferTmpFile2LocalFile()
	// if err != nil {
	// 	return err
	// }
	// err = or.readFromCsv()
	// if err != nil {
	// 	return err
	// }
	or.getFromMongo()
	or.generateStatistics()

	return nil
}

func (or *OfflineReader) generateStatistics() {
	// pull latest time
	src := or.sourceStatistics

	if or.timeReceived.After(src.Time) {
		return
	}

	curMetrics := make(map[string][]metricstype.Metric)
	for label, item := range src.Data {
		arr := make([]metricstype.Metric, 0)
		if len(label) == 0 {
			continue
		}

		metricsName := "statistic"
		// fmt.Println(label)
		cpuStd := item.CpuStd
		arr = append(arr, metricstype.Metric{Name: metricsName, Type: metricstype.CPU, Operator: metricstype.Std, Rollup: "", Unit: metricstype.M, Value: cpuStd})
		cpuAvg := item.CpuAvg
		arr = append(arr, metricstype.Metric{Name: metricsName, Type: metricstype.CPU, Operator: metricstype.Average, Rollup: "", Unit: metricstype.M, Value: cpuAvg})
		memStd := item.MemStd
		arr = append(arr, metricstype.Metric{Name: metricsName, Type: metricstype.Memory, Operator: metricstype.Std, Rollup: "", Unit: metricstype.MiB, Value: memStd})
		memAvg := item.MemAvg
		arr = append(arr, metricstype.Metric{Name: metricsName, Type: metricstype.Memory, Operator: metricstype.Average, Rollup: "", Unit: metricstype.MiB, Value: memAvg})
		curMetrics[label] = append(curMetrics[label], arr...)
	}

	or.timeReceived = src.Time
	or.generatedStatistics = &curMetrics
}

func (or *OfflineReader) getFromMongo() error {
	collection := or.client.Database(mongodbDatabase).Collection(mongodbStatisticCollection)
	// var result interface{}
	// s := map[string]int{
	// 	"_id": -1,
	// }
	var opts = options.FindOne()
	opts.SetSort(bson.D{{Key: "time", Value: -1}})
	var res Statistics
	err := collection.FindOne(context.TODO(), bson.D{{}}, opts).Decode(&res)
	if err != nil {
		return err
	}
	log.Infof("read from mongodb %v", res)
	or.sourceStatistics = &res
	return nil
}

func (or *OfflineReader) GetMetrics() *map[string][]metricstype.Metric {
	return or.generatedStatistics
}

// func (or *OfflineReader) pullFromEtcd() error {
// 	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
// 	res, err := or.client.Get(ctx, statisticsUrlKey)
// 	if len(res.Kvs) == 0 {
// 		log.Debug("statistics url empty")
// 		return nil
// 	}
// 	filePath := string(res.Kvs[0].Value)
// 	if err != nil {
// 		return err
// 	}

// 	filePath = filepath.Join(or.remoteBasePath, filePath)
// 	// fmt.Printf("%s", filePath)
// 	res, err = or.client.Get(ctx, statisticsMD5UrlKey)
// 	if len(res.Kvs) == 0 {
// 		log.Debug("statistics md5 empty")
// 		return nil
// 	}
// 	fileMd5 := string(res.Kvs[0].Value)
// 	if err != nil {
// 		return err
// 	}

// 	or.remoteFileMD5 = fileMd5
// 	or.remoteFilePath = filePath

// 	return nil
// }

// func (or *OfflineReader) fetchStatisticsFile() error {
// 	sourceFile := or.remoteFilePath
// 	if len(sourceFile) == 0 {
// 		log.Debug("fetchStatisticsFile: sourceFile empty")
// 		return nil
// 	}
// 	fileName := filepath.Base(sourceFile)

// 	desFileName := tmpFilePrefix + fileName
// 	desPath := filepath.Join(localBaseDir, desFileName)
// 	err := util.CopyFile(sourceFile, desPath)
// 	if err != nil {
// 		return err
// 	}

// 	or.tmpFilePath = desPath
// 	or.tmpFileMD5 = or.remoteFileMD5

// 	return nil
// }

// func (or *OfflineReader) checkFileMd5(filePath string, fileMD5 string) (bool, error) {
// 	if len(filePath) == 0 {
// 		return false, errors.New("filePath empty")
// 	}
// 	calMd5, err := util.GetFileMd5(filePath)
// 	if err != nil {
// 		return false, err
// 	}
// 	if fileMD5 != calMd5 {
// 		return false, nil
// 	}
// 	return true, nil
// }

// func (or *OfflineReader) transferTmpFile2LocalFile() error {
// 	var err error = nil
// 	filePath := or.tmpFilePath
// 	fileMd5 := or.tmpFileMD5
// 	if len(filePath) == 0 || len(fileMd5) == 0 {
// 		log.Debug("transferTmpFile2LocalFile: sourceFile empty")
// 		return nil
// 	}
// 	// same
// 	if or.localFileMD5 == fileMd5 {
// 		return nil
// 	}

// 	res, err := or.checkFileMd5(filePath, fileMd5)

// 	if err != nil {
// 		return err
// 	}

// 	if res {
// 		fileName := filepath.Base(filePath)
// 		newFileName := strings.TrimPrefix(fileName, tmpFilePrefix)
// 		newPath := filepath.Join(or.localBasePath, newFileName)
// 		util.RenameFile(filePath, newPath)
// 		or.localFileMD5 = fileMd5
// 		or.localFilePath = newPath
// 	} else {
// 		err = errors.New("file and md5 no matcher")
// 	}

// 	or.tmpFileMD5 = ""
// 	or.tmpFilePath = ""

// 	return err
// }

// func (or *OfflineReader) readFromCsv() error {
// 	filePath := or.localFilePath

// 	if len(filePath) == 0 {
// 		log.Debug("readFromCsv: sourceFile empty")
// 		return nil
// 	}

// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return err
// 	}
// 	// skip first row
// 	row1, err := bufio.NewReader(file).ReadSlice('\n')
// 	if err != nil {
// 		return err
// 	}
// 	_, err = file.Seek(int64(len(row1)), io.SeekStart)
// 	if err != nil {
// 		return err
// 	}

// 	reader := csv.NewReader(file)
// 	// reader.FieldsPerRecord = -1

// 	record, err := reader.ReadAll()
// 	if err != nil {
// 		panic(err)
// 	}

// 	curMetrics := make(map[string][]metricstype.Metric)
// 	for _, item := range record {
// 		arr := make([]metricstype.Metric, 0)
// 		label := item[0]
// 		if len(label) == 0 {
// 			continue
// 		}
// 		// fmt.Println(label)
// 		cpuStd, _ := strconv.ParseFloat(item[1], 64)
// 		arr = append(arr, metricstype.Metric{Name: "statistic", Type: metricstype.CPU, Operator: metricstype.Std, Rollup: "", Unit: metricstype.M, Value: cpuStd})
// 		cpuAvg, _ := strconv.ParseFloat(item[2], 64)
// 		arr = append(arr, metricstype.Metric{Name: "statistic", Type: metricstype.CPU, Operator: metricstype.Average, Rollup: "", Unit: metricstype.M, Value: cpuAvg})
// 		memStd, _ := strconv.ParseFloat(item[3], 64)
// 		arr = append(arr, metricstype.Metric{Name: "statistic", Type: metricstype.Memory, Operator: metricstype.Std, Rollup: "", Unit: metricstype.MiB, Value: memStd})
// 		memAvg, _ := strconv.ParseFloat(item[4], 64)
// 		arr = append(arr, metricstype.Metric{Name: "statistic", Type: metricstype.Memory, Operator: metricstype.Average, Rollup: "", Unit: metricstype.MiB, Value: memAvg})
// 		curMetrics[label] = append(curMetrics[label], arr...)
// 	}

// 	or.statisData = &curMetrics
// 	return nil
// }
