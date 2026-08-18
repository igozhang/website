# nginx_metrics

```
nginx/1.10.3
cent76

192.168.14.166:8088/status/format/prometheus
https://github.com/vozlt/nginx-module-vts
```

```
# HELP nginx_vts_info Nginx info
# TYPE nginx_vts_info gauge
nginx_vts_info{hostname="cndgdlbvdc0c-14-166",module_version="0.1.19.dev.91bdb14",version="1.10.3"} 1
# HELP nginx_vts_start_time_seconds Nginx start time
# TYPE nginx_vts_start_time_seconds gauge
nginx_vts_start_time_seconds 1646207892.733
# HELP nginx_vts_main_connections Nginx connections
# TYPE nginx_vts_main_connections gauge
nginx_vts_main_connections{status="accepted"} 21
nginx_vts_main_connections{status="active"} 2
nginx_vts_main_connections{status="handled"} 21
nginx_vts_main_connections{status="reading"} 0
nginx_vts_main_connections{status="requests"} 107
nginx_vts_main_connections{status="waiting"} 1
nginx_vts_main_connections{status="writing"} 1
# HELP nginx_vts_main_shm_usage_bytes Shared memory [ngx_http_vhost_traffic_status] info
# TYPE nginx_vts_main_shm_usage_bytes gauge
nginx_vts_main_shm_usage_bytes{shared="max_size"} 1048575
nginx_vts_main_shm_usage_bytes{shared="used_size"} 3518
nginx_vts_main_shm_usage_bytes{shared="used_node"} 1
# HELP nginx_vts_server_bytes_total The request/response bytes
# TYPE nginx_vts_server_bytes_total counter
# HELP nginx_vts_server_requests_total The requests counter
# TYPE nginx_vts_server_requests_total counter
# HELP nginx_vts_server_request_seconds_total The request processing time in seconds
# TYPE nginx_vts_server_request_seconds_total counter
# HELP nginx_vts_server_request_seconds The average of request processing times in seconds
# TYPE nginx_vts_server_request_seconds gauge
# HELP nginx_vts_server_request_duration_seconds The histogram of request processing time
# TYPE nginx_vts_server_request_duration_seconds histogram
# HELP nginx_vts_server_cache_total The requests cache counter
# TYPE nginx_vts_server_cache_total counter
nginx_vts_server_bytes_total{host="localhost",direction="in"} 48667
nginx_vts_server_bytes_total{host="localhost",direction="out"} 3454284
nginx_vts_server_requests_total{host="localhost",code="1xx"} 0
nginx_vts_server_requests_total{host="localhost",code="2xx"} 103
nginx_vts_server_requests_total{host="localhost",code="3xx"} 3
nginx_vts_server_requests_total{host="localhost",code="4xx"} 0
nginx_vts_server_requests_total{host="localhost",code="5xx"} 0
nginx_vts_server_request_seconds_total{host="localhost"} 0.213
nginx_vts_server_request_seconds{host="localhost"} 0.000
nginx_vts_server_cache_total{host="localhost",status="miss"} 0
nginx_vts_server_cache_total{host="localhost",status="bypass"} 0
nginx_vts_server_cache_total{host="localhost",status="expired"} 0
nginx_vts_server_cache_total{host="localhost",status="stale"} 0
nginx_vts_server_cache_total{host="localhost",status="updating"} 0
nginx_vts_server_cache_total{host="localhost",status="revalidated"} 0
nginx_vts_server_cache_total{host="localhost",status="hit"} 0
nginx_vts_server_cache_total{host="localhost",status="scarce"} 0
nginx_vts_server_bytes_total{host="*",direction="in"} 48667
nginx_vts_server_bytes_total{host="*",direction="out"} 3454284
nginx_vts_server_requests_total{host="*",code="1xx"} 0
nginx_vts_server_requests_total{host="*",code="2xx"} 103
nginx_vts_server_requests_total{host="*",code="3xx"} 3
nginx_vts_server_requests_total{host="*",code="4xx"} 0
nginx_vts_server_requests_total{host="*",code="5xx"} 0
nginx_vts_server_request_seconds_total{host="*"} 0.213
nginx_vts_server_request_seconds{host="*"} 0.000
nginx_vts_server_cache_total{host="*",status="miss"} 0
nginx_vts_server_cache_total{host="*",status="bypass"} 0
nginx_vts_server_cache_total{host="*",status="expired"} 0
nginx_vts_server_cache_total{host="*",status="stale"} 0
nginx_vts_server_cache_total{host="*",status="updating"} 0
nginx_vts_server_cache_total{host="*",status="revalidated"} 0
nginx_vts_server_cache_total{host="*",status="hit"} 0
nginx_vts_server_cache_total{host="*",status="scarce"} 0
```
