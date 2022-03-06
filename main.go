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
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/db"
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

type gvrKind struct {
	gvr  schema.GroupVersionResource
	kind string
}

func main() {
	var stopCh = make(chan struct{})
	var newSubscriberChan = make(chan string)
	subscriptionManager := server.NewSubscriptionManager(newSubscriberChan)
	dbClient := db.NewClient()
	go dbClient.Start()
	go ClientWatch(stopCh, subscriptionManager, dbClient)
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
func NewSharedInformerFactory(clientSet *ClientSet, dbClient *db.DB, mu *sync.RWMutex, synced *bool) (map[string]cache.SharedInformer, error) {
	resyncTimer := time.Minute * 10
	gvrMap, err := getGVRMap(clientSet)
	if err != nil {
		return nil, err
	}

	var sharedInformerMap = make(map[string]cache.SharedInformer)

	/*
		kubeFactory := informers.NewSharedInformerFactory(clientSet.Kube, resyncTimer)
		namespaceInformer := kubeFactory.Core().V1().Namespaces().Informer()
		storeClient.Add("Namespace", namespaceInformer.GetStore())
		namespaceInformer.AddEventHandler(resourceEventHandler(handlers.NewHandler("Namespace"), mu, synced))
		sharedInformerMap["Namespace"] = namespaceInformer
	*/

	handledResources := handlers.GetHandledResources()
	contrailFactory := contrailInformer.NewSharedInformerFactory(clientSet.Contrail, resyncTimer)
	for _, gvr := range gvrMap {
		if _, ok := handledResources[gvr.kind]; ok {
			cInformer, err := contrailFactory.ForResource(gvr.gvr)
			if err != nil {
				return nil, err
			}
			cInformer.Informer().AddEventHandler(resourceEventHandler(handlers.NewHandler(gvr.kind, dbClient), mu, synced))
			sharedInformerMap[gvr.kind] = cInformer.Informer()
		}
	}
	return sharedInformerMap, nil
}

func getGVRMap(clientSet *ClientSet) (map[string]gvrKind, error) {
	var gvrMap = make(map[string]gvrKind)
	contrailResources, err := clientSet.Contrail.DiscoveryClient.ServerResourcesForGroupVersion("core.contrail.juniper.net/v1alpha1")
	if err != nil {
		return nil, err
	}

	for _, contrailResource := range contrailResources.APIResources {
		resourceNameList := strings.Split(contrailResource.Name, "/status")
		gvrMap[resourceNameList[0]] = gvrKind{
			gvr: schema.GroupVersionResource{
				Group:    "core.contrail.juniper.net",
				Version:  "v1alpha1",
				Resource: resourceNameList[0],
			},
			kind: contrailResource.Kind,
		}
	}
	return gvrMap, nil
}

func resourceEventHandler(handler handlers.Handler, mux *sync.RWMutex, synced *bool) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			handler.Add(obj)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			handler.Update(newObj, oldObj)

		},
		DeleteFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			handler.Delete(obj)
		},
	}
}

func cacheSynced(syncMap map[string]bool) bool {
	for _, isSynced := range syncMap {
		if !isSynced {
			return false
		}
	}
	return true
}

func ClientWatch(stopChan chan struct{}, subscriptionManager *server.SubscriptionManager, dbClient *db.DB) error {
	klog.Info("starting client watch")
	clientSet, err := NewClientSet()
	if err != nil {
		return err
	}
	mux := &sync.RWMutex{}
	synced := false
	sharedInformerMap, err := NewSharedInformerFactory(clientSet, dbClient, mux, &synced)

	var syncMap = make(map[string]bool)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	mux.Lock()
	for resource, sharedInformer := range sharedInformerMap {
		dbClient.AddStore(resource, sharedInformer.GetStore())
		go sharedInformer.Run(ctx.Done())
		isSynced := cache.WaitForCacheSync(ctx.Done(), sharedInformer.HasSynced)
		syncMap[resource] = isSynced
	}

	if cacheSynced(syncMap) {
		for kind, handler := range handlers.GetHandledResources() {
			dbClient.AddHandlerInterface(kind, handler)
		}
		dbClient.Init()
		klog.Info("starting watch in 1 sec")
		time.Sleep(time.Second * 1)
		mux.Unlock()
	}

	for _, isSynced := range syncMap {
		if !isSynced {
			return err
		}
	}

	//go HandleNewSubscriber(store, subscriptionManager, graph)
	<-ctx.Done()
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
