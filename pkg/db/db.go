package db

import (
	"fmt"

	"github.com/michaelhenkel/config_controller/pkg/graph"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

type action string

const (
	add action = "add"
	del action = "del"
)

type HandlerInterface interface {
	GetReferences(obj interface{}) []contrail.ResourceReference
}

func (d *DB) AddHandlerInterface(kind string, handlerInterface HandlerInterface) {
	if d.handlerInterfaceMap == nil {
		d.handlerInterfaceMap = map[string]HandlerInterface{}
	}
	d.handlerInterfaceMap[kind] = handlerInterface
}

type control struct {
	action    action
	kind      string
	namespace string
	name      string
}

type DB struct {
	stores              map[string]cache.Store
	graph               graph.ItemGraph
	ctrlChan            chan control
	stopChan            chan struct{}
	handlerInterfaceMap map[string]HandlerInterface
}

func NewClient() *DB {
	return &DB{
		stores:   make(map[string]cache.Store),
		graph:    graph.ItemGraph{},
		ctrlChan: make(chan control),
		stopChan: make(chan struct{}),
	}
}

func (d *DB) AddStore(resource string, store cache.Store) {
	d.stores[resource] = store
}

func (d *DB) Search(from *graph.Node, to *graph.Node, filter []string) []*graph.Node {
	var nodeList []*graph.Node
	d.graph.TraverseFrom(from, to, func(n *graph.Node) {
		if n.Kind == to.Kind {
			nodeList = append(nodeList, n)
		}
	}, filter...)
	return nodeList
}

func (d *DB) Get(kind, key string) interface{} {
	item, ok, _ := d.stores[kind].GetByKey(key)
	if ok {
		return item
	}
	return nil
}

func (d *DB) Init() {
	for res, store := range d.stores {
		items := store.List()
		for _, item := range items {
			obj, ok := item.(metav1.Object)
			if ok {
				n := &graph.Node{Name: obj.GetName(), Namespace: obj.GetNamespace(), Kind: res}
				d.graph.AddNode(n)
				klog.Infof("added %s node %s/%s", res, obj.GetNamespace(), obj.GetName())
			}
		}
	}
	for res, store := range d.stores {
		items := store.List()
		for _, item := range items {
			var srcNode *graph.Node
			obj, ok := item.(metav1.Object)
			if ok {
				if srcNode, ok = d.graph.GetNode(obj.GetName(), obj.GetNamespace(), res); !ok {
					continue
				}
			}
			referenceList := d.handlerInterfaceMap[res].GetReferences(item)
			for _, ref := range referenceList {
				if dstNode, ok := d.graph.GetNode(ref.Name, ref.Namespace, ref.Kind); ok {
					d.graph.AddEdge(srcNode, dstNode)
					//if srcNode.Kind() == "VirtualMachineInterface" && srcNode.String() == "ns1/pod-ns1-7f7341b9" {
					//	klog.Infof("added edge from %s %s to %s %s", srcNode.Kind(), srcNode.String(), dstNode.Kind(), dstNode.String())
					//}
					if dstNode.Kind == "VirtualMachine" && dstNode.Name == "contrail-k8s-kubemanager-cluster1-local-pod-ns1-c2dbc0ed" {
						klog.Infof("added edge from %s %s to %s %s", srcNode.Kind, srcNode.Name, dstNode.Kind, dstNode.Name)
					}
				}
			}
		}
	}
}

func (d *DB) Start() error {
	if len(d.stores) == 0 {
		return fmt.Errorf("no stores, add them first")
	}
	go d.run()
	<-d.stopChan
	return nil
}

func (d *DB) Add(kind, namespace, name string) {
	d.ctrlChan <- control{
		action:    add,
		kind:      kind,
		namespace: namespace,
		name:      name,
	}
}

func (d *DB) run() {
	for ctrl := range d.ctrlChan {
		switch ctrl.action {
		case add:
			if _, ok := d.graph.GetNode(ctrl.name, ctrl.namespace, ctrl.kind); !ok {
				d.graph.AddNode(&graph.Node{Name: ctrl.name, Namespace: ctrl.namespace, Kind: ctrl.kind})
			}
		}
	}
}
