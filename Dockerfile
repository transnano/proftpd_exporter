FROM golang:1.15.0
WORKDIR /go/src/github.com/transnano/proftpd_exporter/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o proftpd-exporter -ldflags "-s -w \
-X github.com/prometheus/common/version.Version=$(git describe --tags --abbrev=0) \
-X github.com/prometheus/common/version.BuildDate=$(date +%FT%T%z) \
-X github.com/prometheus/common/version.Branch=master \
-X github.com/prometheus/common/version.Revision=$(git rev-parse --short HEAD) \
-X github.com/prometheus/common/version.BuildUser=transnano"

FROM alpine:3.12.0
LABEL maintainer="Transnano <transnano.jp@gmail.com>"
RUN apk --no-cache add ca-certificates
EXPOSE 9065
COPY --from=0 /go/src/github.com/transnano/proftpd_exporter/proftpd-exporter /bin/proftpd-exporter
ENTRYPOINT ["/bin/proftpd-exporter"]
