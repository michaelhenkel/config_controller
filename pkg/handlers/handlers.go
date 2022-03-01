package handlers

import (
	"reflect"

	"github.com/michaelhenkel/config_controller/pkg/server"
	"github.com/michaelhenkel/config_controller/pkg/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

var converterMap = map[string]Resource{"VirtualNetwork": &VirtualNetwork{}}

type Resource interface {
	Convert(obj interface{}) error
	Add(subscriptionManager *server.SubscriptionManager, storeClient store.Store) error
	Init(subscriptionManager *server.SubscriptionManager, node string) error
	Update(subscriptionManager *server.SubscriptionManager, storeClient store.Store) error
	Delete()
}

type Handler struct {
	subscriptionManager *server.SubscriptionManager
	storeClient         store.Store
}

func NewHandler(subscriptionManager *server.SubscriptionManager, store store.Store) *Handler {
	return &Handler{
		subscriptionManager: subscriptionManager,
		storeClient:         store,
	}
}

func (h *Handler) Init(obj interface{}, node string) error {
	var kind string
	valueOf := reflect.ValueOf(obj)
	if valueOf.Type().Kind() == reflect.Ptr {
		kind = reflect.Indirect(valueOf).Type().Name()
	} else {
		kind = valueOf.Type().Name()
	}
	newRes, ok := converterMap[kind]
	if ok {
		if err := newRes.Convert(obj); err != nil {
			klog.Error(err)
			return err
		}
		if err := newRes.Init(h.subscriptionManager, node); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) Add(obj interface{}) error {
	var kind string
	valueOf := reflect.ValueOf(obj)
	if valueOf.Type().Kind() == reflect.Ptr {
		kind = reflect.Indirect(valueOf).Type().Name()
	} else {
		kind = valueOf.Type().Name()
	}
	klog.Infof("add event %s", kind)
	newRes, ok := converterMap[kind]
	if ok {
		if err := newRes.Convert(obj); err != nil {
			klog.Error(err)
			return err
		}
		if err := newRes.Add(h.subscriptionManager, h.storeClient); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) Update(obj interface{}) error {
	var kind string
	valueOf := reflect.ValueOf(obj)
	if valueOf.Type().Kind() == reflect.Ptr {
		kind = reflect.Indirect(valueOf).Type().Name()
	} else {
		kind = valueOf.Type().Name()
	}

	newRes, ok := converterMap[kind]
	if ok {
		if err := newRes.Convert(obj); err != nil {
			klog.Error(err)
			return err
		}
		if err := newRes.Update(h.subscriptionManager, h.storeClient); err != nil {
			return err
		}
	}

	return nil
}

func HandleDelete(res *unstructured.Unstructured) {

}
