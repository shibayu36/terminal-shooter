package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/stretchr/testify/require"
)

// テストクライアントの実装
type TestClient struct {
	conn     net.Conn
	clientID string
}

func NewTestClient(address string, clientID string) (*TestClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	client := &TestClient{
		conn:     conn,
		clientID: clientID,
	}

	// CONNECT パケットを送信
	//nolint:forcetypeassert
	connectPacket := packets.NewControlPacket(packets.Connect).(*packets.ConnectPacket)
	connectPacket.ClientIdentifier = clientID
	if err := connectPacket.Write(client.conn); err != nil {
		return nil, err
	}

	// CONNACK パケットを受信
	packet, err := packets.ReadPacket(client.conn)
	if err != nil {
		return nil, err
	}
	connack, ok := packet.(*packets.ConnackPacket)
	if !ok || connack.ReturnCode != packets.Accepted {
		return nil, err
	}

	return client, nil
}

func (c *TestClient) Close() error {
	return c.conn.Close()
}

func TestE2E(t *testing.T) {
	t.Run("クライアントが接続できる", func(t *testing.T) {
		opts := &runOptions{
			MQTTPort:    "11883",
			MetricsPort: "12113",
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// サーバー起動
		errCh := make(chan error)
		go func() {
			errCh <- run(ctx, opts)
		}()

		// サーバーの起動を待つ
		time.Sleep(100 * time.Millisecond)

		// クライアントが接続できることを確認
		client, err := NewTestClient("localhost:"+opts.MQTTPort, "test-client")
		require.NoError(t, err)
		defer client.Close()

		// テスト終了時にサーバーを終了
		cancel()
		err = <-errCh
		require.NoError(t, err)
	})
}
