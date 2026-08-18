# Elastic_APM

```
CentOS Linux release 7.8.2003 (Core)
es7.9.3
apm-server-7.9.3-x86_64
elastic-apm-agent-1.30.0.jar

https://www.elastic.co/guide/en/apm/agent/java/current/index.html
https://www.elastic.co/cn/blog/adding-free-and-open-elastic-apm-as-part-of-your-elastic-observability-deployment
```

elastic\_apm架构

```
apm-server
安装
curl -L -O https://artifacts.elastic.co/downloads/apm-server/apm-server-7.9.3-x86_64.rpm
rpm -vi apm-server-7.9.3-x86_64.rpm

配置
cp /etc/apm-server/apm-server.yml{,.$(date +%Y%m%d)}
tee >/etc/apm-server/apm-server.yml <<EOF
apm-server:
  host: "192.168.81.29:8200"

output.elasticsearch:
    hosts: ["http://192.168.81.34:29200","http://192.168.81.35:29200","http://192.168.81.36:29200"]

启动
systemctl start apm-server
systemctl enable apm-server

验证
# curl  http://192.168.81.29:8200
{
  "build_date": "2020-10-16T08:05:05Z",
  "build_sha": "af7b123f289d79c34c9da11471803f457eafa5c0",
  "version": "7.9.3"
}
```

```
elastic-apm-agent
下载
https://search.maven.org/remotecontent?filepath=co/elastic/apm/elastic-apm-agent/1.30.0/elastic-apm-agent-1.30.0.jar

配置
-javaagent:/data/elastic-apm-agent-1.30.0.jar -Delastic.apm.service_name=eg-was  -Delastic.apm.server_urls=http://192.168.81.29:8200 -Delastic.apm.application_packages=org.example
```

效果图例01

```
结论:
elastic_apm
展示和完整度都比较好;
使用(含搭建)较pinpoint更简单;
服务地图需要企业白金license才可以使用;
```
