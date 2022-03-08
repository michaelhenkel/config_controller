package server

import (
	"fmt"
	"net"

	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/k8s"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

type SubscriptionManager struct {
	Subscriptions     map[string]Subscription
	NewSubscriberChan chan string
}

func (sm *SubscriptionManager) AddSubscription(node string, subscription Subscription) {
	sm.Subscriptions[node] = subscription
}

func (sm *SubscriptionManager) RemoveSubscription(node string) {
	delete(sm.Subscriptions, node)
}

func (sm *SubscriptionManager) GetSubscriptionChannel(node string) chan *pbv1.Response {
	if subscription, ok := sm.Subscriptions[node]; ok {
		return subscription.Channel
	}
	return nil
}

func NewSubscriptionManager(newSubscriberChan chan string) *SubscriptionManager {
	var subscriptionMap = make(map[string]Subscription)
	return &SubscriptionManager{
		Subscriptions:     subscriptionMap,
		NewSubscriberChan: newSubscriberChan,
	}
}

type Subscription struct {
	Channel chan *pbv1.Response
	Init    bool
}

type ConfigController struct {
	pbv1.UnimplementedConfigControllerServer
	SubscriptionManager *SubscriptionManager
	k8sClient           *k8s.Client
}

func New(subscriptionManager *SubscriptionManager, k8sClient *k8s.Client) *ConfigController {
	s := &ConfigController{
		SubscriptionManager: subscriptionManager,
		k8sClient:           k8sClient,
	}
	return s
}

func (c *ConfigController) SubscribeListWatch(req *pbv1.SubscriptionRequest, srv pbv1.ConfigController_SubscribeListWatchServer) error {
	conn := make(chan *pbv1.Response)
	c.SubscriptionManager.AddSubscription(req.Name, Subscription{
		Channel: conn,
		Init:    false,
	})
	klog.Infof("new subscription request from node %s", req.Name)
	var stopChan = make(chan struct{})
	go func() {
		for {
			select {
			case <-srv.Context().Done():
				c.SubscriptionManager.RemoveSubscription(req.GetName())
				stopChan <- struct{}{}
			case response := <-conn:
				if status, ok := status.FromError(srv.Send(response)); ok {
					switch status.Code() {
					case codes.OK:
					case codes.Unavailable, codes.Canceled, codes.DeadlineExceeded:
						stopChan <- struct{}{}
					default:
						stopChan <- struct{}{}
					}
				}
			}
		}
	}()
	klog.Info("sending new subscription msg")
	c.k8sClient.NewSubscriber(req.Name, conn)
	<-stopChan
	return nil
}

func (c *Client) Start(k8sClient *k8s.Client) error {
	for !k8sClient.Initialized() {
	}
	var newSubscriberChan = make(chan string)
	subscriptionManager := NewSubscriptionManager(newSubscriberChan)
	grpcPort := 20443
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", grpcPort))
	if err != nil {
		klog.Error(err, "unable to start grpc server")
		return err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := New(subscriptionManager, k8sClient)
	pbv1.RegisterConfigControllerServer(grpcServer, s)
	klog.Infof("starting GRPC server on port %d", grpcPort)
	grpcServer.Serve(lis)
	return nil
}

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}
