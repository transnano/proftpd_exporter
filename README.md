# proftpd-exporter

Check the deadline of server certificate on local.

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
