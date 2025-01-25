package main

import (
	"context"
	"sync"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/shibayu36/terminal-shooter/shared"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type TestClient struct {
	client   mqtt.Client
	messages []mqtt.Message
	mu       sync.Mutex `exhaustruct:"optional"`
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
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messages = append(c.messages, message)
}

func (c *TestClient) Messages() []mqtt.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	messages := make([]mqtt.Message, len(c.messages))
	copy(messages, c.messages)

	return messages
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

	t.Run("プレイヤーが動くと、他のプレイヤーに通知される", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// サーバー起動
		errCh := make(chan error)
		go func() {
			errCh <- run(ctx, opts)
		}()

		// サーバーの起動を待つ
		time.Sleep(100 * time.Millisecond)

		client1 := NewTestClient(t, "localhost:"+opts.MQTTPort, "test-client-1")
		client2 := NewTestClient(t, "localhost:"+opts.MQTTPort, "test-client-2")

		// client1がプレイヤーの位置を更新
		playerState := &shared.PlayerState{
			PlayerId: "test-client-1",
			Position: &shared.Position{
				X: 10,
				Y: 20,
			},
			Direction: shared.Direction_RIGHT,
		}
		payload, err := proto.Marshal(playerState)
		require.NoError(t, err)

		token := client1.client.Publish("player_state", 0, false, payload)
		token.Wait()
		require.NoError(t, token.Error())

		// client2が更新を受信するまで待つ
		time.Sleep(100 * time.Millisecond)

		// client2が受信したメッセージを確認
		var receivedState shared.PlayerState
		found := false
		for _, msg := range client2.Messages() {
			if msg.Topic() == "player_state" {
				err := proto.Unmarshal(msg.Payload(), &receivedState)
				require.NoError(t, err)
				if receivedState.PlayerId == "test-client-1" {
					found = true
				}
			}
		}

		require.True(t, found, "client2がclient1の位置更新を受信している")
		require.Equal(t, int32(10), receivedState.Position.X)
		require.Equal(t, int32(20), receivedState.Position.Y)
		require.Equal(t, shared.Direction_RIGHT, receivedState.Direction)

		// サーバーを正常に終了
		cancel()
		err = <-errCh
		require.NoError(t, err)
	})
}
