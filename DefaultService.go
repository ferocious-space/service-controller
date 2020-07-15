package ServiceController

import (
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type DefaultService struct {
	name        atomic.Value
	running     atomic.Value
	config      atomic.Value
	logger      atomic.Value
	lastRun     atomic.Value
	lastRestart atomic.Value
}

func NewDefaultService() *DefaultService {
	return &DefaultService{}
}

func (d *DefaultService) Init(logger *zap.Logger, config *ServiceConfig) error {
	d.name.Store(config.Name())
	d.config.Store(config)
	d.logger.Store(logger)
	d.SetState(Undefined)
	return nil
}

func (d *DefaultService) SetLastRun(when time.Time) {
	d.lastRun.Store(when)
}

func (d *DefaultService) GetLastRun() time.Time {
	if d.lastRun.Load() == nil {
		return time.Time{}
	}
	return d.lastRun.Load().(time.Time)
}

func (d *DefaultService) SetLastRestart(when time.Time) {
	d.lastRestart.Store(when)
}

func (d *DefaultService) GetLastRestart() time.Time {
	if d.lastRestart.Load() == nil {
		return time.Time{}
	}
	return d.lastRestart.Load().(time.Time)
}

func (d *DefaultService) Name() string {
	name := d.name.Load()
	if name != nil {
		return name.(string)
	}
	return ""
}

func (d *DefaultService) New() Service {
	d.Log().Debug("New()")
	return new(DefaultService)
}

func (d *DefaultService) GetState() ServiceState {
	state := d.running.Load()
	if state != nil {
		return state.(ServiceState)
	}
	d.Log().Panic("Null state")
	return Failed
}

func (d *DefaultService) SetState(state ServiceState) {
	switch state {
	case Failed:
		d.Log().Info("Setting state", zap.Any("state", state))
	}
	d.running.Store(state)
}

func (d *DefaultService) GetConfig() *ServiceConfig {
	return d.config.Load().(*ServiceConfig)
}

func (d *DefaultService) Run() error {
	return nil
}

func (d *DefaultService) Stop() error {
	d.Log().Debug("Stop()")
	return nil
}

func (d *DefaultService) Log() *zap.Logger {
	return d.logger.Load().(*zap.Logger)
}
