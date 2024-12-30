package main

import (
	"log"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

type Broker struct {
	// クライアント管理
	clients    map[string]*Client
	clientsMux sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[string]*Client),
	}
}

func (b *Broker) AddClient(client *Client) {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	b.clients[client.ID] = client
}

func (b *Broker) RemoveClient(client *Client) {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	delete(b.clients, client.ID)
}

func (b *Broker) CloseAll() {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	for _, client := range b.clients {
		client.Conn.Close()
	}
}

// Broadcast クライアント全員にメッセージを配信する
func (b *Broker) Broadcast(topic string, payload []byte) {
	b.clientsMux.RLock()
	defer b.clientsMux.RUnlock()

	for _, client := range b.clients {
		publishPacket := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
		publishPacket.TopicName = topic
		publishPacket.Payload = payload
		publishPacket.Qos = 0

		client.sendMux.Lock()
		err := publishPacket.Write(client.Conn)
		client.sendMux.Unlock()

		if err != nil {
			log.Printf("Error sending to subscriber %s: %v\n", client.ID, err)
		}
	}
}
