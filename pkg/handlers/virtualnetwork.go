package handlers

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/db"
	"github.com/michaelhenkel/config_controller/pkg/server"
	"github.com/michaelhenkel/config_controller/pkg/store"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

func init() {
	converterMap["VirtualNetwork"] = &VirtualNetwork{}
}

type VirtualNetwork struct {
	NewResource *contrail.VirtualNetwork
	OldResource *contrail.VirtualNetwork
	kind        string
	dbClient    *db.DB
}

func (r *VirtualNetwork) Convert(newObj interface{}, oldObj interface{}) error {
	if newObj != nil {
		r.NewResource = newObj.(*contrail.VirtualNetwork)
	}
	if oldObj != nil {
		r.OldResource = oldObj.(*contrail.VirtualNetwork)
	}
	return nil
}

func (r *VirtualNetwork) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *VirtualNetwork) addKind(kind string) {
	r.kind = kind
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

func (r *VirtualNetwork) Init() error {
	return nil
}

func (r *VirtualNetwork) InitEdges() error {
	return nil
}

func (r *VirtualNetwork) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	//klog.Infof("adding %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
	var namespacedName string
	if r.NewResource.Namespace != "" {
		namespacedName = fmt.Sprintf("%s/%s", r.NewResource.Namespace, r.NewResource.Name)
	}
	nList := r.dbClient.Get(namespacedName, r.kind, "VirtualRouter", "VirtualMachine", "VirtualMachineInterface", "VirtualNetwork", "VirtualRouter")
	for _, n := range nList {
		klog.Infof("%s -> %s is on %s -> %s", r.kind, namespacedName, n.Kind(), n.String())
	}
	return nil
}

func (r *VirtualNetwork) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.NewResource, r.OldResource) {
		klog.Infof("updating %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
		return nil
	}

	return nil
}

func (r *VirtualNetwork) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
	return nil
}

func (r *VirtualNetwork) sendResponse(subscriptionManager *server.SubscriptionManager, storeClient store.Store, action pbv1.Response_Action, vrouterFilter ...string) {
	vmiList := storeClient.ListResource("VirtualMachineInterface")
	vrouterList := storeClient.ListResource("VirtualRouter", vrouterFilter...)
	var vrouterResponseMap = make(map[string]*pbv1.Response)
	for _, vmiObj := range vmiList {
		vmi, ok := vmiObj.(*contrail.VirtualMachineInterface)
		if ok {
			if vmi.Spec.VirtualNetworkReference.Name == r.NewResource.Name && vmi.Spec.VirtualNetworkReference.Namespace == r.NewResource.Namespace {
				vmRefs := vmi.Spec.VirtualMachineReferences
				for _, vrouterObj := range vrouterList {
					vrouter, ok := vrouterObj.(*contrail.VirtualRouter)
					if ok {
						for _, vmRef := range vmRefs {
							for _, vrouterVMRef := range vrouter.Spec.VirtualMachineReferences {
								if vmRef.Name == vrouterVMRef.Name && vmRef.Namespace == vrouterVMRef.Namespace {
									objResource := &pbv1.Resource_VirtualNetwork{
										VirtualNetwork: r.NewResource,
									}
									response := &pbv1.Response{
										New: &pbv1.Resource{
											Resource: objResource,
										},
									}
									switch action {
									case pbv1.Response_UPDATE:
										objResource := &pbv1.Resource_VirtualNetwork{
											VirtualNetwork: r.OldResource,
										}
										response.Old = &pbv1.Resource{
											Resource: objResource,
										}
										response.Action = pbv1.Response_UPDATE
										fmt.Println(cmp.Diff(response.New, response.Old))
									case pbv1.Response_ADD:
										response.Action = pbv1.Response_ADD
									case pbv1.Response_DELETE:
										response.Action = pbv1.Response_DELETE
									}
									vrouterResponseMap[vrouter.Name] = response
								}
							}
						}
					}
				}
			}
		}
	}
	for vrouter, response := range vrouterResponseMap {
		if subChan := subscriptionManager.GetSubscriptionChannel(vrouter); subChan != nil {
			klog.Infof("sending vn %s to vrouter %s", response.New.GetVirtualNetwork().Name, vrouter)
			subChan <- response

		}
	}
}
