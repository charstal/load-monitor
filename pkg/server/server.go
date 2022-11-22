package server

import (
	"github.com/charstal/load-monitor/pkg/metricsprovider"
	"github.com/charstal/load-monitor/pkg/watcher"
)

// used for server
type Server struct {
	dataSourceClient metricsprovider.MetricsProviderClient
	watcher          *watcher.Watcher
	shutdown         chan struct{}
}

func NewServer(ch chan struct{}) (*Server, error) {
	server := Server{}
	sourceClient, err := metricsprovider.NewMetricsProvider()
	if err != nil {
		return nil, err
	}

	server.watcher = watcher.NewWatcher(sourceClient)
	server.dataSourceClient = sourceClient
	server.shutdown = ch

	return &server, nil
}

func (s *Server) Run() {
	s.watcher.StartWatching(s.shutdown)
}
