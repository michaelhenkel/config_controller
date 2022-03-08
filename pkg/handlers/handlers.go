package handlers

import (
	pbv1 "github.com/michaelhenkel/config_controller/pkg/apis/v1"
	"github.com/michaelhenkel/config_controller/pkg/db"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

var converterMap map[string]Handler

func init() {
	converterMap = make(map[string]Handler)
}

type Resource interface {
}

type Handler interface {
	Add(obj interface{}) error
	Update(newObj interface{}, oldObj interface{}) error
	Delete(obj interface{}) error
	GetReferences(obj interface{}) []contrail.ResourceReference
	ListResponses(node string) []pbv1.Response
	addDBClient(dbClient *db.DB)
	addKind(kind string)
}

func NewHandler(kind string, dbClient *db.DB) Handler {
	var newConverterMap = make(map[string]Handler)
	for res, handler := range converterMap {
		handler.addDBClient(dbClient)
		handler.addKind(res)
		newConverterMap[res] = handler
	}
	return newConverterMap[kind]
}

func GetHandledResources() map[string]Handler {
	return converterMap
}
