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
	clientID string
	messages []mqtt.Message
	mu       sync.Mutex `exhaustruct:"optional"`
}

func NewTestClient(t *testing.T, address string, clientID string) *TestClient {
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://" + address).
		SetClientID(clientID)

	c := &TestClient{
		client:   mqtt.NewClient(opts),
		clientID: clientID,
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

func (c *TestClient) GetMessages(topic string) []mqtt.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	var messages []mqtt.Message
	for _, msg := range c.messages {
		if msg.Topic() == topic {
			messages = append(messages, msg)
		}
	}
	return messages
}

func (c *TestClient) MustFindLastPlayerStateMessage(t *testing.T, playerId string) *shared.PlayerState {
	messages := c.GetMessages("player_state")
	for i := len(messages) - 1; i >= 0; i-- {
		var state shared.PlayerState
		err := proto.Unmarshal(messages[i].Payload(), &state)
		require.NoError(t, err)
		if state.PlayerId == playerId {
			return &state
		}
	}
	t.Fatalf("player state not found for player %s", playerId)
	return nil
}

func (c *TestClient) PublishPlayerState(position *shared.Position, direction shared.Direction) error {
	playerState := &shared.PlayerState{
		PlayerId:  c.clientID,
		Position:  position,
		Direction: direction,
	}
	payload, err := proto.Marshal(playerState)
	if err != nil {
		return err
	}

	token := c.client.Publish("player_state", 0, false, payload)
	token.Wait()
	return token.Error()
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

		// client1がプレイヤーの位置を更新すると、client2が受信できる
		{
			err := client1.PublishPlayerState(
				&shared.Position{X: 10, Y: 20},
				shared.Direction_RIGHT,
			)
			require.NoError(t, err)

			// client2が受信したメッセージを確認
			time.Sleep(100 * time.Millisecond)
			receivedState := client2.MustFindLastPlayerStateMessage(t, "test-client-1")
			require.Equal(t, int32(10), receivedState.Position.X)
			require.Equal(t, int32(20), receivedState.Position.Y)
			require.Equal(t, shared.Direction_RIGHT, receivedState.Direction)
		}

		// client2がプレイヤーの位置を更新すると、client1が受信できる
		{
			err := client2.PublishPlayerState(
				&shared.Position{X: 2, Y: 1},
				shared.Direction_UP,
			)
			require.NoError(t, err)

			// client1が受信したメッセージを確認
			time.Sleep(100 * time.Millisecond)
			receivedState := client1.MustFindLastPlayerStateMessage(t, "test-client-2")
			require.Equal(t, int32(2), receivedState.Position.X)
			require.Equal(t, int32(1), receivedState.Position.Y)
			require.Equal(t, shared.Direction_UP, receivedState.Direction)
		}

		// サーバーを正常に終了
		cancel()
		err := <-errCh
		require.NoError(t, err)
	})
}
