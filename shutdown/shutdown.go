package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/indiependente/pkg/logger"
	"golang.org/x/sync/errgroup"
)

// TerminationFn is a callback invoked on context cancellation.
type TerminationFn func(context.Context) error

// Wait allows the service to wait for a termination signal, start the cancellation process by calling
// the context.CancelFunc in order to perform a graceful service shutdown executing the TerminationFn in input.
func Wait(ctx context.Context, cancel context.CancelFunc, termFn TerminationFn) error {
	var (
		gracefulStop = make(chan os.Signal, 1)
		eg           errgroup.Group
	)

	// Get notified for incoming signals
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	// Start termination goroutine
	eg.Go(func() error {
		<-ctx.Done() // Wait for context cancellation
		return termFn(ctx)
	})

	// Wait for signal
	<-gracefulStop

	// Propagate context cancelling
	cancel()

	// Wait for cancellation propagation and termination operations to stop
	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("could not terminate gracefully: %w", err)
	}

	return nil
}

// WaitWithLogger is similar to Wait but it logs on status updates.
func WaitWithLogger(ctx context.Context, cancel context.CancelFunc, termFn TerminationFn, logger logger.Logger) error {
	var (
		gracefulStop = make(chan os.Signal, 1)
		eg           errgroup.Group
	)

	// Get notified for incoming signals
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	// Start termination goroutine
	eg.Go(func() error {
		<-ctx.Done() // Wait for context cancellation
		return termFn(ctx)
	})

	// Wait for signal
	sig := <-gracefulStop
	logger.Event("shutdown").Signal(sig).Info("Starting graceful shutdown process")
	logger.Event("shutdown").Signal(sig).Info("Propagating cancellation")
	// Propagate context cancelling
	cancel()
	logger.Event("shutdown").Signal(sig).Info("Cancellation propagated")

	// Wait for cancellation propagation and termination operations to stop
	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("could not terminate gracefully: %w", err)
	}
	logger.Event("shutdown").Signal(sig).Info("Shutdown process complete")
	return nil
}
