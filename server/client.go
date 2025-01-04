package main

import (
	"net"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Client represents a connected MQTT client
type Client interface {
	ID() string
	Publish(publishPacket *packets.PublishPacket) error
}

type client struct {
	id      string `exhaustruct:"optional"` // idは後から設定される
	conn    net.Conn
	sendMux sync.Mutex `exhaustruct:"optional"`
}

var _ Client = (*client)(nil)

func (c *client) ID() string {
	return c.id
}

// Publish クライアントに対してPublishパケットを送信する
func (c *client) Publish(publishPacket *packets.PublishPacket) error {
	c.sendMux.Lock()
	defer c.sendMux.Unlock()

	err := publishPacket.Write(c.conn)
	if err != nil {
		return errors.Wrapf(err, "failed to write publish packet to client: %s", c.id)
	}

	return nil
}
