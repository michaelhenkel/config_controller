package handlers

import (
	"github.com/michaelhenkel/config_controller/pkg/db"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

func init() {
	converterMap["VirtualMachine"] = &VirtualMachine{}
}

type VirtualMachine struct {
	NewResource *contrail.VirtualMachine
	OldResource *contrail.VirtualMachine
	kind        string
	dbClient    *db.DB
}

func (r *VirtualMachine) Convert(newObj interface{}, oldObj interface{}) error {
	r.kind = "VirtualMachine"
	if newObj != nil {
		r.NewResource = newObj.(*contrail.VirtualMachine)
	}
	if oldObj != nil {
		r.OldResource = oldObj.(*contrail.VirtualMachine)
	}
	return nil
}

func (r *VirtualMachine) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *VirtualMachine) addKind(kind string) {
	r.kind = kind
}

func (r *VirtualMachine) Init() error {
	return nil
}

func (r *VirtualMachine) InitEdges() error {
	return nil
}

func (r *VirtualMachine) GetReferences(obj interface{}) []contrail.ResourceReference {
	var resourceReferenceList []contrail.ResourceReference
	return resourceReferenceList
}

func (r *VirtualMachine) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	//klog.Infof("adding %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
	return nil
}

func (r *VirtualMachine) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.NewResource, r.OldResource) {
		klog.Infof("updating %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
		return nil
	}

	return nil
}

func (r *VirtualMachine) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
	return nil
}
