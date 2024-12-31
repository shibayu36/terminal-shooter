package main

import (
	"net"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Client represents a connected MQTT client
type Client struct {
	ID      string
	Conn    net.Conn
	sendMux sync.Mutex
}

// Publish クライアントに対してPublishパケットを送信する
func (c *Client) Publish(publishPacket *packets.PublishPacket) error {
	c.sendMux.Lock()
	defer c.sendMux.Unlock()

	return publishPacket.Write(c.Conn)
}
