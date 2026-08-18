# linux_agent_install-shell

```
适用agent安装
```

```
#!/bin/bash

file="/tmp/igo-agent-1.9.3-1.0.x86_64.rpm"
libfile="/tmp/lib.tar.gz"
package="igo-agent-1.9.3-1.0.x86_64"
bin="/usr/local/bin/igo-agent"

help(){
    echo "${0} <download|install|remove|update>"
    exit 1
}

download(){
    # 请先将 igo-agent RPM 与 lib.tar.gz 放到 /tmp/ 后再执行 install
    :
}

install(){
    if [ ! -f "$file" ]; then
        echo "no rpm find, please first exec download!"
        exit 1
    fi
    rpm -ivh "$file" --nodeps --force
    export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/opt/igo-agent/embedded/lib/
    tar -zxvf /tmp/lib.tar.gz -C /opt/igo-agent/embedded/lib/python3.8/site-packages
    ln -s /opt/igo-agent/agent /usr/local/bin/igo-agent
}

update(){
    remove && download && install
}

remove(){
 
    if [ -f "$file" ]; then
        rm -rf "$file"
    fi
    if [ -f "$libfile" ]; then
        rm -rf "$libfile"
    fi 
    rpm -e "$package"
    rm -rf "$bin"
    rm -rf /opt/igo-agent
}

case "${1}" in
    download)
        download
        ;;
    install)
        install
        ;;
    update)
        update
        ;;
    remove)
        remove
        ;;
    *)
        help
        ;;
```
