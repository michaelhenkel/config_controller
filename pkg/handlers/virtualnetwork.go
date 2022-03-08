package handlers

import (
	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/db"
	"github.com/michaelhenkel/config_controller/pkg/graph"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

func init() {
	converterMap["VirtualNetwork"] = &VirtualNetwork{}
}

type VirtualNetwork struct {
	*contrail.VirtualNetwork
	old      *contrail.VirtualNetwork
	kind     string
	dbClient *db.DB
}

func (r *VirtualNetwork) FindFromNode(node string) []pbv1.Response {
	var responses []pbv1.Response
	virtualNetwork := NewVirtualNetwork(r.dbClient)
	virtualNetworkList := virtualNetwork.Search(graph.Node{
		Kind: "VirtualRouter",
		Name: node,
	}, &graph.Node{
		Kind: "VirtualNetwork",
	}, []string{"VirtualMachine", "VirtualMachineInterface", "VirtualNetwork"})
	for _, vn := range virtualNetworkList {
		response := &pbv1.Response{
			New: &pbv1.Resource{
				Resource: &pbv1.Resource_VirtualNetwork{
					VirtualNetwork: vn.VirtualNetwork,
				},
			},
		}
		responses = append(responses, *response)
	}
	return responses
}

func (r *VirtualNetwork) Convert(newObj interface{}, oldObj interface{}) error {
	if newObj != nil {
		r.VirtualNetwork = newObj.(*contrail.VirtualNetwork)
	}
	if oldObj != nil {
		r.old = oldObj.(*contrail.VirtualNetwork)
	}
	return nil
}

func (r *VirtualNetwork) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *VirtualNetwork) addKind(kind string) {
	r.kind = kind
}

func NewVirtualNetwork(dbClient *db.DB) *VirtualNetwork {
	return &VirtualNetwork{
		dbClient: dbClient,
		kind:     "VirtualNetwork",
	}
}

func (r *VirtualNetwork) Search(from graph.Node, to *graph.Node, path []string) map[string]*VirtualNetwork {
	var resMap = make(map[string]*VirtualNetwork)
	nodeList := r.dbClient.Search(from, to, path)
	for idx := range nodeList {
		n := r.dbClient.Get(r.kind, nodeList[idx].Name, nodeList[idx].Namespace)
		if r, ok := n.(*contrail.VirtualNetwork); ok {
			resMap[r.Namespace+"/"+r.Name] = &VirtualNetwork{VirtualNetwork: r}
		}
	}
	return resMap
}

func (r *VirtualNetwork) GetReferences(obj interface{}) []contrail.ResourceReference {
	var resourceReferenceList []contrail.ResourceReference
	res, ok := obj.(contrail.VirtualNetwork)
	if ok {
		if res.Spec.ProviderNetworkReference != nil {
			resourceReferenceList = append(resourceReferenceList, *res.Spec.ProviderNetworkReference)
		}
		if res.Spec.V4SubnetReference != nil {
			resourceReferenceList = append(resourceReferenceList, *res.Spec.V4SubnetReference)
		}
		if res.Spec.V6SubnetReference != nil {
			resourceReferenceList = append(resourceReferenceList, *res.Spec.V6SubnetReference)
		}
	}
	return resourceReferenceList
}

func (r *VirtualNetwork) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	virtualRouter := NewVirtualRouter(r.dbClient)
	virtualRouterList := virtualRouter.Search(r.Name, r.Namespace, r.kind, []string{"VirtualMachine", "VirtualMachineInterface", "VirtualRouter"})
	for _, vr := range virtualRouterList {
		klog.Infof("%s %s/%s -> %s %s/%s", r.kind, r.Namespace, r.Name, vr.GetObjectKind().GroupVersionKind().Kind, vr.Namespace, vr.Name)
	}
	return nil
}

func (r *VirtualNetwork) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.VirtualNetwork, r.old) {
		klog.Infof("updating %s: %s/%s", r.kind, r.Namespace, r.Name)
		return nil
	}

	return nil
}

func (r *VirtualNetwork) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.Namespace, r.Name)
	return nil
}
