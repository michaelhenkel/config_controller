package handlers

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/db"
	"github.com/michaelhenkel/config_controller/pkg/server"
	"github.com/michaelhenkel/config_controller/pkg/store"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

func init() {
	converterMap["VirtualRouter"] = &VirtualRouter{}
}

type VirtualRouter struct {
	NewResource *contrail.VirtualRouter
	OldResource *contrail.VirtualRouter
	kind        string
	dbClient    *db.DB
}

func (r *VirtualRouter) Convert(newObj interface{}, oldObj interface{}) error {
	if newObj != nil {
		r.NewResource = newObj.(*contrail.VirtualRouter)
	}
	if oldObj != nil {
		r.OldResource = oldObj.(*contrail.VirtualRouter)
	}
	r.kind = "VirtualRouter"
	return nil
}

func (r *VirtualRouter) Init() error {
	return nil
}

func (r *VirtualRouter) InitEdges() error {
	return nil
}

func (r *VirtualRouter) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *VirtualRouter) addKind(kind string) {
	r.kind = kind
}

func (r *VirtualRouter) GetReferences(obj interface{}) []contrail.ResourceReference {
	var resourceReferenceList []contrail.ResourceReference
	res, ok := obj.(*contrail.VirtualRouter)
	if ok {
		if res.Spec.VirtualMachineReferences != nil {
			resourceReferenceList = append(resourceReferenceList, res.Spec.VirtualMachineReferences...)
		}
	}
	return resourceReferenceList
}

func (r *VirtualRouter) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	//klog.Infof("adding %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
	return nil
}

func (r *VirtualRouter) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.NewResource, r.OldResource) {
		klog.Infof("updating %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
		return nil
	}

	return nil
}

func (r *VirtualRouter) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.NewResource.Namespace, r.NewResource.Name)
	return nil
}

func diffRefs(a []contrail.ResourceReference, b []contrail.ResourceReference) ([]contrail.ResourceReference, []contrail.ResourceReference) {
	var aMap = make(map[types.NamespacedName]contrail.ResourceReference)
	var bMap = make(map[types.NamespacedName]contrail.ResourceReference)
	var notInA, notInB []contrail.ResourceReference
	for _, ref := range a {
		aMap[ref.GetNamespacedName()] = ref
	}
	for _, ref := range b {
		bMap[ref.GetNamespacedName()] = ref
	}
	for refName, ref := range aMap {
		if _, ok := bMap[refName]; !ok {
			notInB = append(notInB, ref)
		}
	}

	for refName, ref := range bMap {
		if _, ok := aMap[refName]; !ok {
			notInA = append(notInA, ref)
		}
	}
	return notInA, notInB
}

func (r *VirtualRouter) sendResponse(subscriptionManager *server.SubscriptionManager, storeClient store.Store, action pbv1.Response_Action, vrouterFilter ...string) {
	objResource := &pbv1.Resource_VirtualRouter{
		VirtualRouter: r.NewResource,
	}
	response := &pbv1.Response{
		New: &pbv1.Resource{
			Resource: objResource,
		},
	}
	switch action {
	case pbv1.Response_UPDATE:
		objResource := &pbv1.Resource_VirtualRouter{
			VirtualRouter: r.OldResource,
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

	if subChan := subscriptionManager.GetSubscriptionChannel(r.NewResource.Name); subChan != nil {
		klog.Infof("sending vn %s to vrouter %s", response.New.GetVirtualRouter().Name, r.NewResource.Name)
		subChan <- response

	}

}
