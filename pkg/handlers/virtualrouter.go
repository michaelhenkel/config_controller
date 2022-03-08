package handlers

import (
	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/db"
	"github.com/michaelhenkel/config_controller/pkg/graph"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

func init() {
	converterMap["VirtualRouter"] = &VirtualRouter{}
}

type VirtualRouter struct {
	*contrail.VirtualRouter
	old      *contrail.VirtualRouter
	kind     string
	dbClient *db.DB
}

func (r *VirtualRouter) Convert(newObj interface{}, oldObj interface{}) error {
	if newObj != nil {
		r.VirtualRouter = newObj.(*contrail.VirtualRouter)
	}
	if oldObj != nil {
		r.old = oldObj.(*contrail.VirtualRouter)
	}
	r.kind = "VirtualRouter"
	return nil
}

func NewVirtualRouter(dbClient *db.DB) *VirtualRouter {
	return &VirtualRouter{
		dbClient: dbClient,
		kind:     "VirtualRouter",
	}
}

func (r *VirtualRouter) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *VirtualRouter) addKind(kind string) {
	r.kind = kind
}

func (r *VirtualRouter) GetReferences(obj interface{}) []contrail.ResourceReference {
	var resourceReferenceList []contrail.ResourceReference
	res, ok := obj.(*contrail.VirtualRouter)
	if ok {
		if res.Spec.VirtualMachineReferences != nil {
			resourceReferenceList = append(resourceReferenceList, res.Spec.VirtualMachineReferences...)
		}
	}
	return resourceReferenceList
}

func (r *VirtualRouter) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	return nil
}

func (r *VirtualRouter) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.VirtualRouter, r.old) {
		klog.Infof("updating %s: %s/%s", r.kind, r.Namespace, r.Name)
		return nil
	}

	return nil
}

func (r *VirtualRouter) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.Namespace, r.Name)
	return nil
}

func (r *VirtualRouter) Search(name, namespace, kind string, path []string) map[string]*VirtualRouter {
	var resMap = make(map[string]*VirtualRouter)
	nodeList := r.dbClient.Search(graph.Node{
		Name:      name,
		Namespace: namespace,
		Kind:      kind,
	},
		&graph.Node{
			Kind: r.kind,
		}, path)

	for idx := range nodeList {
		n := r.dbClient.Get(r.kind, nodeList[idx].Name, nodeList[idx].Namespace)
		if r, ok := n.(*contrail.VirtualRouter); ok {
			resMap[r.Namespace+"/"+r.Name] = &VirtualRouter{VirtualRouter: r}
		}
	}
	return resMap
}

func diffRefs(a []contrail.ResourceReference, b []contrail.ResourceReference) ([]contrail.ResourceReference, []contrail.ResourceReference) {
	var aMap = make(map[types.NamespacedName]contrail.ResourceReference)
	var bMap = make(map[types.NamespacedName]contrail.ResourceReference)
	var notInA, notInB []contrail.ResourceReference
	for _, ref := range a {
		aMap[ref.GetNamespacedName()] = ref
	}
	for _, ref := range b {
		bMap[ref.GetNamespacedName()] = ref
	}
	for refName, ref := range aMap {
		if _, ok := bMap[refName]; !ok {
			notInB = append(notInB, ref)
		}
	}

	for refName, ref := range bMap {
		if _, ok := aMap[refName]; !ok {
			notInA = append(notInA, ref)
		}
	}
	return notInA, notInB
}

func (r *VirtualRouter) FindFromNode(node string) []pbv1.Response {
	var responses []pbv1.Response
	virtualRouter := NewVirtualRouter(r.dbClient)
	virtualRouterList := virtualRouter.Search(node, "", "VirtualRouter", []string{})
	for _, vr := range virtualRouterList {
		response := &pbv1.Response{
			New: &pbv1.Resource{
				Resource: &pbv1.Resource_VirtualRouter{
					VirtualRouter: vr.VirtualRouter,
				},
			},
		}
		responses = append(responses, *response)
	}
	return responses
}
