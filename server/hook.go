package main

import (
	"bytes"
	"fmt"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

// MQTTのパケット送受信のフックを実装する
// 実質ハンドラーとしての役割を持つ

type HookOptions struct{}

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
	return nil
}

func (h *Hook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	if err != nil {
		h.Log.Info("client disconnected", "client", cl.ID, "expire", expire, "error", err)
	} else {
		h.Log.Info("client disconnected", "client", cl.ID, "expire", expire)
	}

}

func (h *Hook) OnSubscribed(cl *mqtt.Client, pk packets.Packet, reasonCodes []byte) {
	h.Log.Info(fmt.Sprintf("subscribed qos=%v", reasonCodes), "client", cl.ID, "filters", pk.Filters)
}

func (h *Hook) OnUnsubscribed(cl *mqtt.Client, pk packets.Packet) {
	h.Log.Info("unsubscribed", "client", cl.ID, "filters", pk.Filters)
}

func (h *Hook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.Log.Info("received from client", "client", cl.ID, "payload", string(pk.Payload))

	pkx := pk
	if string(pk.Payload) == "hello" {
		pkx.Payload = []byte("hello world")
		h.Log.Info("received modified packet from client", "client", cl.ID, "payload", string(pkx.Payload))
	}

	return pkx, nil
}

func (h *Hook) OnPublished(cl *mqtt.Client, pk packets.Packet) {
	h.Log.Info("published to client", "client", cl.ID, "payload", string(pk.Payload))
}
