package shutdowner

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const gracefulShutdownDelay = 5 * time.Second

var (
	onceInstance sync.Once
	onceShutdown sync.Once
)

// shutdowner gracefully shuts down the application without interrupting any application components that have been set.
//
// Set callback functions for application components (like DB) that will be invoked during graceful shutdown.
//
// Example:
// shutdowner.Get().AddCloser(func(ctx context.Context) error {
//     if err := srv.Shutdown(ctx); err != nil {
//         return err
//     }
//     return nil
// })
//
// Listen to channel shutdowner.Get().ChShutdowned to make it work properly.
//
// Example:
// chErr := app.Run()
// select {
// case <-shutdowner.Get().ChShutdowned:
//     return nil
// case err := <-chErr:
//     return fmt.Errorf("failed to start app: %w", err)
// }
type shutdowner struct {
	mu           sync.RWMutex
	callbacks    []func(context.Context) error
	chAllClosed  chan struct{}
	ChShutdowned chan struct{}
}

var instance *shutdowner

func New() {
	onceInstance.Do(func() {
		s := &shutdowner{
			chAllClosed:  make(chan struct{}),
			ChShutdowned: make(chan struct{}),
		}
		go s.catchSignalsAndShutdown()
		instance = s
	})
}

func Get() *shutdowner {
	if instance == nil {
		New()
	}
	return instance
}

func (s *shutdowner) AddCloser(fn func(ctx context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callbacks = append(s.callbacks, fn)
}

func (s *shutdowner) gracefulShutdown() {
	onceShutdown.Do(func() {
		s.mu.RLock()
		defer s.mu.RUnlock()

		ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownDelay)
		defer cancel()

		errs := make(chan error, len(s.callbacks))
		for len(s.callbacks) > 0 {
			lastIndex := len(s.callbacks) - 1
			errs <- s.callbacks[lastIndex](ctx)
			s.callbacks = s.callbacks[:lastIndex]
		}

		close(s.chAllClosed)
	})
}

func (s *shutdowner) catchSignalsAndShutdown() {
	chStopSignalReceived := make(chan os.Signal, 1)
	signal.Notify(chStopSignalReceived, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-chStopSignalReceived

	go s.gracefulShutdown()

	select {
	case <-s.chAllClosed:
		close(s.ChShutdowned)
		return
	case <-time.After(2 * gracefulShutdownDelay):
		zap.L().Warn("graceful shutdown: no response, exit with error")
		os.Exit(int(syscall.SIGTERM))
	}
}
