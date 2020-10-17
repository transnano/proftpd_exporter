# proftpd-exporter ![Releases](https://github.com/transnano/proftpd_exporter/workflows/Releases/badge.svg) ![Publish Docker image](https://github.com/transnano/proftpd_exporter/workflows/Publish%20Docker%20image/badge.svg) ![Vulnerability Scan](https://github.com/transnano/proftpd_exporter/workflows/Vulnerability%20Scan/badge.svg) ![Haskell Dockerfile Linter](https://github.com/transnano/proftpd_exporter/workflows/Haskell%20Dockerfile%20Linter/badge.svg)

![License](https://img.shields.io/github/license/transnano/proftpd_exporter?style=flat)

![Container image version](https://img.shields.io/docker/v/transnano/proftpd-exporter/latest?style=flat)
![Container image size](https://img.shields.io/docker/image-size/transnano/proftpd-exporter/latest?style=flat)
![Container image pulls](https://img.shields.io/docker/pulls/transnano/proftpd-exporter?style=flat)

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/transnano/proftpd_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/transnano/proftpd_exporter)](https://goreportcard.com/report/github.com/transnano/proftpd_exporter)

Prometheus exporter for ProFTPD metrics.  
Supports ProFTPD 1.3.x

## How to use

``` shell
$ proftpd-exporter \
  --config.file=/path/to/proftpd.config
```

- `--config.file`: Specify certificate files with comma separated values.

## Output

```
# v1
proftpd_session_count{file="/path/to/proftpd.scoreboard",state="idle"} 18
proftpd_session_count{file="/path/to/proftpd.scoreboard",state="stor"} 10
proftpd_session_count{file="/path/to/proftpd.scoreboard",state="retr"} 1

# v2
proftpd_up_for_days{file="/path/to/proftpd.scoreboard",pid="1000"} 36
proftpd_session_count{file="/path/to/proftpd.scoreboard",state="idle"} 18
proftpd_session_count{file="/path/to/proftpd.scoreboard",state="stor"} 10
proftpd_session_count{file="/path/to/proftpd.scoreboard",state="retr"} 1
proftpd_session_info{file="/path/to/proftpd.scoreboard",sce_pid="10000",sce_user="",sce_begin_session="",sce_begin_idle="",sce_cmd="",sce_cmd_arg=""} 1
proftpd_session_info{file="/path/to/proftpd.scoreboard",sce_pid="10001",sce_user="",sce_begin_session="",sce_begin_idle="",sce_cmd="",sce_cmd_arg=""} 1
proftpd_session_info{file="/path/to/proftpd.scoreboard",sce_pid="10002",sce_user="",sce_begin_session="",sce_begin_idle="",sce_cmd="",sce_cmd_arg=""} 1
```

- `proftpd_session_count`: Represent the deadline of server certificate in Unixtime.
- `proftpd_session_info`: Represents the difference between the deadline of the server certificate and the current date-time.
