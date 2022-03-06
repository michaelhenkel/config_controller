package handlers

import (
	"fmt"

	"github.com/michaelhenkel/config_controller/pkg/db"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

func init() {
	converterMap["VirtualMachineInterface"] = &VirtualMachineInterface{}
}

type VirtualMachineInterface struct {
	NewResource *contrail.VirtualMachineInterface
	OldResource *contrail.VirtualMachineInterface
	kind        string
	dbClient    *db.DB
}

func (r *VirtualMachineInterface) Convert(newObj interface{}, oldObj interface{}) error {
	if newObj != nil {
		r.NewResource = newObj.(*contrail.VirtualMachineInterface)
	}
	if oldObj != nil {
		r.OldResource = oldObj.(*contrail.VirtualMachineInterface)
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

func (r *VirtualMachineInterface) Init() error {
	return nil
}

func (r *VirtualMachineInterface) InitEdges() error {
	return nil
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
	var namespacedName string
	if r.NewResource.Namespace != "" {
		namespacedName = fmt.Sprintf("%s/%s", r.NewResource.Namespace, r.NewResource.Name)
	}
	nList := r.dbClient.Get(namespacedName, r.kind, "VirtualRouter", "VirtualMachine", "VirtualRouter")
	for _, n := range nList {
		klog.Infof("%s -> %s is on %s -> %s", r.kind, namespacedName, n.Kind(), n.String())
	}
	return nil
}

func (r *VirtualMachineInterface) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.NewResource, r.OldResource) {
		klog.Infof("updating %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
		return nil
	}

	return nil
}

func (r *VirtualMachineInterface) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
	return nil
}
