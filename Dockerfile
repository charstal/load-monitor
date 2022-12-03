FROM golang:1.17
WORKDIR /go/src/github.com/charstal/load-monitor
COPY . .
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN make build

FROM alpine:3.12

COPY --from=0 /go/src/github.com/charstal/load-monitor/bin/load-monitor /bin/load-monitor

CMD ["/bin/load-monitor"]
