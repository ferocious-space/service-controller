package ServiceController

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type ServiceError struct {
	Name   string
	Reason string
}

func (e ServiceError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("ServiceError (%s) : %s", e.Name, e.Reason)
	}
	return fmt.Sprintf("ServiceError %s", e.Reason)
}

type ServiceState int32

const (
	_ ServiceState = iota
	Undefined
	Ready
	Busy
	Failed
)

func (s ServiceState) String() string {
	return [...]string{"Zero", "Undefined", "Ready", "Busy", "Failed"}[s]
}

type Service interface {
	Name() string
	Init(logger *zap.Logger, config *ServiceConfig) error
	New() Service
	GetState() ServiceState
	SetState(state ServiceState)
	GetConfig() *ServiceConfig
	Run() error
	Stop() error
	Log() *zap.Logger
	SetLastRun(when time.Time)
	GetLastRun() time.Time
	SetLastRestart(when time.Time)
	GetLastRestart() time.Time
}
