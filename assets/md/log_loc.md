# 日志位置log_loc

## 常用系统日志位置

### Centos7

/var/log/messages：系统消息日志  
/var/log/cron：cron 守护进程日志  
/var/log/boot.log：系统启动时的日志  
/var/log/dmesg：内核环缓冲区输出，记录了系统启动时硬件自检等信息

### Ubuntu20.04

/var/log/syslog：包含了登录信息、系统消息等在内的系统日志  
/var/log/kern.log：内核日志  
/var/log/dmesg：开机过程中的 kernel ring buffer 输出

journalctl --since "2023-03-03" --until "2023-03-03 03:00"

## 常用软件日志位置(默认)

### Web服务器软件

1. Apache HTTP Server  
   /var/log/apache2 或 /var/log/httpd下  
   access.log：记录所有对服务器的访问请求，包括客户端IP、请求时间、请求方法、URL、HTTP状态码等信息。  
   error.log：记录服务器运行时的错误和警告信息
2. Nginx  
   /var/log/nginx 目录  
   access.log：类似于Apache的访问日志，记录所有对服务器的访问请求。  
   error.log：记录Nginx运行时的错误和警告信息。
3. Microsoft Internet Information Services (IIS)  
   %SystemDrive%\inetpub\logs\LogFile
4. Tomcat (主要用于Java Servlet和JSP应用)  
   一般在安装目录下logs文件夹内  
   catalina.out：  
   这是Tomcat的主要日志文件，记录了启动和关闭Tomcat服务器时的详细信息，以及运行时的系统输出和错误信息。  
   localhost.\*.log：  
   这些日志文件记录了与特定主机（如localhost）相关的请求和错误信息。其中 \*.log 可能包括 access\_log（访问日志）和 localhost.log（错误日志）等。  
   manager.*.log 和 host-manager.*.log：  
   如果你使用了Tomcat的Web应用程序管理器（Manager App）和主机管理器（Host Manager），这些日志文件会记录这两个管理工具的相关信息。  
   tomcat.\*.log：  
   这些日志文件记录了与Tomcat自身相关的消息和错误。
5. WebSphere Application Server  
   位置： `%WAS_INSTALL_DIR%\profiles\<PROFILE_NAME>\logs`  
   %WAS\_INSTALL\_DIR% 是WebSphere Application Server的安装目录。  
   是你的服务器配置文件的名称，例如 AppSrv01。
   1. SystemOut.log：  
      记录系统输出信息，包括各种服务和应用程序的日志消息。
   2. SystemErr.log：  
      记录系统错误信息，包括异常和故障情况。
   3. native\_stderr.log 或 native\_stdout.log：  
      记录Java虚拟机（JVM）的启动和运行时输出，包括垃圾收集（GC）的信息。
   4. ffdc 文件夹：  
      包含Failure First Detection and Correction (FFDC) 报告，这些报告是在遇到严重问题时生成的详细诊断信息。
   5. trace.log：  
      如果启用了跟踪，这个文件会记录详细的执行跟踪信息。
   6. startServer.log 和 stopServer.log：  
      记录服务器启动和停止的相关事件。
   7. 对于特定的应用程序，可能还会有关于该应用程序的日志文件，它们的位置和名称可能会根据应用程序的配置而有所不同。
6. Oracle WebLogic Server  
   位置： `<DOMAIN_HOME>/servers/<SERVER_NAME>/logs` 其中：
   - `<DOMAIN_HOME>` 是你的WebLogic域的主目录。
   - `<SERVER_NAME>` 是你的WebLogic服务器的名称。在这个logs目录下，你可能会找到以下一些常见的日志文件：
   1. `server.log`：  
      这是主要的日志文件，记录了服务器启动、运行时消息和错误信息。
   2. `access.log`：  
      如果启用了HTTP访问日志记录，这个文件会记录所有的HTTP请求和响应。
   3. `diagnostic.log`：  
      记录诊断和跟踪信息。
   4. `stdout.log` 和 `stderr.log`：  
      记录Java虚拟机（JVM）的标准输出和标准错误。
   5. `boot.properties`：  
      包含用于安全引导WebLogic服务器的信息。
   6. `gc.log`：  
      如果启用了垃圾收集日志记录，这个文件会记录Java虚拟机的垃圾收集活动。
   7. `opatch.apply.log` 和 `opatch.rollback.log`：  
      记录应用或回滚补丁的操作。
   8. 对于特定的应用程序，可能还会有关于该应用程序的日志文件，它们的位置和名称可能会根据应用程序的配置而有所不同。

### 中间件

1. Apache Kafka：

- 默认情况下，Kafka的日志文件位于安装目录的 `logs` 文件夹下。例如，在Linux系统中，路径可能是 `/var/log/kafka` 或者是 Kafka 安装路径下的 `logs` 文件夹。
- 主要的日志文件包括 `kafkaServer.log`（服务器日志）和 `kafka-console-consumer.log`、`kafka-console-producer.log` 等与客户端相关的日志。

1. RabbitMQ：

- 在Linux系统中，RabbitMQ的日志文件通常位于 `/var/log/rabbitmq` 目录下。
- 主要的日志文件是 `rabbitmq.log`，记录了RabbitMQ服务器的运行信息和错误消息。

1. ActiveMQ：

- 在Linux系统中，ActiveMQ的日志文件通常位于 `/var/log/activemq` 目录下。
- 主要的日志文件是 `activemq.log`，记录了ActiveMQ服务器的运行信息和错误消息。

1. Redis：

- Redis的日志文件位置取决于你的配置。在默认情况下，Redis可能会将日志输出到标准输出（stdout）。
- 如果你配置了日志文件，它通常会位于Redis配置文件（如 `redis.conf`）中指定的位置，通过 `logfile` 参数设置。  
  ZooKeeper的日志文件位置通常在安装目录的 `logs` 文件夹下。以下是一些常见的日志文件：

1. ZooKeeper服务器日志：

- 在Linux系统中，ZooKeeper服务器的日志文件通常位于 `/var/log/zookeeper` 或者是 ZooKeeper 安装路径下的 `logs` 文件夹。
- 主要的日志文件是 `zookeeper.out` 或 `zookeeper.log`，记录了ZooKeeper服务器的启动、运行时消息和错误信息。  
  ZooKeeper事务日志（Transaction Logs）：
- 事务日志用于持久化所有的ZooKeeper状态更改，这些文件通常位于数据目录下，文件名以 `log.` 开头，后跟一个数字序列。 你可以通过修改ZooKeeper的配置文件 `zoo.cfg` 来更改日志文件的位置。在 `zoo.cfg` 中，相关的配置选项包括：
  - `dataDir`：指定数据目录，事务日志和其他数据文件将存储在这个目录下。
  - `dataLogDir`（可选）：如果设置了这个选项，事务日志将存储在这个单独的目录下。  
    `bash dataDir=/var/lib/zookeeper/data dataLogDir=/var/lib/zookeeper/logs`

请注意，这些是默认或常见的情况，实际的日志文件位置和配置可能会根据你的ZooKeeper安装和设置的具体情况而有所不同。在使用ZooKeeper时，建议查阅官方文档以获取更准确和详细的信息，并注意管理日志文件的大小和轮转策略，以防止磁盘空间被过度占用。

### 数据库服务器软件

1. MySQL：

- MySQL的日志文件位置取决于你的配置，通常在MySQL的数据目录下。
- 主要的日志文件包括：
  - `error.log` 或 `mysql.err`：错误日志，记录了MySQL服务器运行时的错误和警告信息。
  - `slow-query.log`：慢查询日志，记录执行时间超过 long\_query\_time 值的查询。
  - `general.log` 或 `query.log`：通用查询日志，记录所有对MySQL服务器的查询请求。
  - 二进制日志（Binary Logs）：用于复制和恢复，位于数据目录下，文件名通常以 `mysql-bin.` 开头。

1. PostgreSQL：

- PostgreSQL的日志文件位置通常在数据目录下的 `pg_log` 文件夹中。
- 主要的日志文件是 `postgresql-YYYY-MM-DD_HHMMSS.log`，其中 `YYYY-MM-DD_HHMMSS` 是日志文件创建的日期和时间。
- 另外还有 WAL（Write-Ahead Log） 文件，它们位于数据目录下的 `pg_wal` 文件夹中，文件名以 `000000010000000000000001` 等格式命名。

1. Oracle Database：

- Oracle数据库的日志文件位置取决于你的配置，通常在Oracle的快速恢复区（Fast Recovery Area, FRA）或者你指定的位置。
- 主要的日志文件包括：
  - Redo Log Files：重做日志文件，用于事务恢复。
  - Archive Log Files：归档日志文件，当重做日志切换时，旧的重做日志会被归档。
  - Alert Log：警报日志，记录了数据库启动、关闭以及运行过程中的重要事件和错误信息。默认位置通常是 `$ORACLE_HOME/rdbms/log`，文件名为 `alert_<DB_UNIQUE_NAME>.log`。

1. Microsoft SQL Server：

- SQL Server的日志文件位置与数据库的.mdf和.ldf文件在同一目录下。
- 主要的日志文件是 `.ldf` 文件，它是事务日志文件，记录了所有数据库事务的信息。
- 错误日志文件通常位于 `<SQL_SERVER_INSTALL_DIR>\MSSQL\Log` 目录下，文件名为 `ERRORLOG` 和 `ERRORLOG.n`（n表示数字，代表历史错误日志）。

1. MongoDB：

- MongoDB的日志文件位置取决于你的配置。在Linux系统中，通常可以在 `/var/log/mongodb` 或者是你指定的目录下找到日志文件。
- 主要的日志文件是 `mongod.log`，记录了MongoDB服务器的运行信息和错误消息。 要更改MongoDB的日志文件路径和名称，你可以在启动MongoDB时通过命令行参数指定，例如：

```
   mongod --logpath /path/to/your/mongodb.log
```

6. Elasticsearch：

- Elasticsearch的日志文件位置也取决于你的配置。在默认情况下，Elasticsearch的日志文件通常位于以下位置：
  - Linux：`/var/log/elasticsearch`
  - Windows：`%LOG_DIR%` （这个环境变量需要在安装或配置过程中设置）
  - macOS：`/usr/local/var/log/elasticsearch`
- 主要的日志文件是 `elasticsearch.log`，记录了Elasticsearch的运行信息和错误消息。 要更改Elasticsearch的日志文件路径和名称，你可以在Elasticsearch的配置文件（如 `elasticsearch.yml`）中进行设置，例如：

```
   path.logs: /path/to/your/elasticsearch_logs
```
