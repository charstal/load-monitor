package rlsupporter

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charstal/load-monitor/pkg/api"
	"github.com/charstal/load-monitor/pkg/config"
	v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	metricsUpdateIntervalSeconds = 30
	heartbeatIntervalSeconds     = 10
	heartbeatTimeoutSeconds      = 2 * metricsUpdateIntervalSeconds

	KubeConfig = "KUBE_CONFIG"
)

type RLSupporter struct {
	rlClient          api.RLClient
	k8sClient         *kubernetes.Clientset
	lastHeartBeatTime int64
	mu                sync.RWMutex
}

var (
	rlServerAddress   string
	kubeConfigPresent = false
	kubeConfigPath    string
)

func init() {
	var ok bool
	kubeConfigPath, ok = os.LookupEnv(KubeConfig)
	if ok {
		kubeConfigPresent = true
	}
}

func NewRLSupporter() (*RLSupporter, error) {
	var ok bool
	var k8sConfig *rest.Config
	rlServerAddress, ok = os.LookupEnv(api.RLServerAddressKey)
	if !ok {
		rlServerAddress = config.DefaultRLServerAddress
	}
	klog.InfoS("rl server address", rlServerAddress)
	client, err := api.NewRLClient(rlServerAddress)
	if err != nil {
		return nil, err
	}

	kubeConfig := ""
	if kubeConfigPresent {
		kubeConfig = kubeConfigPath
	}
	k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	sharedInformersFactory := informers.NewSharedInformerFactory(k8sClient, time.Minute)

	podinformer := sharedInformersFactory.Core().V1().Pods().Informer()
	// TODO need to update
	podinformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		//pod资源对象创建的时候出发的回调方法
		AddFunc: func(obj interface{}) {
			obja := obj.(v1.Pod)
			fmt.Println(obja)
		},
		//更新回调
		UpdateFunc: func(oldObj, newObj interface{}) {

		},
		//删除回调
		DeleteFunc: func(obj interface{}) {

		},
	})
	//这里

	rlClient := &RLSupporter{
		rlClient: client,
	}

	rlClient.heartbeat()
	go func() {
		ticker := time.NewTicker(time.Second * heartbeatIntervalSeconds)
		for range ticker.C {
			err = rlClient.heartbeat()
			if err != nil {
				klog.ErrorS(err, "Unable to heartbeat load monitor")
			}
		}
	}()

	return rlClient, nil
}

func (c *RLSupporter) Valid() bool {
	return c.heartbeatCheck()
}

func (c *RLSupporter) heartbeatCheck() bool {
	now := time.Now().Unix()
	return now-c.lastHeartBeatTime <= heartbeatTimeoutSeconds
}

func (c *RLSupporter) heartbeat() error {
	err := c.rlClient.Healthy()
	if err != nil {
		klog.Error(err, "fail: cannot get hearbeat rl server")
		return err
	}

	c.mu.Lock()
	c.lastHeartBeatTime = time.Now().Unix()
	c.mu.Unlock()

	return nil
}
