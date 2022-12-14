/*
Copyright 2020 PayPal

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"

	"github.com/charstal/load-monitor/pkg/server"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetReportCaller(true)
	logLevel, evnLogLevelSet := os.LookupEnv("LOG_LEVEL")
	parsedLogLevel, err := log.ParseLevel(logLevel)
	if evnLogLevelSet && err != nil {
		log.Infof("unable to parse log level set; defaulting to: %v", log.GetLevel())
	}
	if err == nil {
		log.SetLevel(parsedLogLevel)
	}
}

func main() {

	ch := make(chan struct{}, 1)

	loadMonitorServer, err := server.NewServer(ch)

	if err != nil {
		log.Errorf("unable to create server: %v", err)
		panic("some error")
	}

	loadMonitorServer.Run()
	log.Infof("server starting")

	<-ch
}
