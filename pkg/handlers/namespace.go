package handlers

import (
	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/db"
	"github.com/michaelhenkel/config_controller/pkg/graph"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog/v2"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

func init() {
	converterMap["Namespace"] = &Namespace{}
}

type Namespace struct {
	*corev1.Namespace
	old      *corev1.Namespace
	kind     string
	dbClient *db.DB
}

func (r *Namespace) Convert(newObj interface{}, oldObj interface{}) error {
	if newObj != nil {
		r.Namespace = newObj.(*corev1.Namespace)
	}
	if oldObj != nil {
		r.old = oldObj.(*corev1.Namespace)
	}
	r.kind = "Namespace"
	return nil
}

func NewNamespace(dbClient *db.DB) *Namespace {
	return &Namespace{
		dbClient: dbClient,
		kind:     "Namespace",
	}
}

func (r *Namespace) addDBClient(dbClient *db.DB) {
	r.dbClient = dbClient
}

func (r *Namespace) addKind(kind string) {
	r.kind = kind
}

func (r *Namespace) GetReferences(obj interface{}) []contrail.ResourceReference {
	var resourceReferenceList []contrail.ResourceReference
	return resourceReferenceList
}

func (r *Namespace) Add(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	return nil
}

func (r *Namespace) Update(newObj interface{}, oldObj interface{}) error {
	if err := r.Convert(newObj, oldObj); err != nil {
		return err
	}

	if !equality.Semantic.DeepDerivative(r.Namespace, r.old) {
		klog.Infof("updating %s: %s/%s", r.kind, r.Namespace, r.Name)
		return nil
	}

	return nil
}

func (r *Namespace) Delete(obj interface{}) error {
	if err := r.Convert(obj, nil); err != nil {
		return err
	}
	klog.Infof("deleting %s: %s/%s", r.kind, r.Namespace, r.Name)
	return nil
}

func (r *Namespace) Search(name, namespace, kind string, path []string) []*Namespace {
	var resList []*Namespace
	nodeList := r.dbClient.Search(&graph.Node{
		Name:      name,
		Namespace: namespace,
		Kind:      kind,
	},
		&graph.Node{
			Kind: r.kind,
		}, path)

	for idx := range nodeList {
		n := r.dbClient.Get(r.kind, nodeList[idx].Name)
		if r, ok := n.(*corev1.Namespace); ok {
			virtualRouter := &Namespace{Namespace: r}
			resList = append(resList, virtualRouter)
		}
	}
	return resList
}

func (r *Namespace) FindFromNode(node string) []pbv1.Response {
	var responses []pbv1.Response
	/*
		namespace := NewNamespace(r.dbClient)
		namespaceList := namespace.Search(node, "", "Namespace", []string{})
		for _, ns := range namespaceList {
			response := &pbv1.Response{
				New: &pbv1.Resource{
					Resource: &pbv1.Resource_Namespace{
						Namespace: ns.Namespace,
					},
				},
			}
			responses = append(responses, *response)
		}
	*/
	return responses
}
