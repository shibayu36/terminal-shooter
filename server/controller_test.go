package main

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	id        string
	published []*packets.PublishPacket
}

func (c *mockClient) ID() string {
	return c.id
}

func (c *mockClient) Publish(publishPacket *packets.PublishPacket) error {
	c.published = append(c.published, publishPacket)
	return nil
}

func TestController_OnConnected(t *testing.T) {
	broker := NewBroker()
	state := NewGameState()
	controller := NewController(broker, state)

	cl1 := &mockClient{id: "id1"}
	controller.OnConnected(cl1, nil)
	assert.Equal(t, state.GetPlayers()[PlayerID("id1")], &PlayerState{Position: &Position{X: 0, Y: 0}}, "cl1が追加された")
	assert.Len(t, cl1.published, 0, "cl1にはメッセージが送信されていない")

	cl2 := &mockClient{id: "id2"}
	controller.OnConnected(cl2, nil)
	assert.Equal(t, state.GetPlayers()[PlayerID("id2")], &PlayerState{Position: &Position{X: 0, Y: 0}}, "cl2が追加された")
	assert.Len(t, cl2.published, 0, "cl2にはメッセージが送信されていない")
}
