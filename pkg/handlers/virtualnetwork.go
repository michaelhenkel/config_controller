package handlers

import (
	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/server"
	"github.com/michaelhenkel/config_controller/pkg/store"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

/*
func init() {
	converterMap["VirtualNetwork"] = &VirtualNetwork{}
}
*/

type VirtualNetwork struct {
	Resource *contrail.VirtualNetwork
}

func (r *VirtualNetwork) Convert(obj interface{}) error {
	r.Resource = obj.(*contrail.VirtualNetwork)
	klog.Infof("converted virtualNetwork %s", r.Resource.Name)
	return nil
}

func (r *VirtualNetwork) Init(subscriptionManager *server.SubscriptionManager, node string) error {
	subChan := subscriptionManager.GetSubscriptionChannel(node)
	objResource := pbv1.Resource_VirtualNetwork{
		VirtualNetwork: r.Resource,
	}
	resource := &pbv1.Resource{
		Resource: &objResource,
	}
	subChan <- resource
	return nil
}

func (r *VirtualNetwork) Add(subscriptionManager *server.SubscriptionManager, storeClient store.Store) error {

	vmiList := storeClient.ListResource("virtualmachineinterfaces")
loop:
	for _, vmiObj := range vmiList {
		vmi, ok := vmiObj.(*contrail.VirtualMachineInterface)
		if ok {
			if vmi.Spec.VirtualNetworkReference.Name == r.Resource.Name && vmi.Spec.VirtualNetworkReference.Namespace == r.Resource.Namespace {
				vmRefs := vmi.Spec.VirtualMachineReferences
				vrouterList := storeClient.ListResource("virtualrouters")
				for _, vrouterObj := range vrouterList {
					vrouter, ok := vrouterObj.(*contrail.VirtualRouter)
					if ok {
						for _, vmRef := range vmRefs {
							for _, vrouterVMRef := range vrouter.Spec.VirtualMachineReferences {
								if vmRef.Name == vrouterVMRef.Name && vmRef.Namespace == vrouterVMRef.Namespace {
									objResource := pbv1.Resource_VirtualNetwork{
										VirtualNetwork: r.Resource,
									}
									resource := &pbv1.Resource{
										Resource: &objResource,
									}
									subscriptionManager.Subscriptions[vrouter.Name].Channel <- resource
									break loop
								}
							}
						}
					}
				}
			}
		}
	}

	klog.Infof("adding VN %s", r.Resource.Name)

	return nil
}

func (r *VirtualNetwork) Update(subscriptionManager *server.SubscriptionManager, storeClient store.Store) error {
	klog.Infof("updating VN %s", r.Resource.Name)
	vmiList := storeClient.ListResource("virtualmachineinterfaces")
	var vrouterMap = make(map[string]*pbv1.Resource)
	for _, vmiObj := range vmiList {
		vmi, ok := vmiObj.(*contrail.VirtualMachineInterface)
		if ok {
			if vmi.Spec.VirtualNetworkReference.Name == r.Resource.Name && vmi.Spec.VirtualNetworkReference.Namespace == r.Resource.Namespace {
				vmRefs := vmi.Spec.VirtualMachineReferences
				vrouterList := storeClient.ListResource("virtualrouters")
				for _, vrouterObj := range vrouterList {
					vrouter, ok := vrouterObj.(*contrail.VirtualRouter)
					if ok {
						for _, vmRef := range vmRefs {
							for _, vrouterVMRef := range vrouter.Spec.VirtualMachineReferences {
								if vmRef.Name == vrouterVMRef.Name && vmRef.Namespace == vrouterVMRef.Namespace {
									objResource := pbv1.Resource_VirtualNetwork{
										VirtualNetwork: r.Resource,
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
			}
		}
	}
	for vrouter, resource := range vrouterMap {
		klog.Infof("sending vn %s to vrouter %s", resource.GetVirtualNetwork().Name, vrouter)
		subscriptionManager.Subscriptions[vrouter].Channel <- resource
	}
	return nil
}

func (r *VirtualNetwork) Delete() {

}
