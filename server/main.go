package main

import (
	"context"
	"log"
	"math/rand"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	broker := NewBroker()

	gameState := NewGameState(30, 30)
	hook := NewController(broker, gameState)
	server, err := NewServer(":1883", hook)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("%+v", err)
		}
	}()

	itemsUpdatedCh := gameState.StartUpdateLoop(ctx)
	hook.StartPublishLoop(ctx, itemsUpdatedCh)

	ticker := time.NewTicker(1234 * time.Millisecond)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			directions := []Direction{DirectionUp, DirectionDown, DirectionLeft, DirectionRight}
			//nolint:gosec
			gameState.AddBullet(&Position{X: rand.Intn(30), Y: rand.Intn(30)}, directions[rand.Intn(4)])
		}
	}()

	// サーバーが中断されるまで実行
	<-ctx.Done()

	if err := server.Shutdown(10 * time.Second); err != nil {
		log.Fatalf("%+v", err)
	}
}
