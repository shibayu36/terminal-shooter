package main

import (
	"context"
	"maps"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/shibayu36/terminal-shooter/server/game"
	"github.com/shibayu36/terminal-shooter/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockClient struct {
	id        string
	published []*packets.PublishPacket
	mu        sync.Mutex
}

func (c *mockClient) ID() string {
	return c.id
}

func (c *mockClient) Publish(publishPacket *packets.PublishPacket) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.published = append(c.published, publishPacket)
	return nil
}

func (c *mockClient) Published() []*packets.PublishPacket {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.published
}

func TestController_OnConnected(t *testing.T) {
	broker := NewBroker()
	state := game.NewGame(30, 30)
	controller := NewController(broker, state)

	cl1 := &mockClient{id: "id1"}
	err := controller.OnConnected(cl1, nil)
	require.NoError(t, err)
	assert.Equal(t, &game.Player{
		PlayerID:  game.PlayerID("id1"),
		Position:  game.Position{X: 0, Y: 0},
		Direction: game.DirectionUp,
		Status:    game.PlayerStatusAlive,
	}, state.GetPlayers()[game.PlayerID("id1")], "cl1が追加された")
	assert.Equal(t, broker.clients[cl1.id], cl1, "cl1がbrokerに追加された")

	cl2 := &mockClient{id: "id2"}
	err = controller.OnConnected(cl2, nil)
	require.NoError(t, err)
	assert.Equal(t, &game.Player{
		PlayerID:  game.PlayerID("id2"),
		Position:  game.Position{X: 0, Y: 0},
		Direction: game.DirectionUp,
		Status:    game.PlayerStatusAlive,
	}, state.GetPlayers()[game.PlayerID("id2")], "cl2が追加された")
	assert.Equal(t, broker.clients[cl2.id], cl2, "cl2がbrokerに追加された")
}

func TestController_OnSubscribed(t *testing.T) {
	// 自分以外の既存プレイヤー全員に自分の位置を送信する

	broker := NewBroker()
	state := game.NewGame(30, 30)
	controller := NewController(broker, state)

	cl1 := &mockClient{id: "id1"}
	err := controller.OnConnected(cl1, nil)
	require.NoError(t, err)
	state.MovePlayer(game.PlayerID("id1"), game.Position{X: 5, Y: 10}, game.DirectionRight)

	cl2 := &mockClient{id: "id2"}
	err = controller.OnConnected(cl2, nil)
	require.NoError(t, err)
	state.MovePlayer(game.PlayerID("id2"), game.Position{X: 10, Y: 20}, game.DirectionLeft)

	cl3 := &mockClient{id: "id3"}
	err = controller.OnConnected(cl3, nil)
	require.NoError(t, err)

	err = controller.OnSubscribed(cl3, nil)
	require.NoError(t, err)

	require.Len(t, cl3.Published(), 2)

	// Topic名はplayer_state
	assert.Equal(t, "player_state", cl3.Published()[0].TopicName)
	assert.Equal(t, "player_state", cl3.Published()[1].TopicName)

	idToState := map[string]*shared.PlayerState{}
	for _, published := range cl3.Published() {
		publishedState := &shared.PlayerState{}
		err := proto.Unmarshal(published.Payload, publishedState)
		require.NoError(t, err)
		idToState[publishedState.GetPlayerId()] = publishedState
	}

	// id1の位置と向きが送信されている
	assert.EqualValues(t, 5, idToState["id1"].GetPosition().GetX())
	assert.EqualValues(t, 10, idToState["id1"].GetPosition().GetY())
	assert.Equal(t, shared.Direction_RIGHT, idToState["id1"].GetDirection())
	assert.Equal(t, shared.Status_ALIVE, idToState["id1"].GetStatus())

	// id2の位置と向きが送信されている
	assert.EqualValues(t, 10, idToState["id2"].GetPosition().GetX())
	assert.EqualValues(t, 20, idToState["id2"].GetPosition().GetY())
	assert.Equal(t, shared.Direction_LEFT, idToState["id2"].GetDirection())
	assert.Equal(t, shared.Status_ALIVE, idToState["id2"].GetStatus())
}

func TestController_OnPublished_PlayerState(t *testing.T) {
	// player_stateパケットを受信したら、そのプレイヤーの位置を更新し、全員にそのプレイヤーの位置を送信する

	broker := NewBroker()
	state := game.NewGame(30, 30)
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
			PlayerId:  "id3",
			Position:  &shared.Position{X: 15, Y: 25},
			Direction: shared.Direction_RIGHT,
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
	assert.EqualValues(t, 15, state.GetPlayers()[game.PlayerID("id3")].Position.X)
	assert.EqualValues(t, 25, state.GetPlayers()[game.PlayerID("id3")].Position.Y)
	assert.Equal(t, game.DirectionRight, state.GetPlayers()[game.PlayerID("id3")].Direction)

	// cl1, cl2, cl3にそれぞれ位置が送信されている
	for _, cl := range []*mockClient{cl1, cl2, cl3} {
		require.Len(t, cl.Published(), 1)
		assert.Equal(t, "player_state", cl.Published()[0].TopicName)
		publishedState := &shared.PlayerState{}
		err := proto.Unmarshal(cl.Published()[0].Payload, publishedState)
		require.NoError(t, err)
		assert.EqualValues(t, 15, publishedState.GetPosition().GetX())
		assert.EqualValues(t, 25, publishedState.GetPosition().GetY())
		assert.Equal(t, shared.Status_ALIVE, publishedState.GetStatus())
	}
}

func TestController_OnPublished_PlayerAction_ShootBullet(t *testing.T) {
	broker := NewBroker()
	state := game.NewGame(30, 30)
	controller := NewController(broker, state)

	cl1 := &mockClient{id: "id1"}
	err := controller.OnConnected(cl1, nil)
	require.NoError(t, err)

	// cl1の位置を更新する
	state.MovePlayer(game.PlayerID("id1"), game.Position{X: 5, Y: 10}, game.DirectionRight)

	// cl1からのplayer_action ShootBulletを受信する
	{
		payload, err := proto.Marshal(&shared.PlayerActionRequest{
			Type: shared.ActionType_SHOOT_BULLET,
		})
		require.NoError(t, err)

		packet := &packets.PublishPacket{
			TopicName: "player_action",
			Payload:   payload,
		}

		err = controller.OnPublished(cl1, packet)
		require.NoError(t, err)
	}

	// cl1の目の前に弾が追加されている
	items := slices.Collect(maps.Values(state.GetItems()))
	assert.Len(t, items, 1)
	bullet := items[0]
	assert.Equal(t, game.ItemTypeBullet, bullet.Type())
	assert.Equal(t, game.Position{X: 6, Y: 10}, bullet.Position())
}

func TestController_OnDisconnected(t *testing.T) {
	// 切断したら、そのプレイヤーを削除し、そのプレイヤーが切断したことを全員に送信する

	broker := NewBroker()
	state := game.NewGame(30, 30)
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

	assert.NotContains(t, state.GetPlayers(), game.PlayerID("id1"))
	assert.Contains(t, state.GetPlayers(), game.PlayerID("id2"))
	assert.Contains(t, state.GetPlayers(), game.PlayerID("id3"))

	// cl1の切断がcl2, cl3に送信されている
	for _, cl := range []*mockClient{cl2, cl3} {
		require.Len(t, cl.Published(), 1)
		assert.Equal(t, "player_state", cl.Published()[0].TopicName)
		publishedState := &shared.PlayerState{}
		err := proto.Unmarshal(cl.Published()[0].Payload, publishedState)
		require.NoError(t, err)
		assert.Equal(t, shared.Status_DISCONNECTED, publishedState.GetStatus())
	}

	// cl1がbrokerから削除されている
	assert.NotContains(t, broker.clients, cl1.id)
}

func TestController_StartPublishLoop(t *testing.T) {
	t.Run("アクティブなアイテムの情報を送れる", func(t *testing.T) {
		broker := NewBroker()
		state := game.NewGame(30, 30)
		controller := NewController(broker, state)

		cl1 := &mockClient{id: "id1"}
		err := controller.OnConnected(cl1, nil)
		require.NoError(t, err)

		cl2 := &mockClient{id: "id2"}
		err = controller.OnConnected(cl2, nil)
		require.NoError(t, err)

		updatedCh := make(chan game.UpdatedResult)
		controller.StartPublishLoop(context.Background(), updatedCh)

		bulletID1 := state.AddBullet(game.Position{X: 1, Y: 2}, game.DirectionRight)
		bulletID2 := state.AddBullet(game.Position{X: 2, Y: 3}, game.DirectionUp)

		updatedCh <- game.UpdatedResult{Type: game.UpdatedResultTypeItemsUpdated}

		// TODO: 待つための良い手法があれば変更
		time.Sleep(10 * time.Millisecond)

		// アイテムの状態が全てのクライアントに送信されている
		for _, cl := range []*mockClient{cl1, cl2} {
			require.Len(t, cl.Published(), 2)
			assert.Equal(t, "item_state", cl.Published()[0].TopicName)
			assert.Equal(t, "item_state", cl.Published()[1].TopicName)

			idToState := map[game.ItemID]*shared.ItemState{}
			for _, published := range cl.Published() {
				publishedState := &shared.ItemState{}
				err := proto.Unmarshal(published.Payload, publishedState)
				require.NoError(t, err)
				idToState[game.ItemID(publishedState.GetItemId())] = publishedState
			}

			assert.EqualValues(t, 1, idToState[bulletID1].GetPosition().GetX())
			assert.EqualValues(t, 2, idToState[bulletID1].GetPosition().GetY())
			assert.Equal(t, shared.ItemType_BULLET, idToState[bulletID1].GetType())

			assert.EqualValues(t, 2, idToState[bulletID2].GetPosition().GetX())
			assert.EqualValues(t, 3, idToState[bulletID2].GetPosition().GetY())
			assert.Equal(t, shared.ItemType_BULLET, idToState[bulletID2].GetType())
		}
	})

	t.Run("アクティブなアイテムと削除済みアイテムを同時に送れる", func(t *testing.T) {
		broker := NewBroker()
		state := game.NewGame(30, 30)
		controller := NewController(broker, state)

		client := &mockClient{id: "id1"}
		err := controller.OnConnected(client, nil)
		require.NoError(t, err)

		bulletID1 := state.AddBullet(game.Position{X: 1, Y: 2}, game.DirectionRight)
		bulletID2 := state.AddBullet(game.Position{X: 2, Y: 3}, game.DirectionUp)
		state.RemoveItem(bulletID1)

		updatedCh := make(chan game.UpdatedResult)
		controller.StartPublishLoop(context.Background(), updatedCh)

		updatedCh <- game.UpdatedResult{Type: game.UpdatedResultTypeItemsUpdated}

		// TODO: 待つための良い手法があれば変更
		time.Sleep(10 * time.Millisecond)

		// アクティブなアイテムと削除済みアイテムが同時に送信されている
		idToState := map[game.ItemID]*shared.ItemState{}
		for _, published := range client.Published() {
			publishedState := &shared.ItemState{}
			err := proto.Unmarshal(published.Payload, publishedState)
			require.NoError(t, err)
			idToState[game.ItemID(publishedState.GetItemId())] = publishedState
		}

		assert.Equal(t, shared.ItemStatus_REMOVED, idToState[bulletID1].GetStatus())

		assert.Equal(t, shared.ItemStatus_ACTIVE, idToState[bulletID2].GetStatus())
		assert.EqualValues(t, 2, idToState[bulletID2].GetPosition().GetX())
		assert.EqualValues(t, 3, idToState[bulletID2].GetPosition().GetY())
		assert.Equal(t, shared.ItemType_BULLET, idToState[bulletID2].GetType())

		// 削除送信が成功したので、stateから削除済みアイテムがクリアされている
		assert.NotContains(t, state.GetRemovedItems(), bulletID1)
	})

	t.Run("プレイヤーの更新を送信できる", func(t *testing.T) {
		broker := NewBroker()
		state := game.NewGame(30, 30)
		controller := NewController(broker, state)

		cl1 := &mockClient{id: "id1"}
		err := controller.OnConnected(cl1, nil)
		require.NoError(t, err)
		state.MovePlayer(game.PlayerID("id1"), game.Position{X: 5, Y: 10}, game.DirectionRight)

		cl2 := &mockClient{id: "id2"}
		err = controller.OnConnected(cl2, nil)
		require.NoError(t, err)
		state.MovePlayer(game.PlayerID("id2"), game.Position{X: 10, Y: 20}, game.DirectionLeft)

		updatedCh := make(chan game.UpdatedResult)
		controller.StartPublishLoop(context.Background(), updatedCh)

		updatedCh <- game.UpdatedResult{Type: game.UpdatedResultTypePlayersUpdated}

		// TODO: 待つための良い手法があれば変更
		time.Sleep(10 * time.Millisecond)

		// cl1, cl2にそれぞれプレイヤーの更新が送信されている
		for _, cl := range []*mockClient{cl1, cl2} {
			require.Len(t, cl.Published(), 2)
			assert.Equal(t, "player_state", cl.Published()[0].TopicName)
			assert.Equal(t, "player_state", cl.Published()[1].TopicName)

			idToState := map[game.PlayerID]*shared.PlayerState{}
			for _, published := range cl.Published() {
				publishedState := &shared.PlayerState{}
				err := proto.Unmarshal(published.Payload, publishedState)
				require.NoError(t, err)
				idToState[game.PlayerID(publishedState.GetPlayerId())] = publishedState
			}

			assert.EqualValues(t, 5, idToState["id1"].GetPosition().GetX())
			assert.EqualValues(t, 10, idToState["id1"].GetPosition().GetY())

			assert.EqualValues(t, 10, idToState["id2"].GetPosition().GetX())
			assert.EqualValues(t, 20, idToState["id2"].GetPosition().GetY())
		}
	})
}
