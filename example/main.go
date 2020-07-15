package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	ServiceController "github.com/ferocious-space/service-controller"
)

func init() {
	l, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(l.Named("example"))
}

func main() {
	c := ServiceController.NewServiceManager(zap.L())
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, waitctx := errgroup.WithContext(ctx)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	g.Go(func() error {
		return c.GetController().Run(waitctx)
	})

	_ = c.GetController().NewService(ServiceController.NewScheduleServiceConfig("test", 10*time.Second, nil, func(logger *zap.Logger, data interface{}) error {
		logger.Info("tick")
		return nil
	}, func(logger *zap.Logger, data interface{}) error {
		logger.Info("stop")
		return nil
	}), ServiceController.NewScheduleService())
	_ = c.GetController().NewService(ServiceController.NewServiceConfig("test2"), ServiceController.NewDefaultService())

	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	cancel()
	err := g.Wait()
	if err != nil {
		println(err.Error())
	}
}
