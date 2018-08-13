package etcdv3

import (
	"context"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/resolver"
)

const scheme = "wonamingv3"

var cli *clientv3.Client

type etcdBuilder struct {
	rawAddr string
}

type etcdResolver struct {
	cc   resolver.ClientConn
	done chan struct{}
}

// NewBuilder initialize an etcd client
func NewBuilder(etcdAddr string) resolver.Builder {
	return &etcdBuilder{
		rawAddr: etcdAddr,
	}
}

func (b *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	var err error

	if cli == nil {
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(b.rawAddr, ";"),
			DialTimeout: 0 * time.Second,
		})
		if err != nil {
			return nil, err
		}
	}

	// cc's type is *grpc.ccResolverWrapper
	log.Printf("watch: target.scheme=%v, target.Endpoint=%v", target.Scheme, target.Endpoint)
	r := &etcdResolver{
		cc:   cc,
		done: make(chan struct{}),
	}
	go r.watch("/" + target.Scheme + "/" + target.Endpoint + "/")
	return r, nil
}

func (r *etcdBuilder) Scheme() string {
	return scheme
}

// ResolveNow() and Close() are from interface resolver.Resolver
func (r *etcdResolver) ResolveNow(rn resolver.ResolveNowOption) {
	log.Println("Resolver ResolveNow") // TODO check
}

// Close closes the resolver.
// Close() will be invoked when close grpc.ClientConn
func (r *etcdResolver) Close() {
	log.Println("Resolver Close")
	close(r.done)
}

func (r *etcdResolver) watch(keyPrefix string) {
	var addrList []resolver.Address

	getResp, err := cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Println("etcdv3 Get error=", err)
	} else {
		for i := range getResp.Kvs {
			log.Printf("key=%v, value=%v, prefix=%v", string(getResp.Kvs[i].Key), string(getResp.Kvs[i].Value), keyPrefix)
			addrList = append(addrList, resolver.Address{Addr: strings.TrimPrefix(string(getResp.Kvs[i].Key), keyPrefix)})
		}
	}

	log.Printf("Init new address list, cc=%v, new addrlist=%v", reflect.TypeOf(r.cc), addrList)
	r.cc.NewAddress(addrList)

	// todo, Watch may be canceled, need to restart Watch()
	rch := cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	//for n := range rch {
	for {
		select {
		case <-r.done:
			log.Println("Connection closed, quit watch")
			return
		case n := <-rch:
			for _, ev := range n.Events {
				log.Println(n)
				if n.Err() != nil {
					log.Fatal("Event error=", n.Err())
				}
				addr := strings.TrimPrefix(string(ev.Kv.Key), keyPrefix)
				switch ev.Type {
				case mvccpb.PUT:
					if !exist(addrList, addr) {
						addrList = append(addrList, resolver.Address{Addr: addr})
						log.Printf("Add new address list, cc=%v, new addr=%v", r.cc, addr)
						r.cc.NewAddress(addrList)
					}
				case mvccpb.DELETE:
					if s, ok := remove(addrList, addr); ok {
						addrList = s
						log.Printf("Del new address list, cc=%v, new addr=%v", r.cc, addr)
						r.cc.NewAddress(addrList)
					}
				}
			}
		}
	}
	log.Println("watch quit")
}

func exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}
