package main

import (
	"log"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/shibayu36/terminal-shooter/shared"

	"google.golang.org/protobuf/proto"
)

// MQTTのパケット送受信のフックを実装する
// 実質ハンドラーとしての役割を持つ

type Hook struct {
	broker *Broker
	game   *GameState
}

var _ Hooker = &Hook{}

func NewHook(broker *Broker, game *GameState) *Hook {
	return &Hook{broker: broker, game: game}
}

func (h *Hook) OnConnect(cl *Client, pk *packets.ConnectPacket) error {
	h.game.AddPlayer(PlayerID(cl.ID), &PlayerState{Position: &Position{X: 0, Y: 0}})

	// Player状態を出力
	log.Printf("all players: %s", h.game.String())

	return nil
}

func (h *Hook) OnSubscribe(cl *Client, pk *packets.SubscribePacket) error {
	return nil
}

func (h *Hook) OnPublish(cl *Client, pk *packets.PublishPacket) error {
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

func (h *Hook) OnDisconnect(cl *Client, pk *packets.DisconnectPacket) error {
	log.Printf("client disconnected: %s", cl.ID)
	h.game.RemovePlayer(PlayerID(cl.ID))

	return nil
}
