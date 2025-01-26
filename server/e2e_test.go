package main

import (
	"context"
	"sync"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/shibayu36/terminal-shooter/shared"
	"github.com/stretchr/testify/assert"
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

	if token := c.client.Subscribe("#", 0, c.OnPublished); token.Wait() && token.Error() != nil {
		t.Fatalf("failed to subscribe to server: %v", token.Error())
	}

	t.Cleanup(func() {
		c.Close()
	})

	return c
}

func (c *TestClient) OnPublished(client mqtt.Client, message mqtt.Message) {
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

func (c *TestClient) MustFindLastPlayerStateMessage(t *testing.T, playerID string) *shared.PlayerState {
	messages := c.GetMessages("player_state")
	for i := len(messages) - 1; i >= 0; i-- {
		var state shared.PlayerState
		err := proto.Unmarshal(messages[i].Payload(), &state)
		require.NoError(t, err)
		if state.PlayerId == playerID {
			return &state
		}
	}
	t.Fatalf("player state not found for player %s", playerID)
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

func (c *TestClient) PublishPlayerAction(actionType shared.ActionType) error {
	req := &shared.PlayerActionRequest{
		Type: actionType,
	}
	payload, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	token := c.client.Publish("player_action", 0, false, payload)
	token.Wait()
	return token.Error()
}

func (c *TestClient) MustFindItemStateMessages(t *testing.T) []*shared.ItemState {
	messages := c.GetMessages("item_state")
	states := make([]*shared.ItemState, 0, len(messages))
	for _, msg := range messages {
		var state shared.ItemState
		err := proto.Unmarshal(msg.Payload(), &state)
		require.NoError(t, err)
		states = append(states, &state)
	}
	return states
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
		NewTestClient(t, "localhost:"+opts.MQTTPort, "connect1")

		// クライアント2が接続できる
		NewTestClient(t, "localhost:"+opts.MQTTPort, "connect2")

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
		t.Cleanup(func() {
			cancel()
			err := <-errCh
			require.NoError(t, err)
		})

		// サーバーの起動を待つ
		time.Sleep(100 * time.Millisecond)

		client1 := NewTestClient(t, "localhost:"+opts.MQTTPort, "player1")
		client2 := NewTestClient(t, "localhost:"+opts.MQTTPort, "player2")

		// client1がプレイヤーの位置を更新すると、client2が受信できる
		{
			err := client1.PublishPlayerState(
				&shared.Position{X: 10, Y: 20},
				shared.Direction_RIGHT,
			)
			require.NoError(t, err)

			// client2が受信したメッセージを確認
			time.Sleep(100 * time.Millisecond)
			receivedState := client2.MustFindLastPlayerStateMessage(t, "player1")
			assert.Equal(t, int32(10), receivedState.Position.X)
			assert.Equal(t, int32(20), receivedState.Position.Y)
			assert.Equal(t, shared.Direction_RIGHT, receivedState.Direction)
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
			receivedState := client1.MustFindLastPlayerStateMessage(t, "player2")
			assert.Equal(t, int32(2), receivedState.Position.X)
			assert.Equal(t, int32(1), receivedState.Position.Y)
			assert.Equal(t, shared.Direction_UP, receivedState.Direction)
		}
	})

	t.Run("アイテムの状態がプレイヤーに配信され続ける", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// サーバー起動
		errCh := make(chan error)
		go func() {
			errCh <- run(ctx, opts)
		}()
		t.Cleanup(func() {
			cancel()
			err := <-errCh
			require.NoError(t, err)
		})

		// サーバーの起動を待つ
		time.Sleep(100 * time.Millisecond)

		client1 := NewTestClient(t, "localhost:"+opts.MQTTPort, "shoot-player1")
		client2 := NewTestClient(t, "localhost:"+opts.MQTTPort, "shoot-player2")

		// client1が右向きで位置を設定
		err := client1.PublishPlayerState(
			&shared.Position{X: 10, Y: 20},
			shared.Direction_RIGHT,
		)
		require.NoError(t, err)

		// client1が銃を発射
		err = client1.PublishPlayerAction(shared.ActionType_SHOOT_BULLET)
		require.NoError(t, err)

		// 弾丸が生成され、client1, client2が受信できることを確認
		// 60FPSで30tickするので、最低でも500ms待つ
		time.Sleep(550 * time.Millisecond)
		for _, client := range []*TestClient{client1, client2} {
			// client1
			itemMessages := client.MustFindItemStateMessages(t)
			require.Len(t, itemMessages, 1)
			assert.Equal(t, shared.ItemType_BULLET, itemMessages[0].Type)
			assert.Equal(t, shared.ItemStatus_ACTIVE, itemMessages[0].Status)
			assert.Equal(t, int32(12), itemMessages[0].Position.X, "右向きに発射され、さらに1進んだ値")
			assert.Equal(t, int32(20), itemMessages[0].Position.Y)
		}

		// さらに1マス進むのを待ち、受け取れることを確認
		time.Sleep(550 * time.Millisecond)
		for _, client := range []*TestClient{client1, client2} {
			itemMessages := client.MustFindItemStateMessages(t)
			require.Len(t, itemMessages, 2)
			assert.Equal(t, int32(13), itemMessages[1].Position.X, "さらに1マス進んた値")
			assert.Equal(t, int32(20), itemMessages[1].Position.Y)
		}
	})
}
