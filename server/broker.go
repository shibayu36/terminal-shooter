package main

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

type Broker struct {
	// クライアント管理
	clients    map[string]Client
	clientsMux sync.RWMutex `exhaustruct:"optional"`
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[string]Client),
	}
}

func (b *Broker) AddClient(client Client) {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	b.clients[client.ID()] = client
}

func (b *Broker) RemoveClient(client Client) {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	delete(b.clients, client.ID())
}

// Broadcast クライアント全員にメッセージを配信する
func (b *Broker) Broadcast(topic string, payload []byte) {
	b.clientsMux.RLock()
	defer b.clientsMux.RUnlock()

	publishPacket := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	publishPacket.TopicName = topic
	publishPacket.Payload = payload
	publishPacket.Qos = 0

	for _, client := range b.clients {
		err := client.Publish(publishPacket)
		if err != nil {
			slog.Error("Error sending to subscriber", "client_id", client.ID(), "error", err)
		}
	}
}

// 特定のクライアントにメッセージを送信する
func (b *Broker) Send(clientID string, topic string, payload []byte) error {
	client := b.clients[clientID]
	if client == nil {
		return errors.New("client not found")
	}

	publishPacket := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	publishPacket.TopicName = topic
	publishPacket.Payload = payload
	publishPacket.Qos = 0

	return client.Publish(publishPacket)
}
