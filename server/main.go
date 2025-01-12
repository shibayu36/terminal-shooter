package main

import (
	"context"
	"log/slog"
	"math/rand"
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
	if err := run(); err != nil {
		slog.Error("failed to run", "error", err)
		os.Exit(1)
	}
}

//nolint:funlen
func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	broker := NewBroker()

	gameState := game.NewGame(30, 30)
	controller := NewController(broker, gameState)
	server, err := NewServer(":1883", controller)
	if err != nil {
		return err
	}

	go func() {
		if err := server.Serve(); err != nil {
			panic(err)
		}
	}()

	itemsUpdatedCh := gameState.StartUpdateLoop(ctx)
	controller.StartPublishLoop(ctx, itemsUpdatedCh)

	ticker := time.NewTicker(1234 * time.Millisecond)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			directions := []game.Direction{
				game.DirectionUp,
				game.DirectionDown,
				game.DirectionLeft,
				game.DirectionRight,
			}
			//nolint:gosec
			gameState.AddBullet(
				game.Position{X: rand.Intn(30), Y: rand.Intn(30)},
				directions[rand.Intn(4)],
			)
		}
	}()

	// Prometheusメトリクスサーバーの起動
	metricsServer := &http.Server{
		Addr:    ":2112",
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
