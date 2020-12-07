FROM golang:1.13 as builder
WORKDIR /go/src/github.com/phanama/apiserver-version-exporter/
COPY apiserver-version-exporter.go go.mod go.sum ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix -o apiserver-version-exporter .


FROM quay.io/prometheus/busybox:latest as app
COPY --from=builder /go/src/github.com/phanama/apiserver-version-exporter/apiserver-version-exporter /usr/bin/
ARG ARCH="amd64"
ARG OS="linux"
EXPOSE      9101
ENTRYPOINT  [ "/usr/bin/apiserver-version-exporter" ]
