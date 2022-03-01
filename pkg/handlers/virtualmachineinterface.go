package handlers

import (
	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/server"
	"github.com/michaelhenkel/config_controller/pkg/store"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

func init() {
	converterMap["VirtualMachineInterface"] = &VirtualMachineInterface{}
}

type VirtualMachineInterface struct {
	Resource *contrail.VirtualMachineInterface
}

func (r *VirtualMachineInterface) Convert(obj interface{}) error {
	r.Resource = obj.(*contrail.VirtualMachineInterface)
	klog.Infof("converted VirtualMachineInterface %s", r.Resource.Name)
	return nil
}

func (r *VirtualMachineInterface) Init(subscriptionManager *server.SubscriptionManager, node string) error {
	subChan := subscriptionManager.GetSubscriptionChannel(node)
	objResource := pbv1.Resource_VirtualMachineInterface{
		VirtualMachineInterface: r.Resource,
	}
	resource := &pbv1.Resource{
		Resource: &objResource,
	}
	subChan <- resource
	return nil
}

func (r *VirtualMachineInterface) Add(subscriptionManager *server.SubscriptionManager, storeClient store.Store) error {
	var vrouterMap = make(map[string]*pbv1.Resource)
	for _, vmRef := range r.Resource.Spec.VirtualMachineReferences {
		vrouterList := storeClient.ListResource("virtualrouters")
		for _, vrouterObj := range vrouterList {
			vrouter, ok := vrouterObj.(*contrail.VirtualRouter)
			if ok {
				for _, vrouterVMRef := range vrouter.Spec.VirtualMachineReferences {
					if vmRef.Name == vrouterVMRef.Name && vmRef.Namespace == vrouterVMRef.Namespace {
						objResource := pbv1.Resource_VirtualMachineInterface{
							VirtualMachineInterface: r.Resource,
						}
						resource := &pbv1.Resource{
							Resource: &objResource,
						}
						if subChan := subscriptionManager.GetSubscriptionChannel(vrouter.Name); subChan != nil {
							vrouterMap[vrouter.Name] = resource
						}
					}
				}

			}
		}
	}
	for vrouter, resource := range vrouterMap {
		klog.Infof("sending vn %s to vrouter %s", resource.GetVirtualNetwork().Name, vrouter)
		subscriptionManager.Subscriptions[vrouter].Channel <- resource
	}
	klog.Infof("adding VMI %s", r.Resource.Name)

	return nil
}

func (r *VirtualMachineInterface) Update(subscriptionManager *server.SubscriptionManager, storeClient store.Store) error {
	var vrouterMap = make(map[string]*pbv1.Resource)
	for _, vmRef := range r.Resource.Spec.VirtualMachineReferences {
		vrouterList := storeClient.ListResource("virtualrouters")
		for _, vrouterObj := range vrouterList {
			vrouter, ok := vrouterObj.(*contrail.VirtualRouter)
			if ok {
				for _, vrouterVMRef := range vrouter.Spec.VirtualMachineReferences {
					if vmRef.Name == vrouterVMRef.Name && vmRef.Namespace == vrouterVMRef.Namespace {
						objResource := pbv1.Resource_VirtualMachineInterface{
							VirtualMachineInterface: r.Resource,
						}
						resource := &pbv1.Resource{
							Resource: &objResource,
						}
						if subChan := subscriptionManager.GetSubscriptionChannel(vrouter.Name); subChan != nil {
							vrouterMap[vrouter.Name] = resource
						}
					}
				}

			}
		}
	}
	for vrouter, resource := range vrouterMap {
		klog.Infof("sending vn %s to vrouter %s", resource.GetVirtualMachineInterface().Name, vrouter)
		subscriptionManager.Subscriptions[vrouter].Channel <- resource
	}
	klog.Infof("updating VMI %s", r.Resource.Name)

	return nil
}

func (r *VirtualMachineInterface) Delete() {

}
