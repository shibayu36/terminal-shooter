package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
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

	// MQTTサーバーの作成
	server := mqtt.New(&mqtt.Options{})
	// 今回は認証なし
	_ = server.AddHook(new(auth.AllowHook), nil)

	// TCPリスナーの追加
	tcp := listeners.NewTCP(listeners.Config{ID: "t1", Address: ":1883"})
	err := server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	// サーバーにフックを追加
	err = server.AddHook(new(Hook), nil)
	if err != nil {
		log.Fatal(err)
	}

	// サーバーの起動
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// サーバーが中断されるまで実行
	<-done

	server.Log.Warn("caught signal, stopping...")
	_ = server.Close()
	server.Log.Info("main.go finished")
}
