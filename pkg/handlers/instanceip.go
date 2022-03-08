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
	converterMap["InstanceIP"] = &InstanceIP{}
}

type InstanceIP struct {
	*contrail.InstanceIP
	old      *contrail.InstanceIP
	kind     string
	dbClient *db.DB
}

func (r *InstanceIP) Convert(newObj interface{}, oldObj interface{}) error {
	r.kind = "InstanceIP"
	if newObj != nil {
		r.InstanceIP = newObj.(*contrail.InstanceIP)
	}
	if oldObj != nil {
		r.InstanceIP = oldObj.(*contrail.InstanceIP)
	}
	return nil
}

func (r *InstanceIP) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *InstanceIP) addKind(kind string) {
	r.kind = kind
}

func (r *InstanceIP) GetReferences(obj interface{}) []contrail.ResourceReference {
	var resourceReferenceList []contrail.ResourceReference
	return resourceReferenceList
}

func (r *InstanceIP) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	return nil
}

func NewInstanceIP(dbClient *db.DB) *InstanceIP {
	return &InstanceIP{
		dbClient: dbClient,
		kind:     "InstanceIP",
	}
}

func (r *InstanceIP) FindFromNode(node string) []pbv1.Response {
	var responses []pbv1.Response
	resource := NewInstanceIP(r.dbClient)
	resourceList := resource.Search(node, "", "VirtualRouter", []string{"VirtualMachineInterface", "VirtualMachine", "InstanceIP"})
	for _, res := range resourceList {
		response := &pbv1.Response{
			New: &pbv1.Resource{
				Resource: &pbv1.Resource_InstanceIP{
					InstanceIP: res.InstanceIP,
				},
			},
		}
		responses = append(responses, *response)
	}
	return responses
}

func (r *InstanceIP) Search(name, namespace, kind string, path []string) []*InstanceIP {
	var resList []*InstanceIP

	nodeList := r.dbClient.Search(&graph.Node{
		Name:      name,
		Namespace: namespace,
		Kind:      kind,
	},
		&graph.Node{
			Kind: r.kind,
		}, path)

	for idx := range nodeList {
		n := r.dbClient.Get("InstanceIP", nodeList[idx].Name)
		if r, ok := n.(*contrail.InstanceIP); ok {
			resource := &InstanceIP{InstanceIP: r}
			resList = append(resList, resource)
		}
	}
	return resList
}

func (r *InstanceIP) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.InstanceIP, r.old) {
		klog.Infof("updating %s: %s/%s", r.kind, r.Namespace, r.Name)
		return nil
	}

	return nil
}

func (r *InstanceIP) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.Namespace, r.Name)
	return nil
}
