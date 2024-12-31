## Getting Started

### docker

```go
make run
```

### k8s by kind

```shell
go install sigs.k8s.io/kind@v0.26.0 && kind create cluster
make build
kind load docker-image notificationhub
kind load docker-image taskmanager
make k8s-deploy


> k get po -n taskhub
NAME                              READY   STATUS    RESTARTS   AGE
broker-747c4f8c98-996n4           1/1     Running   0          1s
db-5db964d545-59cdr               1/1     Running   0          1s
notificationhub-88bc67486-rv59l   1/1     Running   0          1s
taskmanager-6554c7c56c-qnzcq      1/1     Running   0          1s
```


### api

```shell
curl -v -X POST http://127.0.0.1:8080/task -H "Content-Type: application/json" -H "Authorization: tech@domain.com" -H "X-Role: technician" -d '{ "summary": "hello world" }'
```

```shell
curl -v -X POST http://127.0.0.1:8080/task -H "Content-Type: application/json" -H "Authorization: admin@domain.com" -H "X-Role: manager" -d '{ "summary": "hello world" }'
```

```shell
curl -v -X GET http://127.0.0.1:8080/task -H "Content-Type: application/json" -H "X-Role: technician"
```

```shell
curl -v -X GET http://127.0.0.1:8080/task -H "Content-Type: application/json" -H "X-Role: manager"
```