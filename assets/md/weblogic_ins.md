# weblogic_ins

## 3 Types of Installers

1. OS-specific Package installer—This type of installer is a standalone version of the installation program that includes the JDKs for the selected platform. The installer may be either an .exe file or a .bin file, depending on the selected platform.
2. Generic Package installer—This type of installer is a .jar file. It does not include the JRockit SDK and Sun JDK. You can use this type of installer to install the product on any supported platform on which Java is already installed.
3. Upgrade installer—Upgrade installers allow you to upgrade an existing WebLogic Server installation to the current patch release. For example, if you have WebLogic Server 10.3.0 installed, you can use an Upgrade installer to upgrade your installation to WebLogic Server 10.3.6. If a patch release is available for your current WebLogic Server installation, you can download an Upgrade installer from My Oracle Support.

引自[Fusion Middleware Installation Guide for Oracle WebLogic Server 10.3.6]<https://docs.oracle.com/middleware/11119/wls/WLSIG/prepare.htm>

1. BIN (Binary Distribution)  
   BIN 包通常是一个自解压可执行文件，包含完整的安装程序，适用于Windows或Linux等平台。它是一个图形化或者命令行交互式的安装过程，用户可以按照向导一步步完成安装，包括选择安装路径、配置选项等步骤。这种方式提供了较为直观和灵活的安装体验。
2. JAR (Java Archive)  
   JAR 包是一种Java应用程序的归档文件格式，对于WebLogic来说，它是用于无GUI环境下的安装。通过在命令行中运行java -jar wls\*.jar命令来启动安装程序。这种安装方式会进入控制台模式，适合在没有图形界面的服务器环境中进行自动化部署或安装。同样包含所有必要的文件以安装WebLogic Server，但安装过程可能需要手动输入更多的参数来进行配置。
3. ZIP (Compressed Archive)  
   ZIP 包是一个简单的压缩文件格式，它包含了WebLogic Server的所有必要组件，但不含安装脚本或安装程序。用户需要自行解压到目标目录，并手动配置环境变量、设置域以及配置服务器等步骤。这种方式灵活性最大，但要求管理员对WebLogic Server的安装和配置有深入理解。

总结：  
BIN包：完整安装程序，适合桌面系统和需要交互式安装的情况。  
JAR包：轻量级安装方式，适合服务器环境，尤其是没有GUI的情况下进行自动化安装。  
ZIP包：原始文件集合，仅提供基本的压缩文件，需手动完成全部安装与配置工作。

## 下载

https://www.oracle.com/middleware/technologies/weblogic-server-installers-downloads.html  

jar 包安装需预装 JDK（请从 Oracle 官网下载对应版本）。

```
tar -xf jdk-8u271-linux-x64.tar.gz
vim ~/.bashrc
#set java environment
export JAVA_HOME=/opt/jdk1.8.0_271
export CLASSPATH=.:$JAVA_HOME/lib/dt.jar:$JAVA_HOME/lib/tools.jar:$CLASSPATH
export PATH=$JAVA_HOME/bin:$PATH
```

## 重要概念

1. 域Domain  
   每个域都有一个配置文件（config.xml），该文件位于域目录的根目录下，记录了域的结构、组件关系以及服务器配置信息；  
   在域范围内，可以定义并共享各种资源，包括但不限于数据源、JMS 目标、连接器服务等，这些资源可在域中的所有服务器间透明地使用和管理；  
   域还包含了安全相关的配置，如用户和组、角色、访问控制列表（ACL）、SSL 设置等，确保域的安全运行。
2. 管理服务器Admin Server  
   它是域的管理中心，用于创建、配置、启动和停止 Managed Servers，以及管理域内的应用程序、JMS 队列、数据源等资源。
3. 受管服务器Managed Servers  
   受管服务器是在域中实际运行应用程序的实体。可以有多个 Managed Server，并且它们可以在集群中协同工作以实现高可用性和负载均衡。
4. 集群Clusters  
   集群是域中的一个或多个 Managed Servers 的逻辑分组，这些服务器共同协作来提供应用服务，支持故障转移和负载均衡，从而提高系统的可用性和扩展性；

## 3 Installation Modes

1. Graphical mode  
   Graphical-mode installation is an interactive, GUI-based method for installing your software. It can be run on both Windows and UNIX systems. For installation procedures；  
   包括 win的exe和 Linux的bin文件；  
   也包括 jar包的图形化安装方式；  
   注意：  
   带上这个参数：-Djava.security.egd=file:/dev/./urandom 否则时间巨长；  
   exp:  
   wls1036\_win32.exe -mode=console -Djava.io.tmpdir=C:\Temp -log=C:\logs\server\_install.log  
   -Djava.io.tmpdir=tmpdirpath
2. Console mode  
   Console-mode installation is an interactive, text-based method for installing your software from the command line, on either a UNIX system or a Windows system. For information about using this method；  
   exp:  
   java -jar wls1036\_generic.jar  
   java -jar fmw\_12.2.1.4.0\_wls.jar  
   java -Xmx1024m -jar wlsversion\_generic.jar
3. Silent mode  
   Silent-mode installation is a non-interactive method of installing your software. You use an XML properties file to specify the installation options. You can run silent-mode installation from either a script or from the command line. Silent-mode installation allows you to define the installation configuration only once, and then use that configuration to duplicate the installation on many machines. For more information about silent mode installation；

## Installation Screens

引自[Installation Screens]<https://docs.oracle.com/middleware/11119/wls/WLSIG/install_screens.htm#CECJHABF>

1. Welcome  
   always Next.
2. Choose Middleware Home Directory  
   exp: /data/Oracle/MiddleWare
3. Register for Security Updates  
   去掉所有勾选点下一步（内网不需要注册更新）
4. Choose Install Type (This screen is displayed only during an initial installation.)  
   4.1 Typic 典型安装  
   4.2 Custom 定制安装  
   直接典型，下一步
5. Choose Products and Components （You selected Custom installation./You are adding components to an existing installation.）  
   Next  
   根据实际情况选择
6. JDK Selection (This screen is displayed only if you selected a Custom installation. It is not displayed if you are running an Upgrade installer.) （只有Windows平台且第一次初始安装选择定制custom且拥有administrate权限的时候才会出现）  
   根据实际情况选择Next
7. Choose Product Installation Directories (This screen is displayed only during an initial installation.)  
   按默认点下一步就行，默认是第二步目录的下一级比如  
   /data/Oracle/MiddleWare/wlserver\_10.3  
   /data/Oracle/MiddleWare/coherence\_3.7
8. Confirm Product Installation Directory  
   Next
9. Install Windows Service （只有Windows平台且第一次初始安装选择定制custom且拥有administrate权限的时候才会出现）  
   Next
10. Choose Shortcut Location（只有Windows平台且第一次初始安装且拥有administrate权限的时候才会出现）  
    Next
11. Installation Summary  
    Next
12. Status  
    Next
13. Installation Complete  
    Done

## 配置域Domain

```
/home/ops/Oracle/Middleware/wlserver_10.3/common/bin/config.sh
注意：
1. Enable Development or Production Mode for this domain.
 ->1|Development Mode
   2|Production Mode

这两区别是是否启动自动部署，也就是域目录下的/autodeploy里面的文件是否会被自动部署；
模式切换：开发模式 -> 生产模式
Step 1：将domain下的bin/setDomainEnv.sh文件中的PRODUCTION_MODE="false"改为PRODUCTION_MODE="true"。
Step 2：然后重新运行/bin/startWebLogic.sh即可。

2.
```

## 启动Start

1. startWebLogic.sh 管理服务器
2. startManagedWebLogic.sh 受管服务器
3. startNodeManager.sh 节点管理(console)进而间接控制受管服务器的启动。

```
./startWebLogic.sh (Linux或Unix系统)
startWebLogic.cmd (Windows系统)
./startManagedWebLogic.sh localhost MyManagedServer  （启动受管服务器）

cd /home/ops/Oracle/Middleware/user_projects/domains/base_domain
nohup ./startWebLogic.sh &

/home/ops/Oracle/Middleware/wlserver_10.3/server/bin/startNodeManager.sh
nohup sh startNodeManager.sh > nodemanager.log 2>&1 &

/home/ops/Oracle/Middleware/user_projects/domains/base_domain/bin/startManagedWebLogic.sh
/home/ops/Oracle/Middleware/wlserver_10.3/common/bin/startManagedWebLogic.sh
```

## 访问Console

http://10.10.8.212:7001/console
