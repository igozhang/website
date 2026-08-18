# git_crud;常规操作

workspace：工作区  
staging area：暂存区/缓存区  
local repository：版本库或本地仓库  
remote repository：远程仓库

```
git init - 初始化仓库；
git add . - 添加文件到暂存区；
git commit - 将暂存区内容添加到仓库中；
git clone	拷贝一份远程仓库，下载一个项目;

git clone https://gitee.com/igozhang/funclist.git
git add .
git commit -m 'all'
git push

Git 的工作就是创建和保存你的项目的快照及与之后的快照进行对比
提交与修改：
git add	添加文件到暂存区
git status	查看仓库当前的状态，显示有变更的文件；
git diff	比较文件的不同，即暂存区和工作区的差异；
git commit	提交暂存区到本地仓库；
git reset	回退版本；
git rm	删除工作区文件；
git mv	移动或重命名工作区文件；

提交日志：
git log	查看历史提交记录
git blame <file>	以列表形式查看指定文件的历史修改记录

远程操作：
git remote	远程仓库操作
git fetch	从远程获取代码库
git pull	下载远程代码并合并
git push	上传远程代码并合并

git remote rm name  # 删除远程仓库
git remote rename old_name new_name  # 修改仓库名
```

```
git config -help
优先级从高到低
.git/config 目录级git配置
~/.gitconfig 用户级git配置
/etc/gitconfig 系统级git配置

git每次提交都会使用这些信息
git config --global user.name igozhang
git config --global user.email igozhang@163.com
git config --global http.proxy http://<用戶名>:<密碼>@<代理主機的電子郵件>:<代理主機的端口號碼>
```
