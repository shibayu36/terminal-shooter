package main

import (
	"log/slog"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/shibayu36/terminal-shooter/shared"

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

func (c *Controller) OnConnected(cl *Client, pk *packets.ConnectPacket) error {
	c.game.AddPlayer(PlayerID(cl.ID), &PlayerState{Position: &Position{X: 0, Y: 0}})

	// Player状態を出力
	slog.Info("all players", "players", c.game.String())

	return nil
}

func (c *Controller) OnSubscribed(cl *Client, pk *packets.SubscribePacket) error {
	// Subscribeが来たら、現在の他プレイヤーの位置をそのクライアントに送信する
	for playerID, player := range c.game.GetPlayers() {
		if playerID == PlayerID(cl.ID) {
			continue
		}

		playerState := &shared.PlayerState{
			PlayerId: string(playerID),
			Position: &shared.Position{
				X: int32(player.Position.X),
				Y: int32(player.Position.Y),
			},
		}
		payload, err := proto.Marshal(playerState)
		if err != nil {
			slog.Error("failed to marshal player state", "error", err)
			return err
		}

		slog.Info("send player state on subscribe", "player", playerState.String())

		err = c.broker.Send(cl.ID, "player_state", payload)
		if err != nil {
			slog.Error("failed to send player state", "error", err)
			return err
		}
	}

	return nil
}

func (c *Controller) OnPublished(cl *Client, pk *packets.PublishPacket) error {
	switch pk.TopicName {
	case "player_state":
		return c.onReceivePlayerState(cl, pk)
	default:
		slog.Error("invalid topic name", "topic", pk.TopicName)
	}

	return nil
}

func (c *Controller) OnDisconnected(cl *Client, pk *packets.DisconnectPacket) error {
	slog.Info("client disconnected", "client_id", cl.ID)
	c.game.RemovePlayer(PlayerID(cl.ID))

	// TODO: Disconnectしたら、他のクライアントにそのクライアントが消えたことを通知する

	return nil
}

// player_stateパケットを受信した時の処理
func (c *Controller) onReceivePlayerState(cl *Client, pk *packets.PublishPacket) error {
	playerID := PlayerID(cl.ID)
	playerState := &shared.PlayerState{}
	err := proto.Unmarshal(pk.Payload, playerState)
	if err != nil {
		slog.Error("failed to unmarshal player state", "error", err)
		return err
	}
	position := &Position{X: int(playerState.Position.X), Y: int(playerState.Position.Y)}
	c.game.UpdatePlayerPosition(playerID, position)

	c.broker.Broadcast("player_state", pk.Payload)

	slog.Info("all players", "players", c.game.String())

	return nil
}
