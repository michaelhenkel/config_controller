package server

import (
	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
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

func (sm *SubscriptionManager) GetSubscriptionChannel(node string) chan *pbv1.Resource {
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
	Channel chan *pbv1.Resource
	Init    bool
}

type configControllerServer struct {
	pbv1.UnimplementedConfigControllerServer
	SubscriptionManager *SubscriptionManager
}

func New(subscriptionManager *SubscriptionManager) *configControllerServer {
	s := &configControllerServer{
		SubscriptionManager: subscriptionManager,
	}
	return s
}

func (c *configControllerServer) SubscribeListWatch(req *pbv1.SubscriptionRequest, srv pbv1.ConfigController_SubscribeListWatchServer) error {
	conn := make(chan *pbv1.Resource)
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
	c.SubscriptionManager.NewSubscriberChan <- req.GetName()
	<-stopChan
	return nil
}
