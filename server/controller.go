package main

import (
	"fmt"
	"log/slog"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/shibayu36/terminal-shooter/shared"

	"github.com/cockroachdb/errors"
	"google.golang.org/protobuf/proto"
)

// Controller クライアントからのパケットをゲームの状態に反映し、さらに他のクライアントに状態同期をする役割を持つ
type Controller struct {
	broker *Broker
	game   *GameState
}

var _ Hooker = &Controller{}

func NewController(broker *Broker, game *GameState) *Controller {
	return &Controller{broker: broker, game: game}
}

func (c *Controller) OnConnected(cl Client, pk *packets.ConnectPacket) error {
	c.broker.AddClient(cl)
	c.game.AddPlayer(PlayerID(cl.ID()), &PlayerState{Position: &Position{X: 0, Y: 0}})

	// Player状態を出力
	slog.Info("all players", "players", c.game.String())

	return nil
}

func (c *Controller) OnSubscribed(cl Client, pk *packets.SubscribePacket) error {
	// Subscribeが来たら、現在の他プレイヤーの位置をそのクライアントに送信する
	for playerID, player := range c.game.GetPlayers() {
		if playerID == PlayerID(cl.ID()) {
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

		err = c.broker.Send(cl.ID(), "player_state", payload)
		if err != nil {
			return errors.Wrap(err, "failed to send player state")
		}
	}

	return nil
}

func (c *Controller) OnPublished(cl Client, pk *packets.PublishPacket) error {
	switch pk.TopicName {
	case "player_state":
		return c.onReceivePlayerState(cl, pk)
	default:
		return errors.New(fmt.Sprintf("invalid topic name: %s", pk.TopicName))
	}
}

func (c *Controller) OnDisconnected(cl Client) error {
	slog.Info("client disconnected", "client_id", cl.ID())
	c.broker.RemoveClient(cl)
	c.game.RemovePlayer(PlayerID(cl.ID()))

	playerState := &shared.PlayerState{
		PlayerId: cl.ID(),
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
func (c *Controller) onReceivePlayerState(cl Client, pk *packets.PublishPacket) error {
	playerID := PlayerID(cl.ID())
	playerState := &shared.PlayerState{}
	err := proto.Unmarshal(pk.Payload, playerState)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal player state")
	}
	position := &Position{
		X: int(playerState.GetPosition().GetX()),
		Y: int(playerState.GetPosition().GetY()),
	}
	c.game.UpdatePlayerPosition(playerID, position)

	playerState.Status = shared.Status_ALIVE
	payload, err := proto.Marshal(playerState)
	if err != nil {
		return errors.Wrap(err, "failed to marshal player state")
	}
	c.broker.Broadcast("player_state", payload)

	slog.Info("all players", "players", c.game.String())

	return nil
}
