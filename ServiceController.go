package ServiceController

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ServiceController struct {
	logger   *zap.Logger
	ticker   *time.Ticker
	services *serviceRegistry
}

type ServiceManager struct {
	l *zap.Logger
	c *ServiceController
	sync.Mutex
}

func NewServiceManager(l *zap.Logger) *ServiceManager {
	return &ServiceManager{l: l}
}

func (m *ServiceManager) GetController() *ServiceController {
	m.Lock()
	defer m.Unlock()
	if m.l == nil {
		panic("Logger not initialized.")
	}
	if m.c == nil {
		m.c = &ServiceController{
			ticker:   time.NewTicker(30 * time.Millisecond),
			services: newServiceRegistry(),
		}
		m.c.logger = m.l.Named("svcManager")
		if err := m.c.NewService(NewServiceConfig("default"), NewDefaultService()); err != nil {
			m.c.logger.Panic("default service creation failed", zap.Error(err), zap.Stack("trace"))
		}
	}
	return m.c
}

func (s *ServiceController) Log() *zap.Logger {
	return s.logger
}

func (s *ServiceController) GetService(name string) Service {
	return s.services.getService(name)
}

func (s *ServiceController) RestartService(name string) error {
	svc := s.GetService(name)
	if svc == nil {
		return ServiceError{
			Name:   name,
			Reason: "does not exist",
		}
	}
	newsvc := svc.New()
	cfg := svc.GetConfig()
	s.Log().Info("Restarting Service", zap.String("ServiceName", name), zap.String("ServiceType", GetType(svc).String()))
	return s.NewService(cfg, newsvc)
}

func (s *ServiceController) KillService(name string) error {
	svc := s.GetService(name)
	if svc == nil {
		return ServiceError{
			Name:   name,
			Reason: "does not exist",
		}
	}
	s.Log().Info("Stopping Service", zap.String("ServiceName", name), zap.String("ServiceType", GetType(svc).String()))
	if err := svc.Stop(); err != nil {
		return err
	}
	s.services.deleteService(name)
	// s.Lock()
	// delete(s.services, name)
	// s.Unlock()
	return nil
}

func (s *ServiceController) NewService(config *ServiceConfig, svc Service) error {
	msvc := s.GetService(config.Name())
	if msvc != nil {
		if err := s.KillService(config.Name()); err != nil {
			return err
		}
	}
	s.Log().Info("Creating service", zap.String("ServiceName", config.Name()), zap.String("ServiceType", GetType(svc).String()))
	svc.SetLastRestart(time.Now())

	// hystrix
	//

	// s.Lock()
	// s.services[config.Name()] = svc
	// s.Unlock()
	s.services.addService(config.Name(), svc)

	if err := svc.Init(s.Log().Named(config.Name()), config); err != nil {
		svc.SetState(Failed)
		return err
	}
	svc.SetState(Ready)
	return nil
}

func (s *ServiceController) Run(ctx context.Context) error {
	serviceSync := sync.WaitGroup{}
	for {
		serviceList := s.services.listServices()
		select {
		case <-ctx.Done():
			s.Log().Warn("Interrupt Received, Waiting all processes to finish.")
			serviceSync.Wait()
			s.Log().Warn("Interrupt Received, Stopping all services.")
			loopStop := sync.WaitGroup{}
			loopStop.Add(len(serviceList))
			for _, name := range serviceList {
				go func(name string) {
					defer loopStop.Done()
					_ = s.KillService(name)
				}(name)
			}
			loopStop.Wait()
			return nil
		case <-s.ticker.C:
			for _, name := range serviceList {
				svc := s.GetService(name)
				if svc == nil {
					continue
				}
				switch svc.GetState() {
				case Failed:
					serviceSync.Add(1)
					go func(name string, svc Service) {
						defer serviceSync.Done()
						if time.Since(svc.GetLastRestart()) < 5*time.Second {
							return
						}
						if err := s.RestartService(name); err != nil {
							s.Log().Error("Restart Service failed", zap.Error(err), zap.String("ServiceName", name), zap.String("ServiceType", GetType(svc).String()))
						}
					}(name, svc)
				case Ready:
					serviceSync.Add(1)
					go func(name string, svc Service) {
						defer serviceSync.Done()
						svc.SetState(Busy)
						if err := svc.Run(); err != nil {
							s.Log().Error("Run Failed", zap.Error(err), zap.String("ServiceName", name), zap.String("ServiceType", GetType(svc).String()))
							svc.SetState(Failed)
							return
						}
						svc.SetState(Ready)
					}(name, svc)
				}
			}
		}
	}
}
