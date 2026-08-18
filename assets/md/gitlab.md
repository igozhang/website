# gitlab

```
env
gitlab/gitlab-ce:14.7.5-ce.0
CentOS Linux release 7.6.1810 (Core)
docker-ce-19.03.15-3.el7

2C4G基本满负荷运行，极狐几乎跑不动，官版勉强但是会调用avatar的头像效果不好；
```

```
gitlab-ce

mkdir -p /data/gitlab/
export GITLAB_HOME=/data/gitlab
docker pull gitlab/gitlab-ce:14.7.5-ce.0

官方版本gitlab.com:
docker run --detach \
  --hostname h1.igozhang.cn \
  --publish 30443:443 --publish 3080:80 --publish 3022:22 \
  --name gitlab \
  --restart always \
  --volume $GITLAB_HOME/config:/etc/gitlab \
  --volume $GITLAB_HOME/logs:/var/log/gitlab \
  --volume $GITLAB_HOME/data:/var/opt/gitlab \
  --shm-size 256m \
  gitlab/gitlab-ce:14.7.5-ce.0

极狐版本gitlab.cn：
docker run --detach \
  --hostname 119.91.192.241 \
  --publish 30443:443 --publish 3080:80 --publish 3022:22 \
  --name gitlab \
  --restart always \
  --volume $GITLAB_HOME/config:/etc/gitlab \
  --volume $GITLAB_HOME/logs:/var/log/gitlab \
  --volume $GITLAB_HOME/data:/var/opt/gitlab \
  --shm-size 256m \
  gitlab-jh.tencentcloudcr.com/omnibus/gitlab-jh:14.5.0

获取初始密码
docker exec -it gitlab grep 'Password:' /etc/gitlab/initial_root_password
密码文件将在 24 小时后的第一次重新配置运行中自动删除。

备份gitlab
docker exec -t gitlab gitlab-backup create
/data/backups/1647000974_2022_03_11_14.7.5_gitlab_backup.tar
```

```
ssh-keygen -t rsa -C '13088826919@163.com'
cat ~/.ssh/id_rsa.pub
gitlab_login profile_SSH-Keys

mkdir funclist
cd funclist

git config --global user.name "igozhang"
git config --global user.email "13088826919@163.com"
git config --global --list
mkdir funclist
cd funclist
git init
touch igo.test
git commit -m "add igo.test"
git remote add origin http://119.91.192.241:3080/igozhang/funclist.git
git push -u origin "master"

用ssh死活git不成功，已经添加SSH Keys，报权限错误，原因待查，怀疑和端口非80有关,也可能是腾讯的公司网映射导致的，被迫使用http解决
（git remote add origin git@119.91.192.241:3080:igozhang/funclist.git）
（$ git push -u origin "master"）
（Permission denied, please try again）
```

```
码云gitee.com

Git 全局设置:
git config --global user.name "igozhang"
git config --global user.email "13088826919@163.com"
创建 git 仓库:
mkdir funclist
cd funclist
git init 
touch README.md
git add README.md
git commit -m "first commit"
git remote add origin https://gitee.com/igozhang/funclist.git
git push -u origin "master"
已有仓库：
cd existing_git_repo
git remote add origin https://gitee.com/igozhang/funclist.git
git push -u origin "master"
```
