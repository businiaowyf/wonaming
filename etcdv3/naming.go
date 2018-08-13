package etcdv3

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
)

// Register register service with name as prefix to etcd, multi etcd addr should use ; to split
func Register(etcdAddr, name string, addr string, ttl int64) error {
	var err error

	if cli == nil {
		// If DialTimeout = 0, New() will return immediately.
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(etcdAddr, ";"),
			DialTimeout: 0 * time.Second,
		})
		if err != nil {
			log.Println("New etcdv3 client fail")
			return err
		}
	}

	log.Println("New etcdv3 client succeed")

	ticker := time.NewTicker(time.Second * time.Duration(ttl))

	go func() {
		for {
			getResp, err := cli.Get(context.Background(), "/"+scheme+"/"+name+"/"+addr)
			if err != nil {
				log.Println(err)
			} else if getResp.Count == 0 {
				err = withAlive(name, addr, ttl)
				if err != nil {
					log.Println(err)
				} else {
					log.Println("Set service in etcdv3 succeed")
				}
			} else {
				// do nothing
			}

			<-ticker.C
		}
	}()

	return nil
}

func withAlive(name string, addr string, ttl int64) error {
	leaseResp, err := cli.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}

	_, err = cli.Put(context.Background(), "/"+scheme+"/"+name+"/"+addr, addr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	_, err = cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	return nil
}

// UnRegister remove service from etcd
// The service will be removed automatically if the process is killed
func UnRegister(name string, addr string) {
	if cli != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		cli.Delete(ctx, "/"+scheme+"/"+name+"/"+addr)
		cli.Close()
	}
}
