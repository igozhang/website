# zookeeper_metrics

```
zookeeper-3.4.9
cent76

http://192.168.14.163:9143/metrics
https://github.com/carlpett/zookeeper_exporter/releases
```

```
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 3.4351e-05
go_gc_duration_seconds{quantile="0.25"} 4.1663e-05
go_gc_duration_seconds{quantile="0.5"} 4.5624e-05
go_gc_duration_seconds{quantile="0.75"} 5.4701e-05
go_gc_duration_seconds{quantile="1"} 9.5665e-05
go_gc_duration_seconds_sum 0.001142157
go_gc_duration_seconds_count 23
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 15
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 731888
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 4.100464e+06
# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table.
# TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 1.444095e+06
# HELP go_memstats_frees_total Total number of frees.
# TYPE go_memstats_frees_total counter
go_memstats_frees_total 8528
# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata.
# TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes 405504
# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use.
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes 731888
# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used.
# TYPE go_memstats_heap_idle_bytes gauge
go_memstats_heap_idle_bytes 3.981312e+06
# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use.
# TYPE go_memstats_heap_inuse_bytes gauge
go_memstats_heap_inuse_bytes 1.687552e+06
# HELP go_memstats_heap_objects Number of allocated objects.
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 5093
# HELP go_memstats_heap_released_bytes_total Total number of heap bytes released to OS.
# TYPE go_memstats_heap_released_bytes_total counter
go_memstats_heap_released_bytes_total 3.956736e+06
# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system.
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes 5.668864e+06
# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
# TYPE go_memstats_last_gc_time_seconds gauge
go_memstats_last_gc_time_seconds 1.6462748466768599e+09
# HELP go_memstats_lookups_total Total number of pointer lookups.
# TYPE go_memstats_lookups_total counter
go_memstats_lookups_total 72
# HELP go_memstats_mallocs_total Total number of mallocs.
# TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 13621
# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures.
# TYPE go_memstats_mcache_inuse_bytes gauge
go_memstats_mcache_inuse_bytes 6944
# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system.
# TYPE go_memstats_mcache_sys_bytes gauge
go_memstats_mcache_sys_bytes 16384
# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures.
# TYPE go_memstats_mspan_inuse_bytes gauge
go_memstats_mspan_inuse_bytes 28272
# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system.
# TYPE go_memstats_mspan_sys_bytes gauge
go_memstats_mspan_sys_bytes 49152
# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place.
# TYPE go_memstats_next_gc_bytes gauge
go_memstats_next_gc_bytes 4.194304e+06
# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations.
# TYPE go_memstats_other_sys_bytes gauge
go_memstats_other_sys_bytes 1.040377e+06
# HELP go_memstats_stack_inuse_bytes Number of bytes in use by the stack allocator.
# TYPE go_memstats_stack_inuse_bytes gauge
go_memstats_stack_inuse_bytes 622592
# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator.
# TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 622592
# HELP go_memstats_sys_bytes Number of bytes obtained by system. Sum of all system allocations.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 9.246968e+06
# HELP http_request_duration_microseconds The HTTP request latencies in microseconds.
# TYPE http_request_duration_microseconds summary
http_request_duration_microseconds{handler="prometheus",quantile="0.5"} NaN
http_request_duration_microseconds{handler="prometheus",quantile="0.9"} NaN
http_request_duration_microseconds{handler="prometheus",quantile="0.99"} NaN
http_request_duration_microseconds_sum{handler="prometheus"} 23102.192
http_request_duration_microseconds_count{handler="prometheus"} 4
# HELP http_request_size_bytes The HTTP request sizes in bytes.
# TYPE http_request_size_bytes summary
http_request_size_bytes{handler="prometheus",quantile="0.5"} NaN
http_request_size_bytes{handler="prometheus",quantile="0.9"} NaN
http_request_size_bytes{handler="prometheus",quantile="0.99"} NaN
http_request_size_bytes_sum{handler="prometheus"} 1377
http_request_size_bytes_count{handler="prometheus"} 4
# HELP http_requests_total Total number of HTTP requests made.
# TYPE http_requests_total counter
http_requests_total{code="200",handler="prometheus",method="get"} 4
# HELP http_response_size_bytes The HTTP response sizes in bytes.
# TYPE http_response_size_bytes summary
http_response_size_bytes{handler="prometheus",quantile="0.5"} NaN
http_response_size_bytes{handler="prometheus",quantile="0.9"} NaN
http_response_size_bytes{handler="prometheus",quantile="0.99"} NaN
http_response_size_bytes_sum{handler="prometheus"} 13975
http_response_size_bytes_count{handler="prometheus"} 4
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 0.19
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 1024
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 8
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 4.104192e+06
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.64604301896e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 1.6297984e+07
# HELP zk_approximate_data_size Approximate size of data set
# TYPE zk_approximate_data_size gauge
zk_approximate_data_size 35502
# HELP zk_avg_latency Average latency of requests
# TYPE zk_avg_latency gauge
zk_avg_latency 2
# HELP zk_ephemerals_count Number of ephemeral nodes
# TYPE zk_ephemerals_count gauge
zk_ephemerals_count 50
# HELP zk_max_file_descriptor_count Maximum number of open file descriptors
# TYPE zk_max_file_descriptor_count counter
zk_max_file_descriptor_count 2.048e+07
# HELP zk_max_latency Maximum seen latency of requests
# TYPE zk_max_latency gauge
zk_max_latency 55
# HELP zk_min_latency Minimum seen latency of requests
# TYPE zk_min_latency gauge
zk_min_latency 0
# HELP zk_num_alive_connections Number of active connections
# TYPE zk_num_alive_connections gauge
zk_num_alive_connections 3
# HELP zk_open_file_descriptor_count Number of open file descriptors
# TYPE zk_open_file_descriptor_count gauge
zk_open_file_descriptor_count 32
# HELP zk_outstanding_requests Number of outstanding requests
# TYPE zk_outstanding_requests gauge
zk_outstanding_requests 0
# HELP zk_packets_received Number of packets received
# TYPE zk_packets_received counter
zk_packets_received 9588
# HELP zk_packets_sent Number of packets sent
# TYPE zk_packets_sent counter
zk_packets_sent 10547
# HELP zk_server_state Server state (leader/follower)
# TYPE zk_server_state untyped
zk_server_state{state="standalone"} 1
# HELP zk_up Exporter successful
# TYPE zk_up gauge
zk_up 1
# HELP zk_watch_count Number of watches
# TYPE zk_watch_count gauge
zk_watch_count 846
# HELP zk_znode_count Number of znodes
# TYPE zk_znode_count gauge
zk_znode_count 430
# HELP zookeeper_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which zookeeper_exporter was built.
# TYPE zookeeper_exporter_build_info gauge
zookeeper_exporter_build_info{branch="HEAD",goversion="go1.10.8",revision="0e30f3afdcb8e36be1b2a41fa432f4626012e5c0",version="v1.1.0"} 1
```
