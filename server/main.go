package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// サーバーが中断されるまで実行するためのシグナルチャネルを作成
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	broker := NewBroker()
	hook := NewController(broker, NewGameState())
	server, err := NewServer(":1883", hook, broker)
	if err != nil {
		log.Fatal(fmt.Sprintf("%+v", err))
	}

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatal(fmt.Sprintf("%+v", err))
		}
	}()

	// サーバーが中断されるまで実行
	<-done

	if err := server.Shutdown(10 * time.Second); err != nil {
		log.Fatal(fmt.Sprintf("%+v", err))
	}
}
