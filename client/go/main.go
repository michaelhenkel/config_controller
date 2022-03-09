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
		switch response.Action {
		case pbv1.Response_ADD:
			add(response)
		case pbv1.Response_UPDATE:
			update(response)
		}
	}
}

func update(response *pbv1.Response) error {
	switch t := response.New.Resource.(type) {
	case *pbv1.Resource_VirtualNetwork:
		klog.Infof("got vn update: new %s old %s", t.VirtualNetwork.Name, response.Old.GetVirtualNetwork().Name)
	case *pbv1.Resource_VirtualMachineInterface:
		klog.Infof("got vmi update: new %s old %s", t.VirtualMachineInterface.Name, response.Old.GetVirtualMachineInterface().Name)
	case *pbv1.Resource_VirtualRouter:
		klog.Infof("got vrouter update: new %s old %s", t.VirtualRouter.Name, response.Old.GetVirtualRouter().Name)
	}
	return nil
}

func add(response *pbv1.Response) {
	switch t := response.New.Resource.(type) {
	case *pbv1.Resource_VirtualNetwork:
		klog.Infof("got vn %s add", t.VirtualNetwork.Name)
	case *pbv1.Resource_VirtualMachineInterface:
		klog.Infof("got vmi %s add", t.VirtualMachineInterface.Name)
	case *pbv1.Resource_VirtualRouter:
		klog.Infof("got vrouter %s add", t.VirtualRouter.Name)
	case *pbv1.Resource_VirtualMachine:
		klog.Infof("got vm %s add", t.VirtualMachine.Name)
	case *pbv1.Resource_RoutingInstance:
		klog.Infof("got ri %s add", t.RoutingInstance.Name)
	case *pbv1.Resource_InstanceIP:
		klog.Infof("got iip %s add", t.InstanceIP.Name)
	default:
		klog.Infof("got unknown %s add", t)
	}
}
