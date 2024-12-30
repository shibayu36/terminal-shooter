package main

import (
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

	hook := NewHook(&HookOptions{game: NewGameState()})

	server, err := NewServer(":1883", hook)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatal(err)
		}
	}()

	// サーバーが中断されるまで実行
	<-done

	if err := server.Shutdown(10 * time.Second); err != nil {
		log.Fatal(err)
	}

	// server.Log.Warn("caught signal, stopping...")
	// _ = server.Close()
	// server.Log.Info("main.go finished")
}
