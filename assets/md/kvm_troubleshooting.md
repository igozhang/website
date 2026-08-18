# kvm_TroubleShooting

```
1.
# virsh start my-k8s-node-01
error: Failed to start domain my-k8s-node-01
error: internal error: process exited while connecting to monitor: warning: host doesn't support requested feature: CPUID.01H:EDX.ds [bit 21]
warning: host doesn't support requested feature: CPUID.01H:EDX.acpi [bit 22]
warning: host doesn't support requested feature: CPUID.01H:EDX.ht [bit 28]
warning: host doesn't support requested feature: CPUID.01H:EDX.tm [bit 29]
warning: host doesn't support requested feature: CPUID.01H:EDX.pbe [bit 31]
warning: host doesn't support requested feature: CPUID.01H:ECX.dtes64 [bit 2]
warning: host doesn't support requested feature: CPUID.01H:ECX.monitor [bit 3]
warning: host doesn't support requested feature: CPUID.01H:ECX.ds_cpl [bit 4]
warning: host doesn't support requested feature: CPUID.01H:ECX.smx [bit 6]
warning: host doesn't support requested feature: CPUID.01H:ECX.est [bit 7]
warning: host doesn't support requested feature: CPUID.01H:ECX.tm2 [bit 8]
warning: host doesn't support requested feature: CPUID.01H:ECX.xtpr [bit 14]
warning: host doesn't support requested feature: CPUID.01H:ECX.

solution:
# virsh edit my-k8s-node-01

mode='host-passthrough'
instead of:
mode='host-model'
```
