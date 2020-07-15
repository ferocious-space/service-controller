package ServiceController

type ServiceConfig struct {
	name string
	kv   map[string]interface{}
}

func NewServiceConfig(name string) *ServiceConfig {
	return &ServiceConfig{name: name, kv: make(map[string]interface{})}
}

func (cfg *ServiceConfig) Name() string {
	return cfg.name
}

func (cfg *ServiceConfig) Set(key string, value interface{}) *ServiceConfig {
	cfg.kv[key] = value
	return cfg
}

func (cfg *ServiceConfig) Get(key string) interface{} {
	if value, ok := cfg.kv[key]; ok {
		return value
	}
	return nil
}
