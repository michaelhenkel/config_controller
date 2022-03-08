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
	converterMap["VirtualMachineInterface"] = &VirtualMachineInterface{}
}

type VirtualMachineInterface struct {
	*contrail.VirtualMachineInterface
	old      *contrail.VirtualMachineInterface
	kind     string
	dbClient *db.DB
}

func (r *VirtualMachineInterface) Convert(newObj interface{}, oldObj interface{}) error {
	if newObj != nil {
		r.VirtualMachineInterface = newObj.(*contrail.VirtualMachineInterface)
	}
	if oldObj != nil {
		r.old = oldObj.(*contrail.VirtualMachineInterface)
	}
	r.kind = "VirtualMachineInterface"
	return nil
}

func (r *VirtualMachineInterface) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *VirtualMachineInterface) addKind(kind string) {
	r.kind = kind
}

func NewVirtualMachineInterface(dbClient *db.DB) *VirtualMachineInterface {
	return &VirtualMachineInterface{
		dbClient: dbClient,
		kind:     "VirtualMachineInterface",
	}
}

func (r *VirtualMachineInterface) Search(name, namespace, kind string, path []string) []*VirtualMachineInterface {
	var resList []*VirtualMachineInterface
	nodeList := r.dbClient.Search(&graph.Node{
		Name:      name,
		Namespace: namespace,
		Kind:      kind,
	},
		&graph.Node{
			Kind: r.kind,
		}, path)

	for idx := range nodeList {
		n := r.dbClient.Get("VirtualMachineInterface", nodeList[idx].Name)
		if r, ok := n.(*contrail.VirtualMachineInterface); ok {
			virtualMachineInterface := &VirtualMachineInterface{VirtualMachineInterface: r}
			resList = append(resList, virtualMachineInterface)
		}
	}
	return resList
}

func (r *VirtualMachineInterface) FindFromNode(node string) []pbv1.Response {
	var responses []pbv1.Response
	virtualMachineInterface := NewVirtualMachineInterface(r.dbClient)
	virtualMachineInterfaceList := virtualMachineInterface.Search(node, "", "VirtualRouter", []string{"VirtualMachine", "VirtualMachineInterface"})
	for _, vmi := range virtualMachineInterfaceList {
		response := &pbv1.Response{
			New: &pbv1.Resource{
				Resource: &pbv1.Resource_VirtualMachineInterface{
					VirtualMachineInterface: vmi.VirtualMachineInterface,
				},
			},
		}
		responses = append(responses, *response)
	}
	return responses
}

func (r *VirtualMachineInterface) GetReferences(obj interface{}) []contrail.ResourceReference {
	var resourceReferenceList []contrail.ResourceReference
	res, ok := obj.(*contrail.VirtualMachineInterface)
	if ok {
		if res.Spec.TagReferences != nil {
			resourceReferenceList = append(resourceReferenceList, res.Spec.TagReferences...)
		}
		if res.Spec.VirtualMachineInterfaceReferences != nil {
			resourceReferenceList = append(resourceReferenceList, res.Spec.VirtualMachineInterfaceReferences...)
		}
		if res.Spec.VirtualMachineReferences != nil {
			resourceReferenceList = append(resourceReferenceList, res.Spec.VirtualMachineReferences...)
		}
		if res.Spec.VirtualNetworkReference != nil {
			resourceReferenceList = append(resourceReferenceList, *res.Spec.VirtualNetworkReference)
		}
	}
	return resourceReferenceList
}

func (r *VirtualMachineInterface) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	virtualRouter := NewVirtualRouter(r.dbClient)
	virtualRouterList := virtualRouter.Search(r.Name, r.Namespace, r.kind, []string{"VirtualMachine", "VirtualRouter"})
	for _, vr := range virtualRouterList {
		klog.Infof("%s %s/%s -> %s %s/%s", r.kind, r.Namespace, r.Name, vr.GetObjectKind().GroupVersionKind().Kind, vr.Namespace, vr.Name)
	}
	return nil
}

func (r *VirtualMachineInterface) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.VirtualMachineInterface, r.old) {
		klog.Infof("updating %s: %s/%s", r.kind, r.Namespace, r.Name)
		return nil
	}

	return nil
}

func (r *VirtualMachineInterface) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.Namespace, r.Name)
	return nil
}
