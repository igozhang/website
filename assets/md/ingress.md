# ingress

k8s\_1.24

## 安装

```
sealos run registry.cn-shanghai.aliyuncs.com/labring/ingress-nginx:4.1.0
一条命令解决（需要使用sealos高版本'v5.1.1'）：
sealos run registry.cn-shanghai.aliyuncs.com/labring/ingress-nginx:v1.8.1   \
-e HELM_OPTS="--set controller.kind=DaemonSet --set controller.hostNetwork=true --set controller.service.enabled=false --set controller.extraArgs.default-ssl-certificate=ingress-nginx/tls-default"

kubectl -n ingress-nginx create secret \
  tls tls-default \
  --cert=./tls.pem \
  --key=./tls.key
```

## 修补改造

```
方案一：等效“hostNetwork=true 和 service.enabled=false”
kubectl patch deployment ingress-nginx-controller -n ingress-nginx -p '{"spec":{"template":{"spec":{"hostNetwork":true}}}}'
kubectl patch service ingress-nginx-controller -n ingress-nginx -p '{"spec":{"type":"ClusterIP","clusterIP":"None"}}'
kubectl patch configmap ingress-nginx-controller -n ingress-nginx -p '{"data":{"values.yaml":"controller:\n  hostNetwork: true\n  service:\n    enabled: true\n    type: ClusterIP\n    clusterIP: None"}}'

kubectl patch deployment ingress-nginx-controller -n ingress-nginx -p '{"spec":{"replicas":6}}'

方案二：deployment改daemonset
kubectl get deploy ingress-nginx-controller -n ingress-nginx -o yaml > ingress-deploy-backup.yaml
kubectl apply -f ingress-daemonset.yaml

方案三：使用默认证书
说明：配置一个ingress-nginx默认证书，在业务ingress没有指定secret的情况下，使用默认证书
所有业务 Ingress 只需配置spec.tls.hosts，省略 secretName，自动使用全局默认证书
kubectl -n ingress-nginx create secret \
  tls tls-default \
  --cert=./tls.pem \
  --key=./tls.key

kubectl patch daemonset ingress-nginx-controller -n ingress-nginx --type=json -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--default-ssl-certificate=ingress-nginx/tls-default"}]'
kubectl rollout restart ds ingress-nginx-controller -n ingress-nginx
kubectl get ds ingress-nginx-controller -n ingress-nginx -o yaml | grep default-ssl

测试：
# cat /igo/tmp/test-tls-default.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-default-cert
  namespace: default
spec:
  ingressClassName: nginx
  rules:
  - host: test.igozhang.cn
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: rancher
            port:
              number: 80
  # 关键：这里不写任何 tls、不写 secretName
```

## tips

### daemonset文件

```
# cat ingress-daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.2.0
    helm.sh/chart: ingress-nginx-4.1.0
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: controller
      app.kubernetes.io/instance: ingress-nginx
      app.kubernetes.io/name: ingress-nginx
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/restartedAt: "2026-03-17T17:41:14+08:00"
      labels:
        app.kubernetes.io/component: controller
        app.kubernetes.io/instance: ingress-nginx
        app.kubernetes.io/name: ingress-nginx
    spec:
      # 关键：Worker 节点亲和性（不调度到 master）
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: node-role.kubernetes.io/control-plane
                operator: DoesNotExist
      # 容忍所有污点（必须加，否则无法调度）
      tolerations:
      - operator: Exists
        effect: NoSchedule
      - operator: Exists
        effect: NoExecute
      containers:
      - args:
        - /nginx-ingress-controller
        - --publish-service=$(POD_NAMESPACE)/ingress-nginx-controller
        - --election-id=ingress-controller-leader
        - --controller-class=k8s.io/ingress-nginx
        - --ingress-class=nginx
        - --configmap=$(POD_NAMESPACE)/ingress-nginx-controller
        - --validating-webhook=:8443
        - --validating-webhook-certificate=/usr/local/certificates/cert
        - --validating-webhook-key=/usr/local/certificates/key
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: LD_PRELOAD
          value: /usr/local/lib/libmimalloc.so
        image: k8s.gcr.io/ingress-nginx/controller:v1.2.0
        imagePullPolicy: IfNotPresent
        lifecycle:
          preStop:
            exec:
              command:
              - /wait-shutdown
        livenessProbe:
          failureThreshold: 5
          httpGet:
            path: /healthz
            port: 10254
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        name: controller
        ports:
        - containerPort: 80
          hostPort: 80
          name: http
          protocol: TCP
        - containerPort: 443
          hostPort: 443
          name: https
          protocol: TCP
        - containerPort: 8443
          hostPort: 8443
          name: webhook
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /healthz
            port: 10254
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        resources:
          requests:
            cpu: 100m
            memory: 90Mi
        securityContext:
          allowPrivilegeEscalation: true
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - ALL
          runAsUser: 101
        volumeMounts:
        - mountPath: /usr/local/certificates/
          name: webhook-cert
          readOnly: true
      dnsPolicy: ClusterFirst
      hostNetwork: true
      restartPolicy: Always
      serviceAccount: ingress-nginx
      serviceAccountName: ingress-nginx
      terminationGracePeriodSeconds: 300
      volumes:
      - name: webhook-cert
        secret:
          defaultMode: 420
          secretName: ingress-nginx-admission
```
