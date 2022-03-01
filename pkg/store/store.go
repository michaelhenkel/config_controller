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

func (s Store) ListResource(resource string) []interface{} {
	return s[resource].List()
}

func (s Store) List() []interface{} {
	var resourceList []interface{}
	for _, store := range s {
		resourceList = append(resourceList, store.List()...)
	}
	return resourceList
}
