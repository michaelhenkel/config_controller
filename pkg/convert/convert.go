package convert

import (
	"github.com/michaelhenkel/config_controller/pkg/handlers"
	contrail "ssd-git.juniper.net/contrail/cn2/contrail/pkg/apis/core/v1alpha1"
)

func GetReferences(resource string, obj interface{}) []contrail.ResourceReference {
	resourceHandler := handlers.GetHandledResources()[resource]
	return resourceHandler.GetReferences(obj)
}
