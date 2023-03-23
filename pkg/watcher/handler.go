package watcher

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charstal/load-monitor/pkg/metricstype"
	"github.com/francoispqt/gojay"
	log "github.com/sirupsen/logrus"
)

const (
	BaseUrl        = "/watcher"
	HealthCheckUrl = "/healthy"
	MertricUrl     = "/metric"
	JobUrl         = "/job"
	ScheduleUrl    = "/schedule"
)

// GetLatestWatcherMetrics It starts from 15 minute window, and falls back to 10 min, 5 min windows subsequently
// if metrics are not present. StartWatching() should be called before calling this.
func (w *Watcher) GetLatestWatcherMetrics(duration string) (*metricstype.WatcherMetrics, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	if !w.isStarted {
		return nil, errors.New("need to call StartWatching() first")
	}

	switch {
	case duration == metricstype.FifteenMinutes && len(w.fifteenMinute) > 0:
		return w.deepCopyWatcherMetrics(&w.fifteenMinute[len(w.fifteenMinute)-1]), nil
	case (duration == metricstype.FifteenMinutes || duration == metricstype.TenMinutes) && len(w.tenMinute) > 0:
		return w.deepCopyWatcherMetrics(&w.tenMinute[len(w.tenMinute)-1]), nil
	case (duration == metricstype.TenMinutes || duration == metricstype.FiveMinutes) && len(w.fiveMinute) > 0:
		return w.deepCopyWatcherMetrics(&w.fiveMinute[len(w.fiveMinute)-1]), nil
	default:
		return nil, errors.New("unable to get any latest metrics")
	}
}

func (w *Watcher) Healthy() error {
	_, err := w.client.Healthy()
	return err
}

func (w *Watcher) startHttpServer(shutdown chan struct{}) {
	http.HandleFunc(BaseUrl, w.handler)
	http.HandleFunc(HealthCheckUrl, w.healthCheckHandler)
	http.HandleFunc(JobUrl, w.jobFinishedHandler)
	http.HandleFunc(MertricUrl, w.metricHandler)
	http.HandleFunc(ScheduleUrl, w.scheduleHandler)

	server := &http.Server{
		Addr:    ":2020",
		Handler: http.DefaultServeMux,
	}

	go func() {
		log.Warn(server.ListenAndServe())
	}()

	signal.Notify(w.shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-w.shutdown
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Errorf("Unable to shutdown server: %v", err)
		}
		shutdown <- struct{}{}
	}()

}

// HTTP Handler for BaseUrl endpoint
func (w *Watcher) handler(resp http.ResponseWriter, r *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	metrics, err := w.GetLatestWatcherMetrics(metricstype.FifteenMinutes)
	if metrics == nil {
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			log.Error(err)
			return
		}
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	host := r.URL.Query().Get("host")
	var bytes []byte
	if host != "" {
		if _, ok := metrics.Data.NodeMetricsMap[host]; ok {
			hostMetricsData := make(map[string]metricstype.NodeMetrics)
			hostMetricsData[host] = metrics.Data.NodeMetricsMap[host]
			hostMetrics := metricstype.WatcherMetrics{Timestamp: metrics.Timestamp,
				Window: metrics.Window,
				Source: metrics.Source,
				Data:   metricstype.Data{NodeMetricsMap: hostMetricsData},
			}
			bytes, err = gojay.MarshalJSONObject(&hostMetrics)
		} else {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		bytes, err = gojay.MarshalJSONObject(metrics)
	}

	if err != nil {
		log.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = resp.Write(bytes)
	if err != nil {
		log.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
	}
}

// Simple server status handler
func (w *Watcher) metricHandler(resp http.ResponseWriter, r *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	duration := r.URL.Query().Get("duration")

	metrics, err := w.GetLatestWatcherMetrics(duration)
	if metrics == nil {
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			log.Error(err)
			return
		}
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	bytes, err := gojay.MarshalJSONObject(metrics)

	if err != nil {
		log.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = resp.Write(bytes)
	if err != nil {
		log.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
	}
}

type JobRequest struct {
	FilePath string `json:"file_path"`
	MD5      string `json:"md5"`
}

// Simple server status handler
func (w *Watcher) jobFinishedHandler(resp http.ResponseWriter, r *http.Request) {
	// body, _ := io.ReadAll(r.Body)
	// var req JobRequest

	// if err := json.Unmarshal(body, &req); err == nil {
	// 	// fmt.Printf("%v", req)
	// 	resp.Write([]byte("Please add filepath and md5"))
	// }
	if err := w.statisticsReader.Update(); err != nil {
		log.Warnf("job fail: %v", err)
		resp.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	log.Info("job finshed")
	resp.WriteHeader(http.StatusOK)
}

func (w *Watcher) scheduleHandler(resp http.ResponseWriter, r *http.Request) {
	resp.Write([]byte("unimplement"))
	resp.WriteHeader(http.StatusOK)
}

// Simple server status handler
func (w *Watcher) healthCheckHandler(resp http.ResponseWriter, r *http.Request) {
	if status, err := w.client.Healthy(); status != 0 {
		log.Warnf("health check failed with: %v", err)
		resp.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	resp.WriteHeader(http.StatusOK)
}
