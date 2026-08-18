# redis_metrics

```
Redis  v=5.0.9
cent76

http://192.168.14.164:9121/metrics
https://github.com/oliver006/redis_exporter
./redis_exporter -redis.addr redis://localhost:7001 -redis-only-metrics
```

```
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 0
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 1024
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 11
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 1.1862016e+07
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.64656657384e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 7.27597056e+08
# HELP process_virtual_memory_max_bytes Maximum amount of virtual memory available in bytes.
# TYPE process_virtual_memory_max_bytes gauge
process_virtual_memory_max_bytes 1.8446744073709552e+19
# HELP redis_active_defrag_running active_defrag_running metric
# TYPE redis_active_defrag_running gauge
redis_active_defrag_running 0
# HELP redis_allocator_active_bytes allocator_active_bytes metric
# TYPE redis_allocator_active_bytes gauge
redis_allocator_active_bytes 2.904064e+06
# HELP redis_allocator_allocated_bytes allocator_allocated_bytes metric
# TYPE redis_allocator_allocated_bytes gauge
redis_allocator_allocated_bytes 2.691136e+06
# HELP redis_allocator_frag_bytes allocator_frag_bytes metric
# TYPE redis_allocator_frag_bytes gauge
redis_allocator_frag_bytes 212928
# HELP redis_allocator_frag_ratio allocator_frag_ratio metric
# TYPE redis_allocator_frag_ratio gauge
redis_allocator_frag_ratio 1.08
# HELP redis_allocator_resident_bytes allocator_resident_bytes metric
# TYPE redis_allocator_resident_bytes gauge
redis_allocator_resident_bytes 7.499776e+06
# HELP redis_allocator_rss_bytes allocator_rss_bytes metric
# TYPE redis_allocator_rss_bytes gauge
redis_allocator_rss_bytes 4.595712e+06
# HELP redis_allocator_rss_ratio allocator_rss_ratio metric
# TYPE redis_allocator_rss_ratio gauge
redis_allocator_rss_ratio 2.58
# HELP redis_aof_current_rewrite_duration_sec aof_current_rewrite_duration_sec metric
# TYPE redis_aof_current_rewrite_duration_sec gauge
redis_aof_current_rewrite_duration_sec -1
# HELP redis_aof_enabled aof_enabled metric
# TYPE redis_aof_enabled gauge
redis_aof_enabled 0
# HELP redis_aof_last_bgrewrite_status aof_last_bgrewrite_status metric
# TYPE redis_aof_last_bgrewrite_status gauge
redis_aof_last_bgrewrite_status 1
# HELP redis_aof_last_cow_size_bytes aof_last_cow_size_bytes metric
# TYPE redis_aof_last_cow_size_bytes gauge
redis_aof_last_cow_size_bytes 0
# HELP redis_aof_last_rewrite_duration_sec aof_last_rewrite_duration_sec metric
# TYPE redis_aof_last_rewrite_duration_sec gauge
redis_aof_last_rewrite_duration_sec -1
# HELP redis_aof_last_write_status aof_last_write_status metric
# TYPE redis_aof_last_write_status gauge
redis_aof_last_write_status 1
# HELP redis_aof_rewrite_in_progress aof_rewrite_in_progress metric
# TYPE redis_aof_rewrite_in_progress gauge
redis_aof_rewrite_in_progress 0
# HELP redis_aof_rewrite_scheduled aof_rewrite_scheduled metric
# TYPE redis_aof_rewrite_scheduled gauge
redis_aof_rewrite_scheduled 0
# HELP redis_blocked_clients blocked_clients metric
# TYPE redis_blocked_clients gauge
redis_blocked_clients 0
# HELP redis_client_recent_max_input_buffer_bytes client_recent_max_input_buffer_bytes metric
# TYPE redis_client_recent_max_input_buffer_bytes gauge
redis_client_recent_max_input_buffer_bytes 2
# HELP redis_client_recent_max_output_buffer_bytes client_recent_max_output_buffer_bytes metric
# TYPE redis_client_recent_max_output_buffer_bytes gauge
redis_client_recent_max_output_buffer_bytes 0
# HELP redis_cluster_current_epoch cluster_current_epoch metric
# TYPE redis_cluster_current_epoch gauge
redis_cluster_current_epoch 6
# HELP redis_cluster_enabled cluster_enabled metric
# TYPE redis_cluster_enabled gauge
redis_cluster_enabled 1
# HELP redis_cluster_known_nodes cluster_known_nodes metric
# TYPE redis_cluster_known_nodes gauge
redis_cluster_known_nodes 6
# HELP redis_cluster_messages_received_total cluster_messages_received_total metric
# TYPE redis_cluster_messages_received_total gauge
redis_cluster_messages_received_total 67629
# HELP redis_cluster_messages_sent_total cluster_messages_sent_total metric
# TYPE redis_cluster_messages_sent_total gauge
redis_cluster_messages_sent_total 67629
# HELP redis_cluster_my_epoch cluster_my_epoch metric
# TYPE redis_cluster_my_epoch gauge
redis_cluster_my_epoch 1
# HELP redis_cluster_size cluster_size metric
# TYPE redis_cluster_size gauge
redis_cluster_size 3
# HELP redis_cluster_slots_assigned cluster_slots_assigned metric
# TYPE redis_cluster_slots_assigned gauge
redis_cluster_slots_assigned 16384
# HELP redis_cluster_slots_fail cluster_slots_fail metric
# TYPE redis_cluster_slots_fail gauge
redis_cluster_slots_fail 0
# HELP redis_cluster_slots_ok cluster_slots_ok metric
# TYPE redis_cluster_slots_ok gauge
redis_cluster_slots_ok 16384
# HELP redis_cluster_slots_pfail cluster_slots_pfail metric
# TYPE redis_cluster_slots_pfail gauge
redis_cluster_slots_pfail 0
# HELP redis_cluster_state cluster_state metric
# TYPE redis_cluster_state gauge
redis_cluster_state 1
# HELP redis_cluster_stats_messages_meet_received cluster_stats_messages_meet_received metric
# TYPE redis_cluster_stats_messages_meet_received gauge
redis_cluster_stats_messages_meet_received 5
# HELP redis_cluster_stats_messages_ping_received cluster_stats_messages_ping_received metric
# TYPE redis_cluster_stats_messages_ping_received gauge
redis_cluster_stats_messages_ping_received 34288
# HELP redis_cluster_stats_messages_ping_sent cluster_stats_messages_ping_sent metric
# TYPE redis_cluster_stats_messages_ping_sent gauge
redis_cluster_stats_messages_ping_sent 33336
# HELP redis_cluster_stats_messages_pong_received cluster_stats_messages_pong_received metric
# TYPE redis_cluster_stats_messages_pong_received gauge
redis_cluster_stats_messages_pong_received 33336
# HELP redis_cluster_stats_messages_pong_sent cluster_stats_messages_pong_sent metric
# TYPE redis_cluster_stats_messages_pong_sent gauge
redis_cluster_stats_messages_pong_sent 34293
# HELP redis_commands_duration_seconds_total Total amount of time in seconds spent per command
# TYPE redis_commands_duration_seconds_total counter
redis_commands_duration_seconds_total{cmd="client"} 5e-06
redis_commands_duration_seconds_total{cmd="cluster"} 0.017587
redis_commands_duration_seconds_total{cmd="command"} 1.900791
redis_commands_duration_seconds_total{cmd="config"} 0.000626
redis_commands_duration_seconds_total{cmd="dbsize"} 2e-06
redis_commands_duration_seconds_total{cmd="get"} 1.3e-05
redis_commands_duration_seconds_total{cmd="info"} 0.021512
redis_commands_duration_seconds_total{cmd="keys"} 2.4e-05
redis_commands_duration_seconds_total{cmd="ping"} 0
redis_commands_duration_seconds_total{cmd="psync"} 0.093431
redis_commands_duration_seconds_total{cmd="randomkey"} 4e-06
redis_commands_duration_seconds_total{cmd="replconf"} 0.047665
redis_commands_duration_seconds_total{cmd="set"} 0.000618
redis_commands_duration_seconds_total{cmd="type"} 3e-06
# HELP redis_commands_processed_total commands_processed_total metric
# TYPE redis_commands_processed_total counter
redis_commands_processed_total 43901
# HELP redis_commands_total Total number of calls per command
# TYPE redis_commands_total counter
redis_commands_total{cmd="client"} 1
redis_commands_total{cmd="cluster"} 86
redis_commands_total{cmd="command"} 304
redis_commands_total{cmd="config"} 35
redis_commands_total{cmd="dbsize"} 2
redis_commands_total{cmd="get"} 4
redis_commands_total{cmd="info"} 940
redis_commands_total{cmd="keys"} 4
redis_commands_total{cmd="ping"} 1
redis_commands_total{cmd="psync"} 35
redis_commands_total{cmd="randomkey"} 1
redis_commands_total{cmd="replconf"} 42387
redis_commands_total{cmd="set"} 100
redis_commands_total{cmd="type"} 1
# HELP redis_config_maxclients config_maxclients metric
# TYPE redis_config_maxclients gauge
redis_config_maxclients 10000
# HELP redis_config_maxmemory config_maxmemory metric
# TYPE redis_config_maxmemory gauge
redis_config_maxmemory 0
# HELP redis_connected_clients connected_clients metric
# TYPE redis_connected_clients gauge
redis_connected_clients 2
# HELP redis_connected_slave_lag_seconds Lag of connected slave
# TYPE redis_connected_slave_lag_seconds gauge
redis_connected_slave_lag_seconds{slave_ip="192.168.3.172",slave_port="7004",slave_state="online"} 0
redis_connected_slave_lag_seconds{slave_ip="192.168.3.191",slave_port="9320",slave_state="online"} 0
# HELP redis_connected_slave_offset_bytes Offset of connected slave
# TYPE redis_connected_slave_offset_bytes gauge
redis_connected_slave_offset_bytes{slave_ip="192.168.3.172",slave_port="7004",slave_state="online"} 49946
redis_connected_slave_offset_bytes{slave_ip="192.168.3.191",slave_port="9320",slave_state="online"} 49946
# HELP redis_connected_slaves connected_slaves metric
# TYPE redis_connected_slaves gauge
redis_connected_slaves 2
# HELP redis_connections_received_total connections_received_total metric
# TYPE redis_connections_received_total counter
redis_connections_received_total 454
# HELP redis_cpu_sys_children_seconds_total cpu_sys_children_seconds_total metric
# TYPE redis_cpu_sys_children_seconds_total counter
redis_cpu_sys_children_seconds_total 0.084401
# HELP redis_cpu_sys_seconds_total cpu_sys_seconds_total metric
# TYPE redis_cpu_sys_seconds_total counter
redis_cpu_sys_seconds_total 85.546046
# HELP redis_cpu_user_children_seconds_total cpu_user_children_seconds_total metric
# TYPE redis_cpu_user_children_seconds_total counter
redis_cpu_user_children_seconds_total 0.019359
# HELP redis_cpu_user_seconds_total cpu_user_seconds_total metric
# TYPE redis_cpu_user_seconds_total counter
redis_cpu_user_seconds_total 1.009593
# HELP redis_db_avg_ttl_seconds Avg TTL in seconds
# TYPE redis_db_avg_ttl_seconds gauge
redis_db_avg_ttl_seconds{db="db0"} 0
# HELP redis_db_keys Total number of keys by DB
# TYPE redis_db_keys gauge
redis_db_keys{db="db0"} 66
# HELP redis_db_keys_expiring Total number of expiring keys by DB
# TYPE redis_db_keys_expiring gauge
redis_db_keys_expiring{db="db0"} 0
# HELP redis_defrag_hits defrag_hits metric
# TYPE redis_defrag_hits gauge
redis_defrag_hits 0
# HELP redis_defrag_key_hits defrag_key_hits metric
# TYPE redis_defrag_key_hits gauge
redis_defrag_key_hits 0
# HELP redis_defrag_key_misses defrag_key_misses metric
# TYPE redis_defrag_key_misses gauge
redis_defrag_key_misses 0
# HELP redis_defrag_misses defrag_misses metric
# TYPE redis_defrag_misses gauge
redis_defrag_misses 0
# HELP redis_evicted_keys_total evicted_keys_total metric
# TYPE redis_evicted_keys_total counter
redis_evicted_keys_total 0
# HELP redis_expired_keys_total expired_keys_total metric
# TYPE redis_expired_keys_total counter
redis_expired_keys_total 0
# HELP redis_expired_stale_percentage expired_stale_percentage metric
# TYPE redis_expired_stale_percentage gauge
redis_expired_stale_percentage 0
# HELP redis_expired_time_cap_reached_total expired_time_cap_reached_total metric
# TYPE redis_expired_time_cap_reached_total gauge
redis_expired_time_cap_reached_total 0
# HELP redis_exporter_build_info redis exporter build_info
# TYPE redis_exporter_build_info gauge
redis_exporter_build_info{build_date="2022-02-15-12:44:18",commit_sha="a8f5b9cb192081fea5c95cf07dd5d13edd169389",golang_version="go1.17.7",version="v1.35.1"} 1
# HELP redis_exporter_last_scrape_connect_time_seconds exporter_last_scrape_connect_time_seconds metric
# TYPE redis_exporter_last_scrape_connect_time_seconds gauge
redis_exporter_last_scrape_connect_time_seconds 0.001559751
# HELP redis_exporter_last_scrape_duration_seconds exporter_last_scrape_duration_seconds metric
# TYPE redis_exporter_last_scrape_duration_seconds gauge
redis_exporter_last_scrape_duration_seconds 0.003644622
# HELP redis_exporter_last_scrape_error The last scrape error status.
# TYPE redis_exporter_last_scrape_error gauge
redis_exporter_last_scrape_error{err=""} 0
# HELP redis_exporter_scrape_duration_seconds Durations of scrapes by the exporter
# TYPE redis_exporter_scrape_duration_seconds summary
redis_exporter_scrape_duration_seconds_sum 0.003644622
redis_exporter_scrape_duration_seconds_count 1
# HELP redis_exporter_scrapes_total Current total redis scrapes.
# TYPE redis_exporter_scrapes_total counter
redis_exporter_scrapes_total 1
# HELP redis_instance_info Information about the Redis instance
# TYPE redis_instance_info gauge
redis_instance_info{maxmemory_policy="noeviction",os="Linux 3.10.0-1160.el7.x86_64 x86_64",process_id="15343",redis_build_id="f681d72b56ef8330",redis_mode="cluster",redis_version="5.0.9",role="master",run_id="7bbe5ed778fcc6edb877ee52a12e2be004537924",tcp_port="7001"} 1
# HELP redis_keyspace_hits_total keyspace_hits_total metric
# TYPE redis_keyspace_hits_total counter
redis_keyspace_hits_total 3
# HELP redis_keyspace_misses_total keyspace_misses_total metric
# TYPE redis_keyspace_misses_total counter
redis_keyspace_misses_total 2
# HELP redis_last_key_groups_scrape_duration_milliseconds Duration of the last key group metrics scrape in milliseconds
# TYPE redis_last_key_groups_scrape_duration_milliseconds gauge
redis_last_key_groups_scrape_duration_milliseconds 0
# HELP redis_last_slow_execution_duration_seconds The amount of time needed for last slow execution, in seconds
# TYPE redis_last_slow_execution_duration_seconds gauge
redis_last_slow_execution_duration_seconds 0.047373
# HELP redis_latest_fork_seconds latest_fork_seconds metric
# TYPE redis_latest_fork_seconds gauge
redis_latest_fork_seconds 0.000724
# HELP redis_lazyfree_pending_objects lazyfree_pending_objects metric
# TYPE redis_lazyfree_pending_objects gauge
redis_lazyfree_pending_objects 0
# HELP redis_loading_dump_file loading_dump_file metric
# TYPE redis_loading_dump_file gauge
redis_loading_dump_file 0
# HELP redis_master_repl_offset master_repl_offset metric
# TYPE redis_master_repl_offset gauge
redis_master_repl_offset 49946
# HELP redis_mem_clients_normal mem_clients_normal metric
# TYPE redis_mem_clients_normal gauge
redis_mem_clients_normal 66616
# HELP redis_mem_clients_slaves mem_clients_slaves metric
# TYPE redis_mem_clients_slaves gauge
redis_mem_clients_slaves 33844
# HELP redis_mem_fragmentation_bytes mem_fragmentation_bytes metric
# TYPE redis_mem_fragmentation_bytes gauge
redis_mem_fragmentation_bytes 1.2355008e+07
# HELP redis_mem_fragmentation_ratio mem_fragmentation_ratio metric
# TYPE redis_mem_fragmentation_ratio gauge
redis_mem_fragmentation_ratio 5.68
# HELP redis_mem_not_counted_for_eviction_bytes mem_not_counted_for_eviction_bytes metric
# TYPE redis_mem_not_counted_for_eviction_bytes gauge
redis_mem_not_counted_for_eviction_bytes 0
# HELP redis_memory_max_bytes memory_max_bytes metric
# TYPE redis_memory_max_bytes gauge
redis_memory_max_bytes 0
# HELP redis_memory_used_bytes memory_used_bytes metric
# TYPE redis_memory_used_bytes gauge
redis_memory_used_bytes 2.703408e+06
# HELP redis_memory_used_dataset_bytes memory_used_dataset_bytes metric
# TYPE redis_memory_used_dataset_bytes gauge
redis_memory_used_dataset_bytes 101036
# HELP redis_memory_used_lua_bytes memory_used_lua_bytes metric
# TYPE redis_memory_used_lua_bytes gauge
redis_memory_used_lua_bytes 37888
# HELP redis_memory_used_overhead_bytes memory_used_overhead_bytes metric
# TYPE redis_memory_used_overhead_bytes gauge
redis_memory_used_overhead_bytes 2.602372e+06
# HELP redis_memory_used_peak_bytes memory_used_peak_bytes metric
# TYPE redis_memory_used_peak_bytes gauge
redis_memory_used_peak_bytes 6.83544e+06
# HELP redis_memory_used_rss_bytes memory_used_rss_bytes metric
# TYPE redis_memory_used_rss_bytes gauge
redis_memory_used_rss_bytes 1.4995456e+07
# HELP redis_memory_used_scripts_bytes memory_used_scripts_bytes metric
# TYPE redis_memory_used_scripts_bytes gauge
redis_memory_used_scripts_bytes 0
# HELP redis_memory_used_startup_bytes memory_used_startup_bytes metric
# TYPE redis_memory_used_startup_bytes gauge
redis_memory_used_startup_bytes 1.449672e+06
# HELP redis_migrate_cached_sockets_total migrate_cached_sockets_total metric
# TYPE redis_migrate_cached_sockets_total gauge
redis_migrate_cached_sockets_total 0
# HELP redis_net_input_bytes_total net_input_bytes_total metric
# TYPE redis_net_input_bytes_total counter
redis_net_input_bytes_total 1.707007e+06
# HELP redis_net_output_bytes_total net_output_bytes_total metric
# TYPE redis_net_output_bytes_total counter
redis_net_output_bytes_total 4.089576e+06
# HELP redis_process_id process_id metric
# TYPE redis_process_id gauge
redis_process_id 15343
# HELP redis_pubsub_channels pubsub_channels metric
# TYPE redis_pubsub_channels gauge
redis_pubsub_channels 0
# HELP redis_pubsub_patterns pubsub_patterns metric
# TYPE redis_pubsub_patterns gauge
redis_pubsub_patterns 0
# HELP redis_rdb_bgsave_in_progress rdb_bgsave_in_progress metric
# TYPE redis_rdb_bgsave_in_progress gauge
redis_rdb_bgsave_in_progress 0
# HELP redis_rdb_changes_since_last_save rdb_changes_since_last_save metric
# TYPE redis_rdb_changes_since_last_save gauge
redis_rdb_changes_since_last_save 0
# HELP redis_rdb_current_bgsave_duration_sec rdb_current_bgsave_duration_sec metric
# TYPE redis_rdb_current_bgsave_duration_sec gauge
redis_rdb_current_bgsave_duration_sec -1
# HELP redis_rdb_last_bgsave_duration_sec rdb_last_bgsave_duration_sec metric
# TYPE redis_rdb_last_bgsave_duration_sec gauge
redis_rdb_last_bgsave_duration_sec 0
# HELP redis_rdb_last_bgsave_status rdb_last_bgsave_status metric
# TYPE redis_rdb_last_bgsave_status gauge
redis_rdb_last_bgsave_status 1
# HELP redis_rdb_last_cow_size_bytes rdb_last_cow_size_bytes metric
# TYPE redis_rdb_last_cow_size_bytes gauge
redis_rdb_last_cow_size_bytes 2.33472e+06
# HELP redis_rdb_last_save_timestamp_seconds rdb_last_save_timestamp_seconds metric
# TYPE redis_rdb_last_save_timestamp_seconds gauge
redis_rdb_last_save_timestamp_seconds 1.646565009e+09
# HELP redis_rejected_connections_total rejected_connections_total metric
# TYPE redis_rejected_connections_total counter
redis_rejected_connections_total 0
# HELP redis_repl_backlog_first_byte_offset repl_backlog_first_byte_offset metric
# TYPE redis_repl_backlog_first_byte_offset gauge
redis_repl_backlog_first_byte_offset 1
# HELP redis_repl_backlog_history_bytes repl_backlog_history_bytes metric
# TYPE redis_repl_backlog_history_bytes gauge
redis_repl_backlog_history_bytes 49946
# HELP redis_repl_backlog_is_active repl_backlog_is_active metric
# TYPE redis_repl_backlog_is_active gauge
redis_repl_backlog_is_active 1
# HELP redis_replica_partial_resync_accepted replica_partial_resync_accepted metric
# TYPE redis_replica_partial_resync_accepted gauge
redis_replica_partial_resync_accepted 0
# HELP redis_replica_partial_resync_denied replica_partial_resync_denied metric
# TYPE redis_replica_partial_resync_denied gauge
redis_replica_partial_resync_denied 1
# HELP redis_replica_resyncs_full replica_resyncs_full metric
# TYPE redis_replica_resyncs_full gauge
redis_replica_resyncs_full 35
# HELP redis_replication_backlog_bytes replication_backlog_bytes metric
# TYPE redis_replication_backlog_bytes gauge
redis_replication_backlog_bytes 1.048576e+06
# HELP redis_second_repl_offset second_repl_offset metric
# TYPE redis_second_repl_offset gauge
redis_second_repl_offset -1
# HELP redis_slave_expires_tracked_keys slave_expires_tracked_keys metric
# TYPE redis_slave_expires_tracked_keys gauge
redis_slave_expires_tracked_keys 0
# HELP redis_slowlog_last_id Last id of slowlog
# TYPE redis_slowlog_last_id gauge
redis_slowlog_last_id 52
# HELP redis_slowlog_length Total slowlog
# TYPE redis_slowlog_length gauge
redis_slowlog_length 53
# HELP redis_start_time_seconds Start time of the Redis instance since unix epoch in seconds.
# TYPE redis_start_time_seconds gauge
redis_start_time_seconds 1.646531176e+09
# HELP redis_target_scrape_request_errors_total Errors in requests to the exporter
# TYPE redis_target_scrape_request_errors_total counter
redis_target_scrape_request_errors_total 0
# HELP redis_up Information about the Redis instance
# TYPE redis_up gauge
redis_up 1
# HELP redis_uptime_in_seconds uptime_in_seconds metric
# TYPE redis_uptime_in_seconds gauge
redis_uptime_in_seconds 35416
```
