/**
 * Copyright 2015-2017, Wothing Co., Ltd.
 * All rights reserved.
 *
 * Created by elvizlai on 2017/11/28 11:48.
 */

package main

import (
	"fmt"
	"time"

	"github.com/businiaowyf/wonaming/etcdv3"
	"github.com/businiaowyf/wonaming/etcdv3/example/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

func main() {
	r := etcdv3.NewResolver("localhost:2379;localhost:22379;localhost:32379")
	resolver.Register(r)

	//https://github.com/grpc/grpc/blob/master/doc/naming.md
	conn, err := grpc.Dial(r.Scheme()+"://authority/project/test", grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := pb.NewHelloServiceClient(conn)

	for {
		resp, err := client.Echo(context.Background(), &pb.Payload{Data: "hello"}, grpc.FailFast(true))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(resp)
		}

		<-time.After(time.Second)
	}
}
