package main

import (
	"bytes"
	"fmt"

	"github.com/shibayu36/terminal-shooter/shared"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"google.golang.org/protobuf/proto"
)

// MQTTのパケット送受信のフックを実装する
// 実質ハンドラーとしての役割を持つ

type HookOptions struct {
	game *GameState
}

type Hook struct {
	mqtt.HookBase
	config *HookOptions
}

func (h *Hook) ID() string {
	return "events-example"
}

func (h *Hook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnect,
		mqtt.OnDisconnect,
		mqtt.OnPublished,
		mqtt.OnPublish,
	}, []byte{b})
}

func (h *Hook) Init(config any) error {
	h.Log.Info("initialised")
	if _, ok := config.(*HookOptions); !ok && config != nil {
		return mqtt.ErrInvalidConfigType
	}

	h.config = config.(*HookOptions)
	return nil
}

func (h *Hook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	h.Log.Info("client connected", "client", cl.ID)
	h.config.game.AddPlayer(PlayerID(cl.ID), &PlayerState{Position: &Position{X: 0, Y: 0}})
	return nil
}

func (h *Hook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	if err != nil {
		h.Log.Info("client disconnected", "client", cl.ID, "expire", expire, "error", err)
	} else {
		h.Log.Info("client disconnected", "client", cl.ID, "expire", expire)
	}
	h.config.game.RemovePlayer(PlayerID(cl.ID))
}

func (h *Hook) OnSubscribed(cl *mqtt.Client, pk packets.Packet, reasonCodes []byte) {
	h.Log.Info(fmt.Sprintf("subscribed qos=%v", reasonCodes), "client", cl.ID, "filters", pk.Filters)
}

func (h *Hook) OnUnsubscribed(cl *mqtt.Client, pk packets.Packet) {
	h.Log.Info("unsubscribed", "client", cl.ID, "filters", pk.Filters)
}

func (h *Hook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.Log.Info("received from client", "client", cl.ID)

	if pk.TopicName == "player_state" {
		playerID := PlayerID(cl.ID)
		playerState := &shared.PlayerState{}
		err := proto.Unmarshal(pk.Payload, playerState)
		if err != nil {
			h.Log.Error("failed to unmarshal player state", "error", err)
			return pk, err
		}
		position := &Position{X: int(playerState.Position.X), Y: int(playerState.Position.Y)}
		h.config.game.UpdatePlayerPosition(playerID, position)

		h.Log.Info("all players", "players", h.config.game.String())
	} else {
		h.Log.Error("invalid topic name", "topic", pk.TopicName)
	}

	return pk, nil
}

func (h *Hook) OnPublished(cl *mqtt.Client, pk packets.Packet) {
	h.Log.Info("published to client", "client", cl.ID, "payload", string(pk.Payload))
}
