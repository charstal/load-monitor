package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/charstal/load-monitor/pkg/metricstype"
	"github.com/francoispqt/gojay"
	"k8s.io/klog/v2"
)

const (
	httpClientTimeoutSecondForRL = 55 * time.Second

	RLServerAddressKey = "RL_SERVER_ADDRESS"

	healthCheckUrl = "/healthy"
	updateUrl      = "/update"
	predictUrl     = "/predict"
)

type RLClient struct {
	httpClient http.Client
	address    string
}

func NewRLClient(address string) (RLClient, error) {
	return RLClient{
		httpClient: http.Client{
			Timeout: httpClientTimeoutSecondForRL,
		},
		address: address,
	}, nil
}

func (c RLClient) Healthy() error {
	req, err := http.NewRequest(http.MethodGet, c.address+healthCheckUrl, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	klog.V(6).Infof("received status code %v from rl server", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		err = fmt.Errorf("received status code %v from rl server", resp.StatusCode)
		klog.Error(err)
		return err
	}
}

func (c RLClient) Predict(request RLClientPredictRequest) (*RLClientPredictResponce, error) {
	requestBody := new(bytes.Buffer)
	json.NewEncoder(requestBody).Encode(request)
	req, err := http.NewRequest(http.MethodPost, c.address+predictUrl, requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	klog.V(6).Infof("received status code %v from rl server", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		body := PredictResponceResult{}
		result := RLClientPredictResponce{Result: body}
		dec := gojay.BorrowDecoder(resp.Body)
		defer dec.Release()
		err = dec.Decode(&result)
		if err != nil {
			klog.Errorf("unable to decode predict responce: %v", err)
			return nil, err
		} else {
			return &result, nil
		}

	} else {
		err = fmt.Errorf("received status code %v from watcher", resp.StatusCode)
		klog.Error(err)
		return nil, err
	}
}

func (c RLClient) Update(request RLClientUpdateRequest) error {
	requestBody := new(bytes.Buffer)
	json.NewEncoder(requestBody).Encode(request)
	req, err := http.NewRequest(http.MethodGet, c.address+updateUrl, requestBody)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	klog.V(6).Infof("received status code %v from rl server", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		err = fmt.Errorf("received status code %v from  rl server", resp.StatusCode)
		klog.Error(err)
		return err
	}
}

func MakePredictRequest(podName, podLabel string, nodes []string) (RLClientPredictRequest, error) {
	return RLClientPredictRequest{
		PodName:  podName,
		PodLabel: podLabel,
		Nodes:    nodes,
	}, nil
}

func MakeUpdateReuqest(podName string, metrics metricstype.WatcherMetrics) (RLClientUpdateRequest, error) {
	return RLClientUpdateRequest{
		PodName:   podName,
		Metrics:   metrics,
		Timestamp: time.Now().Unix(),
	}, nil
}
