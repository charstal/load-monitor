# Load Monitor

```
.
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
├── bin
├── go.mod
├── go.sum
├── main.go     // enter
├── manifest
├── offline     // offline project
└── pkg         // load monitor source code
```


# Caution

There are two label needed to be implement, search `Todo(label)`

```
courseLabel                        = "course_id"
nodeNameLabel                      = "kubernetes_node"
```

# Deploy

Requirement: `etcd`，`influxdb`

recommand:

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install my-release bitnami/etcd
```
