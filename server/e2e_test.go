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

func NewTestClient(address string, clientID string) (*TestClient, error) {
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://" + address).
		SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &TestClient{
		client: client,
	}, nil
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
		client1, err := NewTestClient("localhost:"+opts.MQTTPort, "test-client-1")
		require.NoError(t, err)
		defer client1.Close()

		// クライアント2が接続できる
		client2, err := NewTestClient("localhost:"+opts.MQTTPort, "test-client-2")
		require.NoError(t, err)
		defer client2.Close()

		// サーバーを正常に終了できる
		cancel()
		err = <-errCh
		require.NoError(t, err)
	})
}
