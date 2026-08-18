# KVM使用ceph存储

```
env
Ubuntu 20.04.1
```

```
apt-get -y install ceph-common

添加密钥
cat > /tmp/key.xml << EOF
<secret ephemeral='no' private='no'>
<uuid>2c79423e-1c0b-4e97-ac93-5d80f211d0e8</uuid>
<usage type='ceph'>
<name>client.game_storage secret</name>
</usage>
</secret>
EOF

virsh secret-define /tmp/key.xml

virsh secret-set-value --secret 2c79423e-1c0b-4e97-ac93-5d80f211d0e8 --base64 AQBye8VhSUpDGxAArsoViVyf9Q2uh9KmAwC6yg==

添加网络盘
    <disk type='network' device='disk'>
      <driver name='qemu' type='raw' cache='none' discard='unmap'/>
      <auth username='game_storage'>
        <secret type='ceph' uuid='2c79423e-1c0b-4e97-ac93-5d80f211d0e8'/>
      </auth>
      <source protocol='rbd' name='data_pool/g1.test3.image'>
        <host name='ceph-01' port='6789'/>
        <host name='ceph-02' port='6789'/>
        <host name='ceph-03' port='6789'/>
        <host name='ceph-04' port='6789'/>
        <host name='ceph-05' port='6789'/>
      </source>
      <target dev='vdb' bus='virtio'/>
      <serial>G0-Z00000000000</serial>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x0d' function='0x0'/>
    </disk>
```
