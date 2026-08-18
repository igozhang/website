# mysql_config

```
v5.7.27
master
# cat /etc/my.cnf
[client]
port		= 3306
socket		= /opt/mysql/var/mysql.sock
#default-character-set = utf8

[mysql]
prompt=\\u@\\d>

[mysqld]
server-id = 1
port		= 23306
bind-address    = 0.0.0.0
socket		= /opt/mysql/var/mysql.sock
skip-name-resolve
datadir = /data/mysql
character-set-server = utf8
max_connections = 1000
open_files_limit = 65535
log-error = /opt/mysql/var/error.log
pid-file = /opt/mysql/var/mysql.pid
back_log = 256
max_connect_errors = 256
table_open_cache = 2048
max_allowed_packet = 64M
binlog_cache_size = 2M
max_heap_table_size = 128M
sort_buffer_size = 8M
join_buffer_size = 8M
thread_cache_size = 256
#thread_concurrency = 8
query_cache_size = 256M
query_cache_limit = 2M
thread_stack = 192K
transaction_isolation = REPEATABLE-READ
tmp_table_size = 64M
log-bin = /data/binlog/mysql-bin
expire_logs_days=30
max_binlog_size = 512M
binlog_format = ROW
slow_query_log = 1
slow_query_log_file = /opt/mysql/var/slow_query.log
long_query_time = 0.5
key_buffer_size = 256M
read_buffer_size = 8M
read_rnd_buffer_size = 8M
bulk_insert_buffer_size = 64M
myisam_sort_buffer_size = 128M
myisam_max_sort_file_size = 10G
myisam_repair_threads = 1
explicit_defaults_for_timestamp=true
log_timestamps=SYSTEM

innodb_file_per_table=ON
#innodb_additional_mem_pool_size = 16M
innodb_buffer_pool_size = 24G
#innodb_data_file_path = ibdata1:10M:autoextend
#innodb_file_io_threads = 4
innodb_thread_concurrency = 16
innodb_flush_log_at_trx_commit = 1
innodb_log_buffer_size = 18M
innodb_log_file_size = 512M
innodb_open_files = 65535
innodb_log_files_in_group = 2
#innodb_max_dirty_pages_pct = 90
innodb_lock_wait_timeout = 120

[mysqldump]
quick
max_allowed_packet = 64M

[mysql]
no-auto-rehash
prompt=\\u@\\d>

[myisamchk]
key_buffer_size = 512M
sort_buffer_size = 512M
read_buffer = 8M
write_buffer = 8M
```

```
v5.7.27
slave
# cat /etc/my.cnf
[client]
port		= 3306
socket		= /opt/mysql/var/mysql.sock
default-character-set = utf8

[mysql]
prompt=\\u@\\d>

[mysqld]
server-id = 3
port		= 23306
bind-address    = 0.0.0.0
socket		= /opt/mysql/var/mysql.sock
skip-name-resolve
datadir = /data/mysql
character-set-server = utf8
max_connections = 1000
open_files_limit = 65535
log-error = /opt/mysql/var/error.log
pid-file = /opt/mysql/var/mysql.pid
back_log = 256
max_connect_errors = 256
table_open_cache = 2048
max_allowed_packet = 64M
binlog_cache_size = 2M
max_heap_table_size = 128M
sort_buffer_size = 8M
join_buffer_size = 8M
thread_cache_size = 256
#thread_concurrency = 8
query_cache_size = 256M
query_cache_limit = 2M
thread_stack = 192K
transaction_isolation = REPEATABLE-READ
tmp_table_size = 64M
log-bin = /data/binlog/mysql-bin
expire_logs_days=30
max_binlog_size = 512M
binlog_format = ROW
slow_query_log = 1
slow_query_log_file = /opt/mysql/var/slow_query.log
long_query_time = 1
key_buffer_size = 256M
read_buffer_size = 8M
read_rnd_buffer_size = 8M
bulk_insert_buffer_size = 64M
myisam_sort_buffer_size = 128M
myisam_max_sort_file_size = 10G
myisam_repair_threads = 1
explicit_defaults_for_timestamp=true
log_timestamps=SYSTEM

innodb_file_per_table=ON
#innodb_additional_mem_pool_size = 16M
innodb_buffer_pool_size = 24G
#innodb_data_file_path = ibdata1:10M:autoextend
#innodb_file_io_threads = 4
innodb_thread_concurrency = 16
innodb_flush_log_at_trx_commit = 1
innodb_log_buffer_size = 18M
innodb_log_file_size = 512M
innodb_open_files = 65535
innodb_log_files_in_group = 2
#innodb_max_dirty_pages_pct = 90
innodb_lock_wait_timeout = 120

replicate-wild-ignore-table = mysql.%
replicate-wild-ignore-table = sys.%
replicate-wild-ignore-table = information_schema.%
replicate-wild-ignore-table = performance_schema.%
log_slave_updates=ON

[mysqldump]
quick
max_allowed_packet = 64M

[mysql]
no-auto-rehash
prompt=\\u@\\d>

[myisamchk]
key_buffer_size = 512M
sort_buffer_size = 512M
read_buffer = 8M
write_buffer = 8M
```
