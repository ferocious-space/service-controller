package ServiceController

import (
	"time"

	"go.uber.org/zap"
)

type ScheduleService struct {
	*DefaultService
	config *ScheduleServiceConfig
}

type ScheduleServiceConfig struct {
	Name  string
	Timer time.Duration
	Data  interface{}
	FN    ScheduleHandler
	STOP  StopHandler
}

func NewScheduleServiceConfig(name string, timer time.Duration, ptr interface{}, run ScheduleHandler, stop StopHandler) *ServiceConfig {
	return NewServiceConfig(name).Set("config", &ScheduleServiceConfig{
		Name:  name,
		Timer: timer,
		Data:  ptr,
		FN:    run,
		STOP:  stop,
	})
}

type ScheduleHandler func(logger *zap.Logger, ptr interface{}) error
type StopHandler func(logger *zap.Logger, ptr interface{}) error

func NewScheduleService() *ScheduleService {
	return &ScheduleService{DefaultService: &DefaultService{}}
}

func (s *ScheduleService) Init(logger *zap.Logger, config *ServiceConfig) error {
	if err := s.DefaultService.Init(logger, config); err != nil {
		return err
	}
	logger.Info("Init", zap.Any("name", config.Name()))
	if cfg, ok := config.Get("config").(*ScheduleServiceConfig); ok {
		s.config = cfg
	} else {
		return ServiceError{
			Name:   config.Name(),
			Reason: "Missing or invalid config",
		}
	}
	defer s.DefaultService.SetState(Ready)
	return nil
}

func (s *ScheduleService) New() Service {
	return NewScheduleService()
}

func (s *ScheduleService) Restore() error {
	return s.Init(s.Log(), s.GetConfig())
}

func (s *ScheduleService) Run() error {
	if s.GetLastRun().IsZero() {
		defer s.SetLastRun(time.Now())
		return s.config.FN(s.Log(), s.config.Data)
	}
	if time.Since(s.GetLastRun()) > s.config.Timer {
		defer s.SetLastRun(time.Now())
		return s.config.FN(s.Log(), s.config.Data)
	}
	return nil
}

func (s *ScheduleService) Stop() error {
	if s.config.STOP != nil {
		return s.config.STOP(s.Log(), s.config.Data)
	}
	return nil
}
