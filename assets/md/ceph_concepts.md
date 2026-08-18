# ceph概念

```
PG(Placement Group)
PGP(Placement Group for Placement purpose)
1. PG是指定存储池存储对象的归属组有多少个，PGP是存储池PG的OSD分布组合个数
2. PG的增加会引起PG内的数据进行迁移，迁移到不同的OSD上新生成的PG中
3. PGP的增加会引起部分PG的分布变化，但是不会引起PG内对象的变动。

三种存储接口：块存储、文件存储、对象存储
对象存储（RadosGW）：提供RESTful接口，也提供多种编程语言绑定。兼容S3、Swift；
块存储（RDB）：由RBD提供，可以直接作为磁盘挂载，内置了容灾机制；
文件系统（CephFS）：提供POSIX兼容的网络文件系统CephFS，专注于高性能、大容量存储；

服务角色定义
Monitors：监视器，维护集群状态的多种映射，同时提供认证和日志记录服务，包括有关monitor 节点端到端的信息，其中包括 Ceph 集群ID，监控主机名和IP以及端口。并且存储当前版本信息以及最新更改信息，通过 "ceph mon dump"查看 monitor map。

MDS（Metadata Server）：Ceph 元数据，主要保存的是Ceph文件系统的元数据。注意：ceph的块存储和ceph对象存储都不需要MDS。

OSD：即对象存储守护程序，但是它并非针对对象存储。是物理磁盘驱动器，将数据以对象的形式存储到集群中的每个节点的物理磁盘上。OSD负责存储数据、处理数据复制、恢复、回（Backfilling）、再平衡。完成存储数据的工作绝大多数是由 OSD daemon 进程实现。在构建 Ceph OSD的时候，建议采用SSD 磁盘以及xfs文件系统来格式化分区。此外OSD还对其它OSD进行心跳检测，检测结果汇报给Monitor

RADOS：Reliable Autonomic Distributed Object Store。RADOS是ceph存储集群的基础。在ceph中，所有数据都以对象的形式存储，并且无论什么数据类型，RADOS对象存储都将负责保存这些对象。RADOS层可以确保数据始终保持一致。

librados：librados库，为应用程度提供访问接口。同时也为块存储、对象存储、文件系统提供原生的接口。

RADOSGW：网关接口，提供对象存储服务。它使用librgw和librados来实现允许应用程序与Ceph对象存储建立连接。并且提供S3 和 Swift（openstack） 兼容的RESTful API接口。

RBD：块设备，它能够自动精简配置并可调整大小，而且将数据分散存储在多个OSD上。

CephFS：Ceph文件系统，与POSIX兼容的文件系统，基于librados封装原生接口。
```

![](doc-pic/ceph_install_cephadm_1.png)

ceph访问逻辑路径
