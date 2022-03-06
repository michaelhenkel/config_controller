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

func (d *DB) Get(name, kind string, dstKind string, filter ...string) []*graph.Node {
	var nodeList []*graph.Node
	d.graph.TraverseFrom(graph.NewNode(name, kind), func(n *graph.Node) {
		if n.Kind() == dstKind {
			nodeList = append(nodeList, n)
		}
	}, filter...)
	return nodeList
}

func (d *DB) Init() {
	for res, store := range d.stores {
		items := store.List()
		for _, item := range items {
			obj, ok := item.(metav1.Object)
			if ok {
				var namespacedName string
				if obj.GetNamespace() != "" {
					namespacedName = fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
				} else {
					namespacedName = obj.GetName()
				}
				n := graph.NewNode(namespacedName, res)
				d.graph.AddNode(n)
				klog.Infof("added %s node %s", res, namespacedName)
			}
		}
	}
	for res, store := range d.stores {
		items := store.List()
		for _, item := range items {
			referenceList := d.handlerInterfaceMap[res].GetReferences(item)
			for _, ref := range referenceList {
				var dstNamespacedName string
				if ref.Namespace != "" {
					dstNamespacedName = fmt.Sprintf("%s/%s", ref.Namespace, ref.Name)
				} else {
					dstNamespacedName = ref.Name
				}
				if dstNode, ok := d.graph.GetNode(ref.Kind, dstNamespacedName); ok {
					obj, ok := item.(metav1.Object)
					if ok {
						var srcNamespacedName string
						if obj.GetNamespace() != "" {
							srcNamespacedName = fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
						} else {
							srcNamespacedName = obj.GetName()
						}
						if srcNode, ok := d.graph.GetNode(res, srcNamespacedName); ok {
							d.graph.AddEdge(srcNode, dstNode)
							klog.Infof("added edge from %s %s to %s %s", srcNode.Kind(), srcNode.String(), dstNode.Kind(), dstNode.String())
						}
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
			var namespacedName string
			if ctrl.namespace != "" {
				namespacedName = fmt.Sprintf("%s/%s", ctrl.namespace, ctrl.name)
			} else {
				namespacedName = ctrl.name
			}
			if _, ok := d.graph.GetNode(ctrl.kind, namespacedName); !ok {
				d.graph.AddNode(graph.NewNode(namespacedName, ctrl.kind))
			}
		}
	}
}
