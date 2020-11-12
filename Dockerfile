FROM golang:1.15.5-buster as build

WORKDIR /go/src/github.com/transnano/proftpd_exporter/
# For building Go Module required
ENV GOPROXY=direct
ENV GO111MODULE=on
ENV GOARCH=amd64
ENV GOOS=linux
ENV CGO_ENABLED=0
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN  go mod download
# Copy the go source
COPY . .
RUN go build -o proftpd-exporter -ldflags "-s -w \
-X github.com/prometheus/common/version.Version=$(git describe --tags --abbrev=0) \
-X github.com/prometheus/common/version.BuildDate=$(date +%FT%T%z) \
-X github.com/prometheus/common/version.Branch=main \
-X github.com/prometheus/common/version.Revision=$(git rev-parse --short HEAD) \
-X github.com/prometheus/common/version.BuildUser=transnano"

# hadolint ignore=DL3006
FROM gcr.io/distroless/base-debian10
LABEL maintainer="Transnano <transnano.jp@gmail.com>"
COPY --from=build /go/src/github.com/transnano/proftpd_exporter/proftpd-exporter /proftpd-exporter
ENTRYPOINT ["/proftpd-exporter"]
