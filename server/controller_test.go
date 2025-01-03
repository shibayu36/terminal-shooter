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
	err := controller.OnConnected(cl1, nil)
	require.NoError(t, err)
	assert.Equal(t, &PlayerState{Position: &Position{X: 0, Y: 0}}, state.GetPlayers()[PlayerID("id1")], "cl1が追加された")
	assert.Equal(t, broker.clients[cl1.id], cl1, "cl1がbrokerに追加された")

	cl2 := &mockClient{id: "id2"}
	err = controller.OnConnected(cl2, nil)
	require.NoError(t, err)
	assert.Equal(t, &PlayerState{Position: &Position{X: 0, Y: 0}}, state.GetPlayers()[PlayerID("id2")], "cl2が追加された")
	assert.Equal(t, broker.clients[cl2.id], cl2, "cl2がbrokerに追加された")
}

func TestController_OnSubscribed(t *testing.T) {
	// 自分以外の既存プレイヤー全員に自分の位置を送信する

	broker := NewBroker()
	state := NewGameState()
	controller := NewController(broker, state)

	cl1 := &mockClient{id: "id1"}
	err := controller.OnConnected(cl1, nil)
	require.NoError(t, err)
	state.UpdatePlayerPosition(PlayerID("id1"), &Position{X: 5, Y: 10})

	cl2 := &mockClient{id: "id2"}
	err = controller.OnConnected(cl2, nil)
	require.NoError(t, err)
	state.UpdatePlayerPosition(PlayerID("id2"), &Position{X: 10, Y: 20})

	cl3 := &mockClient{id: "id3"}
	err = controller.OnConnected(cl3, nil)
	require.NoError(t, err)

	err = controller.OnSubscribed(cl3, nil)
	require.NoError(t, err)

	require.Len(t, cl3.published, 2)

	// Topic名はplayer_state
	assert.Equal(t, "player_state", cl3.published[0].TopicName)
	assert.Equal(t, "player_state", cl3.published[1].TopicName)

	idToState := map[string]*shared.PlayerState{}
	for _, published := range cl3.published {
		publishedState := &shared.PlayerState{}
		err := proto.Unmarshal(published.Payload, publishedState)
		require.NoError(t, err)
		idToState[publishedState.GetPlayerId()] = publishedState
	}

	// id1の位置が送信されている
	assert.EqualValues(t, 5, idToState["id1"].GetPosition().GetX())
	assert.EqualValues(t, 10, idToState["id1"].GetPosition().GetY())
	assert.Equal(t, shared.Status_ALIVE, idToState["id1"].GetStatus())

	// id2の位置が送信されている
	assert.EqualValues(t, 10, idToState["id2"].GetPosition().GetX())
	assert.EqualValues(t, 20, idToState["id2"].GetPosition().GetY())
	assert.Equal(t, shared.Status_ALIVE, idToState["id2"].GetStatus())
}

func TestController_OnPublished_PlayerState(t *testing.T) {
	// player_stateパケットを受信したら、そのプレイヤーの位置を更新し、全員にそのプレイヤーの位置を送信する

	broker := NewBroker()
	state := NewGameState()
	controller := NewController(broker, state)

	cl1 := &mockClient{id: "id1"}
	err := controller.OnConnected(cl1, nil)
	require.NoError(t, err)

	cl2 := &mockClient{id: "id2"}
	err = controller.OnConnected(cl2, nil)
	require.NoError(t, err)

	cl3 := &mockClient{id: "id3"}
	err = controller.OnConnected(cl3, nil)
	require.NoError(t, err)

	// cl3からのplayer_stateを受信する
	{
		payload, err := proto.Marshal(&shared.PlayerState{
			PlayerId: "id3",
			Position: &shared.Position{X: 15, Y: 25},
		})
		require.NoError(t, err)

		packet := &packets.PublishPacket{
			TopicName: "player_state",
			Payload:   payload,
		}

		err = controller.OnPublished(cl3, packet)
		require.NoError(t, err)
	}

	// cl3の位置が更新されている
	assert.EqualValues(t, 15, state.GetPlayers()[PlayerID("id3")].Position.X)
	assert.EqualValues(t, 25, state.GetPlayers()[PlayerID("id3")].Position.Y)

	// cl1, cl2, cl3にそれぞれ位置が送信されている
	for _, cl := range []*mockClient{cl1, cl2, cl3} {
		require.Len(t, cl.published, 1)
		assert.Equal(t, "player_state", cl.published[0].TopicName)
		publishedState := &shared.PlayerState{}
		err := proto.Unmarshal(cl.published[0].Payload, publishedState)
		require.NoError(t, err)
		assert.EqualValues(t, 15, publishedState.GetPosition().GetX())
		assert.EqualValues(t, 25, publishedState.GetPosition().GetY())
		assert.Equal(t, shared.Status_ALIVE, publishedState.GetStatus())
	}
}

func TestController_OnDisconnected(t *testing.T) {
	// 切断したら、そのプレイヤーを削除し、そのプレイヤーが切断したことを全員に送信する

	broker := NewBroker()
	state := NewGameState()
	controller := NewController(broker, state)

	cl1 := &mockClient{id: "id1"}
	err := controller.OnConnected(cl1, nil)
	require.NoError(t, err)

	cl2 := &mockClient{id: "id2"}
	err = controller.OnConnected(cl2, nil)
	require.NoError(t, err)

	cl3 := &mockClient{id: "id3"}
	broker.AddClient(cl3)
	err = controller.OnConnected(cl3, nil)
	require.NoError(t, err)

	err = controller.OnDisconnected(cl1)
	require.NoError(t, err)

	assert.NotContains(t, state.GetPlayers(), PlayerID("id1"))
	assert.Contains(t, state.GetPlayers(), PlayerID("id2"))
	assert.Contains(t, state.GetPlayers(), PlayerID("id3"))

	// cl1の切断がcl2, cl3に送信されている
	for _, cl := range []*mockClient{cl2, cl3} {
		require.Len(t, cl.published, 1)
		assert.Equal(t, "player_state", cl.published[0].TopicName)
		publishedState := &shared.PlayerState{}
		err := proto.Unmarshal(cl.published[0].Payload, publishedState)
		require.NoError(t, err)
		assert.Equal(t, shared.Status_DISCONNECTED, publishedState.GetStatus())
	}

	// cl1がbrokerから削除されている
	assert.NotContains(t, broker.clients, cl1.id)
}
