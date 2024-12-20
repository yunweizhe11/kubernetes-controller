# aliyun-clb-controller

<img src="https://github.com/yunweizhe11/kubernetes-controller/blob/master/image/arch.jpg">

## Building from source
```bash
#build
git clone https://github.com/yunweizhe11/kubernetes-controller.git
cd kubernetes-controller
docker build -t aliyun-clb-controller:latest -f Dockerfile .
```
## Configuration
Requires required parameter list:
- **`ACCESS_KEY_ID`**，Alibaba Cloud Account ACCESS_KEY_ID
- **`ACCESS_KEY_SECRET`**, Alibaba Cloud Account ACCESS_KEY_SECRET

Permissions required by Alibaba Cloud account:
- ECS resource query permissions
- CLB resource add, delete, modify and modify permissions

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: golang-service
  labels:
    app: golang-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aliyun-collector-service
  template:
    metadata:
      labels:
        app: aliyun-collector-service
    spec:
      serviceAccountName: service-watch-aliyun-collector
      containers:
      - name: aliyun-collector-service
        image: aliyun-clb-collector:latest
        env:
        - name: ACCESS_KEY_ID
          value: ""
        - name: ACCESS_KEY_SECRET
          value: ""
        livenessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - "ps aux | grep 'aliyun-clb-controller' | grep -v grep"
          initialDelaySeconds: 5
          periodSeconds: 5
          failureThreshold: 3
```
## Deploy services in kubernetes
Add roles and accounts in Kubernetes
```bash
kubectl apply -f deploy/ClusterRole.yaml
```
Add Service in Kubernetes
```bash
kubectl apply -f deploy/deployment.yaml 
```

## Usage 
Requires required parameter list:
- **`aliyun/clb_id `**，Corresponds to Alibaba Cloud CLB Resource Id
- **`aliyun/clb_port`**,  Corresponds to the Port that needs to be monitored on the Alibaba Cloud CLB Instance.
- **`aliyun/vpc_id`**, Corresponds to the VPCID to which ECS and clb belong
- **`aliyun/regionid`**, The Alibaba Cloud Region to which the resource belongs

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    aliyun/clb_id: "lb-bp117qz1mztx0r64oo6eg"
    aliyun/clb_port: "80"
    aliyun/vpc_id: "vpc-bp1km8lhqausyndk3gwol"
    aliyun/regionid: "cn-hangzhou"
  name: my-service
  labels:
    app: my-app
spec:
  type: NodePort
  ports:
  - port: 802
    targetPort: 8081
    protocol: TCP
    name: http
  selector:
    app: my-app
```

## Todo

**功能**
+ [x] TCP/UDP
+ [ ] HTTP/HTTPS
