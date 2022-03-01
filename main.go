package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/michaelhenkel/config_controller/pkg/handlers"
	"github.com/michaelhenkel/config_controller/pkg/server"
	"github.com/michaelhenkel/config_controller/pkg/store"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	contrailClient "ssd-git.juniper.net/contrail/cn2/contrail/pkg/client/clientset_generated/clientset"
	contrailInformer "ssd-git.juniper.net/contrail/cn2/contrail/pkg/client/informers_generated/externalversions"
)

const (
	Closed   = 0
	Added    = 1
	Modified = 2
	Deleted  = 3
	Error    = -1
)

type ClientSet struct {
	Kube     *kubernetes.Clientset
	Contrail *contrailClient.Clientset
	Dynamic  dynamic.Interface
}

func main() {
	var stopCh = make(chan struct{})
	var newSubscriberChan = make(chan string)
	subscriptionManager := server.NewSubscriptionManager(newSubscriberChan)
	go ClientWatch(stopCh, subscriptionManager)
	go RunGRPCServer(subscriptionManager)
	<-stopCh
}

func NewClientSet() (*ClientSet, error) {
	var err error
	var kconfig string
	config, _ := rest.InClusterConfig()
	if config == nil {
		if home := homedir.HomeDir(); home != "" {
			kconfig = filepath.Join(home, ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kconfig)
		if err != nil {
			return nil, err
		}
	}
	contrailClientSet, err := contrailClient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	kubernetesClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	dynamicClientSet, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &ClientSet{
		Kube:     kubernetesClientSet,
		Contrail: contrailClientSet,
		Dynamic:  dynamicClientSet,
	}, nil
}
func NewSharedInformerFactory2(clientSet *ClientSet, subscriptionManager *server.SubscriptionManager, storeClient store.Store) (map[string]cache.SharedInformer, error) {
	gvrMap, err := getGVRMap(clientSet)
	if err != nil {
		return nil, err
	}

	var sharedInformerMap = make(map[string]cache.SharedInformer)
	kubeFactory := informers.NewSharedInformerFactory(clientSet.Kube, time.Minute*10)
	namespaceInformer := kubeFactory.Core().V1().Namespaces().Informer()
	storeClient.Add("namespaces", namespaceInformer.GetStore())
	namespaceInformer.AddEventHandler(resourceEventHandler(&watchHandlerFunc{
		Handler: handlers.NewHandler(subscriptionManager, storeClient),
	}))
	sharedInformerMap["namespaces"] = namespaceInformer

	contrailFactory := contrailInformer.NewSharedInformerFactory(clientSet.Contrail, time.Minute*10)
	for _, gvr := range gvrMap {
		cInformer, err := contrailFactory.ForResource(gvr)
		if err != nil {
			return nil, err
		}
		storeClient.Add(gvr.Resource, cInformer.Informer().GetStore())
		cInformer.Informer().AddEventHandler(resourceEventHandler(&watchHandlerFunc{
			Handler: handlers.NewHandler(subscriptionManager, storeClient),
		}))

		sharedInformerMap[gvr.Resource] = cInformer.Informer()
	}
	return sharedInformerMap, nil
}

func getGVRMap(clientSet *ClientSet) (map[string]schema.GroupVersionResource, error) {
	var gvrMap = make(map[string]schema.GroupVersionResource)
	contrailResources, err := clientSet.Contrail.DiscoveryClient.ServerResourcesForGroupVersion("core.contrail.juniper.net/v1alpha1")
	if err != nil {
		return nil, err
	}

	for _, contrailResource := range contrailResources.APIResources {
		resourceNameList := strings.Split(contrailResource.Name, "/status")
		gvrMap[resourceNameList[0]] = schema.GroupVersionResource{
			Group:    "core.contrail.juniper.net",
			Version:  "v1alpha1",
			Resource: resourceNameList[0],
		}
	}
	return gvrMap, nil
}

type watchHandlerFunc struct {
	Handler *handlers.Handler
}

func (h *watchHandlerFunc) HandleEvent(eventType int, res *unstructured.Unstructured) error {
	return nil
}

func (h *watchHandlerFunc) HandleAddEvent(obj interface{}) error {
	return h.Handler.Update(obj)
}

func (h *watchHandlerFunc) HandleUpdateEvent(oldRes *unstructured.Unstructured, newRes *unstructured.Unstructured) error {
	return nil
}

func (h *watchHandlerFunc) HandleDeleteEvent(res *unstructured.Unstructured) error {
	return nil
}

func resourceEventHandler(handler WatchEventHandler) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			handler.HandleAddEvent(obj)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			handler.HandleAddEvent(newObj)

		},
		DeleteFunc: func(obj interface{}) {
			res, ok := obj.(*unstructured.Unstructured)
			if ok {
				handler.HandleDeleteEvent(res)
			}

		},
	}
}

type WatchEventHandler interface {
	HandleEvent(eventType int, res *unstructured.Unstructured) error
	HandleAddEvent(obj interface{}) error
	HandleUpdateEvent(oldRes *unstructured.Unstructured, newRes *unstructured.Unstructured) error
	HandleDeleteEvent(res *unstructured.Unstructured) error
}

func ClientWatch(stopChan chan struct{}, subscriptionManager *server.SubscriptionManager) error {
	klog.Info("starting client watch")
	clientSet, err := NewClientSet()
	if err != nil {
		return err
	}
	storeClient := store.New()
	sharedInformerMap, err := NewSharedInformerFactory2(clientSet, subscriptionManager, storeClient)

	var syncMap = make(map[string]bool)
	mux := &sync.RWMutex{}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	store := store.New()
	for resource, sharedInformer := range sharedInformerMap {
		store.Add(resource, sharedInformer.GetStore())
		go sharedInformer.Run(ctx.Done())
		isSynced := cache.WaitForCacheSync(ctx.Done(), sharedInformer.HasSynced)
		mux.Lock()
		syncMap[resource] = isSynced
		mux.Unlock()
	}

	for _, isSynced := range syncMap {
		if !isSynced {
			return err
		}
	}

	go HandleNewSubscriber(store, subscriptionManager)
	<-ctx.Done()
	return nil
}

func HandleNewSubscriber(storeClient store.Store, subscriptionManager *server.SubscriptionManager) error {
	hdl := handlers.NewHandler(subscriptionManager, storeClient)
	for node := range subscriptionManager.NewSubscriberChan {
		list := storeClient.List()
		for _, item := range list {
			if err := hdl.Init(item, node); err != nil {
				klog.Error(err)
			}
		}
	}
	return nil
}

func RunGRPCServer(subscriptionManager *server.SubscriptionManager) error {
	grpcPort := 20443
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", grpcPort))
	if err != nil {
		klog.Error(err, "unable to start grpc server")
		return err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := server.New(subscriptionManager)
	pbv1.RegisterConfigControllerServer(grpcServer, s)
	klog.Infof("starting GRPC server on port %d", grpcPort)
	grpcServer.Serve(lis)
	return nil
}
