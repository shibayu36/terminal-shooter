package main

import (
	"context"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
)

type TestClient struct {
	client mqtt.Client
}

func NewTestClient(t *testing.T, address string, clientID string) *TestClient {
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://" + address).
		SetClientID(clientID)

	c := &TestClient{
		client: mqtt.NewClient(opts),
	}

	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("failed to connect to server: %v", token.Error())
	}

	if token := c.client.Subscribe("#", 0, c.Callback); token.Wait() && token.Error() != nil {
		t.Fatalf("failed to subscribe to server: %v", token.Error())
	}

	return c
}

func (c *TestClient) Callback(client mqtt.Client, message mqtt.Message) {
	return
}

func (c *TestClient) Close() error {
	c.client.Disconnect(250)
	return nil
}

func TestE2E(t *testing.T) {
	opts := &runOptions{
		MQTTPort:    "11883",
		MetricsPort: "12113",
	}

	t.Run("クライアントが接続でき、サーバーを終了できる", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// サーバー起動
		errCh := make(chan error)
		go func() {
			errCh <- run(ctx, opts)
		}()

		// サーバーの起動を待つ
		time.Sleep(100 * time.Millisecond)

		// クライアント1が接続できる
		NewTestClient(t, "localhost:"+opts.MQTTPort, "test-client-1")

		// クライアント2が接続できる
		NewTestClient(t, "localhost:"+opts.MQTTPort, "test-client-2")

		// サーバーを正常に終了できる
		cancel()
		err := <-errCh
		require.NoError(t, err)
	})
}
