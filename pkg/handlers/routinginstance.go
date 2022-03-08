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
	converterMap["RoutingInstance"] = &RoutingInstance{}
}

type RoutingInstance struct {
	*contrail.RoutingInstance
	old      *contrail.RoutingInstance
	kind     string
	dbClient *db.DB
}

func (r *RoutingInstance) Convert(newObj interface{}, oldObj interface{}) error {
	r.kind = "RoutingInstance"
	if newObj != nil {
		r.RoutingInstance = newObj.(*contrail.RoutingInstance)
	}
	if oldObj != nil {
		r.RoutingInstance = oldObj.(*contrail.RoutingInstance)
	}
	return nil
}

func (r *RoutingInstance) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *RoutingInstance) addKind(kind string) {
	r.kind = kind
}

func (r *RoutingInstance) GetReferences(obj interface{}) []contrail.ResourceReference {
	var resourceReferenceList []contrail.ResourceReference
	return resourceReferenceList
}

func (r *RoutingInstance) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	return nil
}

func NewRoutingInstance(dbClient *db.DB) *RoutingInstance {
	return &RoutingInstance{
		dbClient: dbClient,
		kind:     "RoutingInstance",
	}
}

func (r *RoutingInstance) FindFromNode(node string) []pbv1.Response {
	var responses []pbv1.Response
	virtualMachine := NewRoutingInstance(r.dbClient)
	virtualMachineList := virtualMachine.Search(node, "", "VirtualRouter", []string{"RoutingInstance", "RoutingInstanceInterface"})
	for _, vm := range virtualMachineList {
		response := &pbv1.Response{
			New: &pbv1.Resource{
				Resource: &pbv1.Resource_RoutingInstance{
					RoutingInstance: vm.RoutingInstance,
				},
			},
		}
		responses = append(responses, *response)
	}
	return responses
}

func (r *RoutingInstance) Search(name, namespace, kind string, path []string) []*RoutingInstance {
	var resList []*RoutingInstance

	nodeList := r.dbClient.Search(&graph.Node{
		Name:      name,
		Namespace: namespace,
		Kind:      kind,
	},
		&graph.Node{
			Kind: r.kind,
		}, path)

	for idx := range nodeList {
		n := r.dbClient.Get("RoutingInstance", nodeList[idx].Name)
		if r, ok := n.(*contrail.RoutingInstance); ok {
			resource := &RoutingInstance{RoutingInstance: r}
			resList = append(resList, resource)
		}
	}
	return resList
}

func (r *RoutingInstance) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.RoutingInstance, r.old) {
		klog.Infof("updating %s: %s/%s", r.kind, r.Namespace, r.Name)
		return nil
	}

	return nil
}

func (r *RoutingInstance) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.Namespace, r.Name)
	return nil
}
