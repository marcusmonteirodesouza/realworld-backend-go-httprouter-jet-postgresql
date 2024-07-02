package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve(ctx context.Context) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.logger.InfoContext(ctx, fmt.Sprintf("Caught signal %s", s.String()))

		ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)

		defer cancel()

		err := srv.Shutdown(ctxWithTimeout)
		if err != nil {
			shutdownError <- err
		}

		app.logger.InfoContext(ctx, "Completing background tasks")

		app.wg.Wait()

		shutdownError <- nil
	}()

	app.logger.InfoContext(ctx, fmt.Sprintf("Starting server on port %d", app.config.port))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.InfoContext(ctx, "Server stopped")

	return nil
}
