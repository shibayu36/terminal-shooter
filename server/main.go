package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shibayu36/terminal-shooter/server/game"
)

func main() {
	options := &runOptions{
		MQTTPort:    "1883",
		MetricsPort: "2112",
	}
	if err := run(options); err != nil {
		slog.Error(fmt.Sprintf("failed to run\n%+v", err))
		os.Exit(1)
	}
}

type runOptions struct {
	MQTTPort    string
	MetricsPort string
}

func run(opts *runOptions) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	broker := NewBroker()

	gameState := game.NewGame(30, 30)
	controller := NewController(broker, gameState)

	server, err := NewServer(":"+opts.MQTTPort, controller)
	if err != nil {
		return err
	}

	go func() {
		if err := server.Serve(); err != nil {
			panic(err)
		}
	}()

	updatedCh := gameState.StartUpdateLoop(ctx)
	controller.StartPublishLoop(ctx, updatedCh)

	// Prometheusメトリクスサーバーの起動
	//nolint:exhaustruct,gosec
	metricsServer := &http.Server{
		Addr:    ":" + opts.MetricsPort,
		Handler: promhttp.Handler(),
	}
	go func() {
		err := metricsServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	// サーバーが中断されるまで実行
	<-ctx.Done()

	if err := server.Shutdown(10 * time.Second); err != nil {
		return err
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := metricsServer.Shutdown(ctx); err != nil {
			panic(err)
		}
	}

	return nil
}
