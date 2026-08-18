# k8s_postgre

安装个postgre，建库建表授权，然后通过configmap挂载给业务deployment使用  
ubuntu20.04  
k8s v1.24.6

## 安装部署

```
sealos run registry.cn-shanghai.aliyuncs.com/labring/bitnami-postgresql-ha:v15.4.0 \
-e NAME=postgre-ha -e NAMESPACE=postgre -e HELM_OPTS="--set service.type=NodePort"

CHART NAME: postgresql-ha
CHART VERSION: 11.9.4
APP VERSION: 15.4.0
** Please be patient while the chart is being deployed **
PostgreSQL can be accessed through Pgpool via port 5432 on the following DNS name from within your cluster:

    postgre-ha-postgresql-ha-pgpool.postgre.svc.cluster.local

Pgpool acts as a load balancer for PostgreSQL and forward read/write connections to the primary node while read-only connections are forwarded to standby nodes.

To get the password for "postgres" run:

    export POSTGRES_PASSWORD=$(kubectl get secret --namespace postgre postgre-ha-postgresql-ha-postgresql -o jsonpath="{.data.password}" | base64 -d)

To get the password for "repmgr" run:

    export REPMGR_PASSWORD=$(kubectl get secret --namespace postgre postgre-ha-postgresql-ha-postgresql -o jsonpath="{.data.repmgr-password}" | base64 -d)

To connect to your database run the following command:

    kubectl run postgre-ha-postgresql-ha-client --rm --tty -i --restart='Never' --namespace postgre --image docker.io/bitnami/postgresql-repmgr:15.4.0-debian-11-r31 --env="PGPASSWORD=$POSTGRES_PASSWORD"  \
        --command -- psql -h postgre-ha-postgresql-ha-pgpool -p 5432 -U postgres -d postgres

To connect to your database from outside the cluster execute the following commands:

    export NODE_IP=$(kubectl get nodes --namespace postgre -o jsonpath="{.items[0].status.addresses[0].address}")
    export NODE_PORT=$(kubectl get --namespace postgre -o jsonpath="{.spec.ports[0].nodePort}" services postgre-ha-postgresql-ha-pgpool
    PGPASSWORD="$POSTGRES_PASSWORD" psql -h $NODE_IP -p $NODE_PORT -U postgres -d postgres
```

## 创建需要的库和用户包括密码

```
1. 将密码保存到环境变量
2. 登录到postgre
3. 建用户，建库，授权
-- 1) 创建用户
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'dbigo') THEN
    CREATE ROLE dbigo LOGIN PASSWORD '#4Fz2sM#amah';
  END IF;
END $$;
-- 2) 创建数据库
SELECT format('CREATE DATABASE %I OWNER %I', 'prod_mes', 'dbigo')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'prod_mes')
\gexec
-- 3)（可选但推荐）保证 public schema 权限
\c prod_mes
GRANT ALL ON SCHEMA public TO dbigo;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO dbigo;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO dbigo;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO dbigo;

配置最大连接数：（会自动触发滚动更新）
kubectl -n my-bitnami-postgresql-ha set env sts/my-bitnami-postgresql-ha-postgresql POSTGRESQL_MAX_CONNECTIONS=300
```
