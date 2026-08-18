# Win_cron

## windows 定时任务

需求： 每天凌晨3点定期备份 D:\data 目录 到F:盘下，F盘为网络挂载盘  
工具： nfs客户端，windows任务计划程序

## 步骤

1. 安装nfs客户端  
   服务器管理器 > 管理(M) > 添加角色和功能 > 基于角色或基于功能的安装>  
   NFS客户端 > 安装 > 重启
2. 挂载网络nfs盘

```
mount -o nolock \\ip-10.10.10.192-igozhang.cn\mnt\vpfs_shares\vol-s_nfs_win_cron F:
```

3. 编写批处理脚本或powershell脚本（脚本见下文）
4. 使用windows任务计划程序，指定admin用户每天凌晨运行脚本  
   win 任务计划程序 > 操作 > 创建基本任务 > 操作\_启动程序(T) >  
   完成

notice:  
bat脚本直接选脚本  
ps1 powershell脚本选powershell程序，在参数那里输入脚本位置  
C:\Windows\SysWOW64\WindowsPowerShell\v1.0\PowerShell.exe  
-NonInteractive "c:\BackUp.ps1”

## 脚本

1. 批处理bat脚本  
   此脚本将D盘的data目录复制到F盘的backup目录下，并按当前日期创建子目录存放备份文件

```
@echo off
:: 设置日期变量，格式为 YYYY-MM-DD
for /f "tokens=2 delims==" %%i in ('"wmic os get localdatetime /value | findstr LocalDateTime"') do set datetime=%%i
set date=%datetime:~0,4%-%datetime:~4,2%-%datetime:~6,2%

:: 备份命令
xcopy "D:\data" "F:\backup\%date%" /D /E /C /R /I /K /Y

:: 结束
exit
```

2. powershell脚本

```
# 设置备份日期格式
$date = Get-Date -Format "yyyy-MM-dd"

# 定义源路径和目标路径
$sourcePath = "D:\data"
$destinationPath = "F:\backup\$date"

# 创建目标目录如果不存在
if (-not (Test-Path -Path $destinationPath)) {
    New-Item -ItemType Directory -Path $destinationPath
}

# 使用Copy-Item进行备份
Copy-Item -Path $sourcePath -Destination $destinationPath -Recurse -Force
```
