package main

import (
	"context"
	"io"
	"log"
	"os"

	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

func main() {
	nodeName := "5b3s30"
	if len(os.Args) > 1 {
		nodeName = os.Args[1]
	}
	controllerAddress := "127.0.0.1:20443"
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(controllerAddress, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	var done = make(chan bool)
	client := pbv1.NewConfigControllerClient(conn)
	go subscribe(client, nodeName)
	<-done
}

func subscribe(client pbv1.ConfigControllerClient, nodeName string) {
	req := &pbv1.SubscriptionRequest{
		Name: nodeName,
	}
	stream, err := client.SubscribeListWatch(context.Background(), req)
	if err != nil {
		klog.Fatal(err)
	}
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.req(_) = _, %v", client, err)
		}
		if vn := response.GetVirtualNetwork(); vn != nil {
			klog.Infof("got vn %s", vn.Name)
		}
		if vmi := response.GetVirtualMachineInterface(); vmi != nil {
			klog.Infof("got vmi %s", vmi.Name)
		}
	}
}
