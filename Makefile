# Copyright 2021 PayPal
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

COMMONENVVAR=GOOS=$(shell uname -s | tr A-Z a-z) GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m)))
BUILDENVVAR=CGO_ENABLED=0

.PHONY: all
all: build
	chmod +x bin/load-monitor

.PHONY: build
build:
	$(COMMONENVVAR) $(BUILDENVVAR) go build -o bin/load-monitor main.go

.PHONY: clean
clean:
	rm -rf ./bin

image:
	docker build -t load-monitor:v0.0.1 .

remote: image
	docker tag load-monitor:v0.0.1 registry.cn-hangzhou.aliyuncs.com/charstal/load-monitor:v0.0.1
	docker push registry.cn-hangzhou.aliyuncs.com/charstal/load-monitor:v0.0.1