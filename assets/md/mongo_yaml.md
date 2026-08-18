# mongo-yaml

```
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: mongo-svc
  name: union-mongo-headless
  namespace: union
spec:
  clusterIP: None
  ports:
  - name: mongodb
    port: 27017
    protocol: TCP
    targetPort: 27017
  selector:
    app: mongo-pod
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: union-mongodb
  namespace: union
spec:
  selector:
    matchLabels:
      app: mongo-pod
  serviceName: mongo-svc
  replicas: 1
  template:
    metadata:
      labels:
        app: mongo-pod
    spec:
      containers:
      - name: mongodb
        image: bitnami/mongodb:4.4.5
        env:
        - name: ALLOW_EMPTY_PASSWORD
          value: "true"
        ports:
        - name: mongodb
          containerPort: 27017
        volumeMounts:
        - name: rbd-data
          mountPath: /bitnami
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
  volumeClaimTemplates:
  - metadata:
      name: rbd-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
      storageClassName: csi-rbd-sc
```
