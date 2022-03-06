package handlers

import (
	"github.com/michaelhenkel/config_controller/pkg/db"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

var converterMap map[string]Handler

func init() {
	converterMap = make(map[string]Handler)
}

type Handler interface {
	Add(obj interface{}) error
	Update(newObj interface{}, oldObj interface{}) error
	Delete(obj interface{}) error
	Init() error
	InitEdges() error
	GetReferences(obj interface{}) []contrail.ResourceReference
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
