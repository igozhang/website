# OSPF配置ospf_cfg

基*于华为交换机S3700-EI*  
[OSPF典型配置](https://support.huawei.com/enterprise/zh/doc/EDOC1000069491/a25a2d1a) 什么是OSPFhttps://info.support.huawei.com/info-finder/encyclopedia/zh/OSPF.html

1. 每个OSPF进程的Router ID要保证在OSPF网络中唯一，否则会导致邻居不能正常建立、路由信息不正确的问题。建议在OSPF设备上单独为每个OSPF进程配置全网唯一的Router ID；
2. OSPF协议将自治系统划分成不同的区域（Area），其中区域号（Area ID）是0的称为骨干区域。OSPF要求所有非骨干区域必须与骨干区域保持连通，并且骨干区域的设备之间也要保持连通；

OSPF基本功能组网图

## 配置交换机

1. SwitchA\_Vlan

```
<HUAWEI> system-view
[HUAWEI] sysname SwitchA
[SwitchA] vlan batch 10 20
[SwitchA] interface gigabitethernet 1/0/1
[SwitchA-GigabitEthernet1/0/1] port link-type trunk
[SwitchA-GigabitEthernet1/0/1] port trunk allow-pass vlan 10
[SwitchA-GigabitEthernet1/0/1] quit
[SwitchA] interface gigabitethernet 1/0/2
[SwitchA-GigabitEthernet1/0/2] port link-type trunk
[SwitchA-GigabitEthernet1/0/2] port trunk allow-pass vlan 20
[SwitchA-GigabitEthernet1/0/2] quit
```

2. SwitchA\_Vlanif

```
[SwitchA] interface vlanif 10
[SwitchA-Vlanif10] ip address 192.168.0.1 24
[SwitchA-Vlanif10] quit
[SwitchA] interface vlanif 20
[SwitchA-Vlanif20] ip address 192.168.1.1 24
[SwitchA-Vlanif20] quit
```

3. SwitchA\_OSPF

```
[SwitchA] ospf 1 router-id 10.1.1.1   //创建进程号为1，Router ID为10.1.1.1的OSPF进程
[SwitchA-ospf-1] area 0   //创建area 0区域并进入area 0视图
[SwitchA-ospf-1-area-0.0.0.0] network 192.168.0.0 0.0.0.255   //配置area 0所包含的网段
[SwitchA-ospf-1-area-0.0.0.0] quit
[SwitchA-ospf-1] area 1   //创建area 1区域并进入area 1视图
[SwitchA-ospf-1-area-0.0.0.1] network 192.168.1.0 0.0.0.255   //配置area 1所包含的网段
[SwitchA-ospf-1-area-0.0.0.1] return
```

4. SwitchB\_OSPF\_withoutIFcfg 省略接口及初始配置，参考SwitchA就行

```
[SwitchB] ospf 1 router-id 10.2.2.2
[SwitchB-ospf-1] area 0
[SwitchB-ospf-1-area-0.0.0.0] network 192.168.0.0 0.0.0.255
[SwitchB-ospf-1-area-0.0.0.0] return
```

5. SwitchC\_OSPF\_withoutIFcfg

```
[SwitchC] ospf 1 router-id 10.3.3.3
[SwitchC-ospf-1] area 1
[SwitchC-ospf-1-area-0.0.0.1] network 192.168.1.0 0.0.0.255
[SwitchC-ospf-1-area-0.0.0.1] return
```

## 验证

1. 查看SwitchA的OSPF邻居

```
<SwitchA> display ospf peer

          OSPF Process 1 with Router ID 10.1.1.1
                  Neighbors

 Area 0.0.0.0 interface 192.168.0.1(Vlanif10)'s neighbors
Router ID: 10.2.2.2      Address: 192.168.0.2
   State: Full  Mode:Nbr is  Master  Priority: 1
   DR: 192.168.0.2  BDR: 192.168.0.1   MTU: 0
   Dead timer due in 36  sec
   Retrans timer interval: 5
   Neighbor is up for 00:15:04
   Authentication Sequence: [ 0 ]

                  Neighbors

 Area 0.0.0.1 interface 192.168.1.1(Vlanif20)'s neighbors
Router ID: 10.3.3.3       Address: 192.168.1.2
   State: Full  Mode:Nbr is  Master  Priority: 1
   DR: 192.168.1.2  BDR: 192.168.1.1   MTU: 0
   Dead timer due in 39  sec
   Retrans timer interval: 5
   Neighbor is up for 00:07:32
   Authentication Sequence: [ 0 ]
```

2. 查看SwitchC的OSPF邻居

```
<SwitchC> display ospf routing

          OSPF Process 1 with Router ID 10.3.3.3
                   Routing Tables

 Routing for Network
 Destination      Cost  Type         NextHop         AdvRouter       Area
 192.168.1.0/24    1    Transit      192.168.1.2     10.3.3.3        0.0.0.1
 192.168.0.0/24    2    Inter-area   192.168.1.1     10.1.1.1        0.0.0.1

 Total Nets: 2
 Intra Area: 1  Inter Area: 1  ASE: 0  NSSA: 0
```
