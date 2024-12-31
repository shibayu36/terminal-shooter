package main

import (
	"net"
	"sync"
)

// Client represents a connected MQTT client
type Client struct {
	ID      string
	Conn    net.Conn
	sendMux sync.Mutex
}
