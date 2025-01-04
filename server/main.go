package main

import (
	"context"
	"log"
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

	// サーバーが中断されるまで実行
	<-ctx.Done()

	if err := server.Shutdown(10 * time.Second); err != nil {
		log.Fatalf("%+v", err)
	}
}
