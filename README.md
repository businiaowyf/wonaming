# wonaming
This is the naming Resolver & Watcher implementaion for grpc balancer.

Wonaming supports etcd and consul as the service register and discovery backend.

## example

### etcdv3

#### client
go run main.go

#### server
go run main.go -addr="127.0.0.1:2379"

