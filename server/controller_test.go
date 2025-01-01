package main

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/shibayu36/terminal-shooter/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockClient struct {
	id        string
	published []*packets.PublishPacket
}

func (c *mockClient) ID() string {
	return c.id
}

func (c *mockClient) Publish(publishPacket *packets.PublishPacket) error {
	c.published = append(c.published, publishPacket)
	return nil
}

func TestController_OnConnected(t *testing.T) {
	broker := NewBroker()
	state := NewGameState()
	controller := NewController(broker, state)

	cl1 := &mockClient{id: "id1"}
	controller.OnConnected(cl1, nil)
	assert.Equal(t, state.GetPlayers()[PlayerID("id1")], &PlayerState{Position: &Position{X: 0, Y: 0}}, "cl1が追加された")
	assert.Len(t, cl1.published, 0, "cl1にはメッセージが送信されていない")

	cl2 := &mockClient{id: "id2"}
	controller.OnConnected(cl2, nil)
	assert.Equal(t, state.GetPlayers()[PlayerID("id2")], &PlayerState{Position: &Position{X: 0, Y: 0}}, "cl2が追加された")
	assert.Len(t, cl2.published, 0, "cl2にはメッセージが送信されていない")
}

func TestController_OnSubscribed(t *testing.T) {
	// 自分以外の既存プレイヤー全員に自分の位置を送信する

	broker := NewBroker()
	state := NewGameState()
	controller := NewController(broker, state)

	cl1 := &mockClient{id: "id1"}
	broker.AddClient(cl1)
	controller.OnConnected(cl1, nil)
	state.UpdatePlayerPosition(PlayerID("id1"), &Position{X: 5, Y: 10})

	cl2 := &mockClient{id: "id2"}
	broker.AddClient(cl2)
	controller.OnConnected(cl2, nil)
	state.UpdatePlayerPosition(PlayerID("id2"), &Position{X: 10, Y: 20})

	cl3 := &mockClient{id: "id3"}
	broker.AddClient(cl3)
	controller.OnConnected(cl3, nil)

	controller.OnSubscribed(cl3, nil)

	require.Len(t, cl3.published, 2)

	// Topic名はplayer_state
	assert.Equal(t, cl3.published[0].TopicName, "player_state")
	assert.Equal(t, cl3.published[1].TopicName, "player_state")

	idToState := map[string]*shared.PlayerState{}
	for _, published := range cl3.published {
		publishedState := &shared.PlayerState{}
		err := proto.Unmarshal(published.Payload, publishedState)
		require.NoError(t, err)
		idToState[publishedState.PlayerId] = publishedState
	}

	// id1の位置が送信されている
	assert.EqualValues(t, 5, idToState["id1"].Position.X)
	assert.EqualValues(t, 10, idToState["id1"].Position.Y)

	// id2の位置が送信されている
	assert.EqualValues(t, 10, idToState["id2"].Position.X)
	assert.EqualValues(t, 20, idToState["id2"].Position.Y)
}
