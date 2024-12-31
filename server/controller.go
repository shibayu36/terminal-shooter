package main

import (
	"log"

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

func (h *Controller) OnConnect(cl *Client, pk *packets.ConnectPacket) error {
	h.game.AddPlayer(PlayerID(cl.ID), &PlayerState{Position: &Position{X: 0, Y: 0}})

	// Player状態を出力
	log.Printf("all players: %s", h.game.String())

	return nil
}

func (h *Controller) OnSubscribe(cl *Client, pk *packets.SubscribePacket) error {
	return nil
}

func (h *Controller) OnPublish(cl *Client, pk *packets.PublishPacket) error {
	if pk.TopicName == "player_state" {
		playerID := PlayerID(cl.ID)
		playerState := &shared.PlayerState{}
		err := proto.Unmarshal(pk.Payload, playerState)
		if err != nil {
			log.Printf("failed to unmarshal player state: %s", err)
			return err
		}
		position := &Position{X: int(playerState.Position.X), Y: int(playerState.Position.Y)}
		h.game.UpdatePlayerPosition(playerID, position)

		h.broker.Broadcast("player_state", pk.Payload)

		log.Printf("all players: %s", h.game.String())
	} else {
		log.Printf("invalid topic name: %s", pk.TopicName)
	}

	return nil
}

func (h *Controller) OnDisconnect(cl *Client, pk *packets.DisconnectPacket) error {
	log.Printf("client disconnected: %s", cl.ID)
	h.game.RemovePlayer(PlayerID(cl.ID))

	return nil
}
