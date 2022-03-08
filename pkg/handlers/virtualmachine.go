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
	converterMap["VirtualMachine"] = &VirtualMachine{}
}

type VirtualMachine struct {
	*contrail.VirtualMachine
	old      *contrail.VirtualMachine
	kind     string
	dbClient *db.DB
}

func (r *VirtualMachine) Convert(newObj interface{}, oldObj interface{}) error {
	r.kind = "VirtualMachine"
	if newObj != nil {
		r.VirtualMachine = newObj.(*contrail.VirtualMachine)
	}
	if oldObj != nil {
		r.VirtualMachine = oldObj.(*contrail.VirtualMachine)
	}
	return nil
}

func (r *VirtualMachine) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *VirtualMachine) addKind(kind string) {
	r.kind = kind
}

func (r *VirtualMachine) GetReferences(obj interface{}) []contrail.ResourceReference {
	var resourceReferenceList []contrail.ResourceReference
	return resourceReferenceList
}

func (r *VirtualMachine) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	return nil
}

func NewVirtualMachine(dbClient *db.DB) *VirtualMachine {
	return &VirtualMachine{
		dbClient: dbClient,
		kind:     "VirtualMachine",
	}
}

func (r *VirtualMachine) FindFromNode(node string) []pbv1.Response {
	var responses []pbv1.Response
	virtualMachine := NewVirtualMachine(r.dbClient)
	virtualMachineList := virtualMachine.Search(node, "", "VirtualRouter", []string{"VirtualMachine", "VirtualMachineInterface"})
	for _, vm := range virtualMachineList {
		response := &pbv1.Response{
			New: &pbv1.Resource{
				Resource: &pbv1.Resource_VirtualMachine{
					VirtualMachine: vm.VirtualMachine,
				},
			},
		}
		responses = append(responses, *response)
	}
	return responses
}

func (r *VirtualMachine) Search(name, namespace, kind string, path []string) map[string]*VirtualMachine {
	var resMap = make(map[string]*VirtualMachine)
	nodeList := r.dbClient.Search(graph.Node{
		Name:      name,
		Namespace: namespace,
		Kind:      kind,
	},
		&graph.Node{
			Kind: r.kind,
		}, path)

	for idx := range nodeList {
		n := r.dbClient.Get("VirtualMachine", nodeList[idx].Name, nodeList[idx].Namespace)
		if r, ok := n.(*contrail.VirtualMachine); ok {
			resMap[r.Namespace+"/"+r.Name] = &VirtualMachine{VirtualMachine: r}
		}
	}
	return resMap
}

func (r *VirtualMachine) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.VirtualMachine, r.old) {
		klog.Infof("updating %s: %s/%s", r.kind, r.Namespace, r.Name)
		return nil
	}

	return nil
}

func (r *VirtualMachine) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.Namespace, r.Name)
	return nil
}
