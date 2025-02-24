package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/shibayu36/terminal-shooter/server/game"
	"github.com/shibayu36/terminal-shooter/server/stats"
	"github.com/shibayu36/terminal-shooter/shared"
	"google.golang.org/protobuf/proto"
)

// Controller クライアントからのパケットをゲームの状態に反映し、さらに他のクライアントに状態同期をする役割を持つ
type Controller struct {
	broker *Broker
	game   *game.Game
}

var _ Hooker = (*Controller)(nil)

func NewController(broker *Broker, game *game.Game) *Controller {
	return &Controller{broker: broker, game: game}
}

func (c *Controller) OnConnected(client Client, _ *packets.ConnectPacket) error {
	c.broker.AddClient(client)
	c.game.AddPlayer(game.PlayerID(client.ID()))

	stats.ActiveClients.Inc()

	// Player状態を出力
	slog.Info("all players", "players", c.game.String())

	return nil
}

func (c *Controller) OnSubscribed(client Client, _ *packets.SubscribePacket) error {
	// Subscribeが来たら、現在の他プレイヤーの位置をそのクライアントに送信する
	for playerID, player := range c.game.GetPlayers() {
		if playerID == game.PlayerID(client.ID()) {
			continue
		}

		sharedPlayerState := player.ToSharedPlayerState()

		payload, err := proto.Marshal(sharedPlayerState)
		if err != nil {
			return errors.Wrap(err, "failed to marshal player state")
		}

		slog.Info("send player state on subscribe", "player", sharedPlayerState.String())

		err = c.broker.Send(client.ID(), "player_state", payload)
		if err != nil {
			return errors.Wrap(err, "failed to send player state")
		}
	}

	return nil
}

func (c *Controller) OnPublished(client Client, publishPacket *packets.PublishPacket) error {
	switch publishPacket.TopicName {
	case "player_state":
		return c.onReceivePlayerState(client, publishPacket)
	case "player_action":
		return c.onReceivePlayerAction(client, publishPacket)
	default:
		return errors.New(fmt.Sprintf("invalid topic name: %s", publishPacket.TopicName))
	}
}

func (c *Controller) OnDisconnected(client Client) error {
	slog.Info("client disconnected", "client_id", client.ID())
	c.broker.RemoveClient(client)

	stats.ActiveClients.Dec()

	c.game.RemovePlayer(game.PlayerID(client.ID()))

	playerState := &shared.PlayerState{
		PlayerId: client.ID(),
		Status:   shared.Status_DISCONNECTED,
	}
	payload, err := proto.Marshal(playerState)
	if err != nil {
		return errors.Wrap(err, "failed to marshal player state")
	}
	err = c.broker.Broadcast("player_state", payload)
	if err != nil {
		return errors.Wrap(err, "failed to broadcast player state")
	}

	return nil
}

// player_stateパケットを受信した時の処理
func (c *Controller) onReceivePlayerState(client Client, publishPacket *packets.PublishPacket) error {
	playerID := game.PlayerID(client.ID())
	playerState := &shared.PlayerState{}
	err := proto.Unmarshal(publishPacket.Payload, playerState)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal player state")
	}

	direction, err := game.FromSharedDirection(playerState.GetDirection())
	if err != nil {
		// 方向が不正な場合は無視する
		//nolint:nilerr
		return nil
	}

	updatedPlayer := c.game.MovePlayer(
		playerID,
		game.Position{
			X: int(playerState.GetPosition().GetX()),
			Y: int(playerState.GetPosition().GetY()),
		},
		direction,
	)

	payload, err := proto.Marshal(updatedPlayer.ToSharedPlayerState())
	if err != nil {
		return errors.Wrap(err, "failed to marshal player state")
	}
	err = c.broker.Broadcast("player_state", payload)
	if err != nil {
		return errors.Wrap(err, "failed to broadcast player state")
	}

	slog.Info("all players", "players", c.game.String())

	return nil
}

func (c *Controller) onReceivePlayerAction(client Client, publishPacket *packets.PublishPacket) error {
	playerID := game.PlayerID(client.ID())

	playerActionRequest := &shared.PlayerActionRequest{}
	err := proto.Unmarshal(publishPacket.Payload, playerActionRequest)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal player action request")
	}

	//nolint:gocritic // GetType()が増えることを見越してsingleCaseSwitchをignore
	switch playerActionRequest.GetType() {
	case shared.ActionType_SHOOT_BULLET:
		c.game.ShootBullet(playerID)
	case shared.ActionType_PLACE_BOMB:
		c.game.PlaceBomb(playerID)
	}

	return nil
}

// StartPublishLoop ゲームの状態を定期的にpublishするループを開始する
func (c *Controller) StartPublishLoop(ctx context.Context, updatedCh <-chan game.UpdatedResult) {
	go func() {
		for {
			select {
			case updatedResult, ok := <-updatedCh:
				if !ok {
					return
				}
				c.publishStates(updatedResult)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *Controller) publishStates(updatedResult game.UpdatedResult) {
	start := time.Now()
	defer func() {
		stats.PublishStatesDuration.Observe(time.Since(start).Seconds())
	}()

	switch updatedResult.Type {
	case game.UpdatedResultTypeItemsUpdated:
		c.publishItemStates()
	case game.UpdatedResultTypePlayersUpdated:
		c.publishPlayerStates()
	}
}

func (c *Controller) publishItemStates() {
	// Activeなアイテムを送信する
	for _, item := range c.game.GetItems() {
		var itemType shared.ItemType
		switch item.Type() {
		case game.ItemTypeBullet:
			itemType = shared.ItemType_BULLET
		case game.ItemTypeBomb:
			itemType = shared.ItemType_BOMB
		}

		itemState := &shared.ItemState{
			ItemId: string(item.ID()),
			Type:   itemType,
			Position: &shared.Position{
				X: int32(item.Position().X),
				Y: int32(item.Position().Y),
			},
			Status: shared.ItemStatus_ACTIVE,
		}

		payload, err := proto.Marshal(itemState)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to marshal item state\n%+v", err))
			continue
		}
		err = c.broker.Broadcast("item_state", payload)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to broadcast item state\n%+v", err))
		}
	}

	// 削除されたアイテムを送信する
	for _, removedItem := range c.game.GetRemovedItems() {
		itemState := &shared.ItemState{
			ItemId: string(removedItem.ID()),
			Status: shared.ItemStatus_REMOVED,
		}

		payload, err := proto.Marshal(itemState)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to marshal item state\n%+v", err))
			continue
		}
		err = c.broker.Broadcast("item_state", payload)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to broadcast item state\n%+v", err))
			continue
		}

		// Broadcastが成功したら削除アイテムは不要になる
		c.game.ClearRemovedItem(removedItem.ID())
	}
}

func (c *Controller) publishPlayerStates() {
	for _, player := range c.game.GetPlayers() {
		payload, err := proto.Marshal(player.ToSharedPlayerState())
		if err != nil {
			slog.Error(fmt.Sprintf("failed to marshal player state\n%+v", err))
			continue
		}
		err = c.broker.Broadcast("player_state", payload)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to broadcast player state\n%+v", err))
		}
	}
}
