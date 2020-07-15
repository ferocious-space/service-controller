package ServiceController

import (
	"sync"
)

type serviceRegistry struct {
	services sync.Map
}

func newServiceRegistry() *serviceRegistry {
	return &serviceRegistry{}
}

func (r *serviceRegistry) getService(name string) Service {
	svc, ok := r.services.Load(name)
	if ok {
		return svc.(Service)
	}
	return nil
}

func (r *serviceRegistry) addService(name string, svc Service) {
	r.services.Store(name, svc)
}

func (r *serviceRegistry) deleteService(name string) {
	r.services.Delete(name)
}

func (r *serviceRegistry) listServices() []string {
	keys := make([]string, 0)
	r.services.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}
