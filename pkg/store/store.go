package store

import "k8s.io/client-go/tools/cache"

type Store map[string]cache.Store

func New() Store {
	s := make(Store)
	return s
}

func (s Store) Add(resource string, store cache.Store) {
	s[resource] = store
}

type NamespacedName struct {
	Name      string
	Namespace string
}

func (s Store) ListResource(resource string, filter ...string) []interface{} {
	if len(filter) == 0 {
		return s[resource].List()
	} else {
		var resourceList []interface{}
		for _, f := range filter {
			item, exists, _ := s[resource].GetByKey(f)
			if exists {
				resourceList = append(resourceList, item)
			}
		}
		return resourceList
	}
}

func (s Store) List() []interface{} {
	var resourceList []interface{}
	for _, store := range s {
		resourceList = append(resourceList, store.List()...)
	}
	return resourceList
}
