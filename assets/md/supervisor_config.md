# supervisor_config

```
iostat
tee >/etc/supervisord.d/iostat.ini <<EOF
[program:iostat]
command = iostat -dxk 1
autostart = true
startsecs = 5
user = root
stdout_logfile=/tmp/iostat.out
stdout_logfile_maxbytes = 10MB
stdout_logfile_backups = 1
EOF

kibana
tee >/etc/supervisord.d/kibana.ini <<EOF
[program:kibana]
directory=/opt/kibana
command=/opt/kibana/bin/kibana
numprocs=1
priority=3
autostart=true
startsecs=10
startretries=3
autorestart=true
redirect_stderr=true
stdout_logfile_maxbytes=50MB
stdout_logfile_backups=10
stdout_logfile=/var/log/supervisor/supervisor_kibana.log
EOF

filebeat
tee >/etc/supervisord.d/filebeat.ini <<EOF
[program:filebeat]
command=/opt/filebeat/filebeat -e -c /opt/filebeat/filebeat.yml
autorestart=true
autostart=true
stderr_logfile=/var/log/supervisor/filebeat_err.log
stdout_logfile=/var/log/supervisor/filebeat_out.log
user=root
stopsignal=INT
startsecs=10
EOF
```
