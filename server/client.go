package main

import (
	"net"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Client represents a connected MQTT client
type Client interface {
	ID() string
	Publish(publishPacket *packets.PublishPacket) error
}

type client struct {
	id      string
	conn    net.Conn
	sendMux sync.Mutex
}

var _ Client = (*client)(nil)

func (c *client) ID() string {
	return c.id
}

// Publish クライアントに対してPublishパケットを送信する
func (c *client) Publish(publishPacket *packets.PublishPacket) error {
	c.sendMux.Lock()
	defer c.sendMux.Unlock()

	return publishPacket.Write(c.conn)
}
