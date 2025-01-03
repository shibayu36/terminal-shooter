package main

import (
	"fmt"
	"log/slog"

	"github.com/cockroachdb/errors"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/shibayu36/terminal-shooter/shared"
	"google.golang.org/protobuf/proto"
)

// Controller クライアントからのパケットをゲームの状態に反映し、さらに他のクライアントに状態同期をする役割を持つ
type Controller struct {
	broker *Broker
	game   *GameState
}

var _ Hooker = (*Controller)(nil)

func NewController(broker *Broker, game *GameState) *Controller {
	return &Controller{broker: broker, game: game}
}

func (c *Controller) OnConnected(client Client, _ *packets.ConnectPacket) error {
	c.broker.AddClient(client)
	c.game.AddPlayer(PlayerID(client.ID()))

	// Player状態を出力
	slog.Info("all players", "players", c.game.String())

	return nil
}

func (c *Controller) OnSubscribed(client Client, _ *packets.SubscribePacket) error {
	// Subscribeが来たら、現在の他プレイヤーの位置をそのクライアントに送信する
	for playerID, player := range c.game.GetPlayers() {
		if playerID == PlayerID(client.ID()) {
			continue
		}

		playerState := &shared.PlayerState{
			PlayerId: string(playerID),
			Position: &shared.Position{
				X: int32(player.Position.X),
				Y: int32(player.Position.Y),
			},
			Status: shared.Status_ALIVE,
		}
		payload, err := proto.Marshal(playerState)
		if err != nil {
			return errors.Wrap(err, "failed to marshal player state")
		}

		slog.Info("send player state on subscribe", "player", playerState.String())

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
	default:
		return errors.New(fmt.Sprintf("invalid topic name: %s", publishPacket.TopicName))
	}
}

func (c *Controller) OnDisconnected(client Client) error {
	slog.Info("client disconnected", "client_id", client.ID())
	c.broker.RemoveClient(client)
	c.game.RemovePlayer(PlayerID(client.ID()))

	playerState := &shared.PlayerState{
		PlayerId: client.ID(),
		Status:   shared.Status_DISCONNECTED,
	}
	payload, err := proto.Marshal(playerState)
	if err != nil {
		return errors.Wrap(err, "failed to marshal player state")
	}
	c.broker.Broadcast("player_state", payload)

	return nil
}

// player_stateパケットを受信した時の処理
func (c *Controller) onReceivePlayerState(client Client, publishPacket *packets.PublishPacket) error {
	playerID := PlayerID(client.ID())
	playerState := &shared.PlayerState{}
	err := proto.Unmarshal(publishPacket.Payload, playerState)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal player state")
	}
	position := &Position{
		X: int(playerState.GetPosition().GetX()),
		Y: int(playerState.GetPosition().GetY()),
	}
	c.game.MovePlayer(playerID, position, Direction(playerState.GetDirection()))

	playerState.Status = shared.Status_ALIVE
	payload, err := proto.Marshal(playerState)
	if err != nil {
		return errors.Wrap(err, "failed to marshal player state")
	}
	c.broker.Broadcast("player_state", payload)

	slog.Info("all players", "players", c.game.String())

	return nil
}
