package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Server represents the MQTT server
type Server struct {
	listener net.Listener

	// クライアント管理
	clients    map[string]*Client
	clientsMux sync.RWMutex

	// サーバーの終了待ち
	inShutdown atomic.Bool
	wg         sync.WaitGroup
}

// Client represents a connected MQTT client
type Client struct {
	ID      string
	Conn    net.Conn
	sendMux sync.Mutex
}

func NewServer(address string) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener: listener,
		clients:  make(map[string]*Client),
	}, nil
}

func (s *Server) Serve() error {
	log.Printf("MQTT Server listening on %s\n", s.listener.Addr())

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.inShutdown.Load() {
				return nil
			}

			log.Printf("Failed to accept connection: %v", err)
			return err
		}

		client := &Client{
			Conn: conn,
		}

		s.wg.Add(1)
		go s.handleConnection(client)
	}
}

func (s *Server) Shutdown(timeout time.Duration) error {
	log.Println("Shutting down server...")
	s.inShutdown.Store(true)

	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("error closing listener: %w", err)
	}

	// すべての接続を閉じる
	s.clientsMux.Lock()
	for _, client := range s.clients {
		client.Conn.Close()
	}
	s.clientsMux.Unlock()

	// タイムアウト付きでgoroutineの終了を待つ
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Server shutdown complete")
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timed out")
	}

	return nil
}

func (s *Server) handleConnection(client *Client) {
	defer func() {
		s.removeClient(client)
		client.Conn.Close()
		s.wg.Done()
	}()

	log.Printf("New client connected: %s\n", client.Conn.RemoteAddr())

	for {
		packet, err := packets.ReadPacket(client.Conn)
		if err != nil {
			if s.inShutdown.Load() {
				return
			}

			if err == io.EOF {
				log.Printf("Client disconnected: %s\n", client.Conn.RemoteAddr())
				return
			}

			log.Printf("Error reading packet: %v\n", err)
			return
		}

		if err := s.handlePacket(client, packet); err != nil {
			log.Printf("Error handling packet: %v\n", err)
			return
		}
	}
}

func (s *Server) handlePacket(client *Client, packet packets.ControlPacket) error {
	switch p := packet.(type) {
	case *packets.ConnectPacket:
		return s.handleConnect(client, p)
	case *packets.PublishPacket:
		return s.handlePublish(client, p)
	case *packets.SubscribePacket:
		return s.handleSubscribe(client, p)
	case *packets.PingreqPacket:
		pingresp := packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket)
		return pingresp.Write(client.Conn)
	case *packets.DisconnectPacket:
		return s.handleDisconnect(client, p)
	default:
		// サポートしていないパケットは無視
		return nil
	}
}

// handleConnect handles CONNECT packets
func (s *Server) handleConnect(client *Client, cp *packets.ConnectPacket) error {
	// CONNACK パケットの作成と送信
	connack := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
	connack.ReturnCode = packets.Accepted
	connack.SessionPresent = false

	if err := connack.Write(client.Conn); err != nil {
		return fmt.Errorf("failed to write CONNACK: %v", err)
	}

	// クライアントの登録
	client.ID = cp.ClientIdentifier

	s.addClient(client)

	return nil
}

// handlePublish handles PUBLISH packets
func (s *Server) handlePublish(client *Client, pp *packets.PublishPacket) error {
	log.Printf("Received publish packet: %+v\n", pp)

	// 購読者全員にメッセージを配信
	for _, client := range s.clients {
		publishPacket := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
		publishPacket.TopicName = pp.TopicName
		publishPacket.Payload = pp.Payload
		publishPacket.Qos = pp.Qos

		client.sendMux.Lock()
		err := publishPacket.Write(client.Conn)
		client.sendMux.Unlock()

		if err != nil {
			log.Printf("Error sending to subscriber %s: %v\n", client.ID, err)
		}
	}

	return nil
}

// handleSubscribe handles SUBSCRIBE packets
func (s *Server) handleSubscribe(client *Client, sp *packets.SubscribePacket) error {
	ack := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
	ack.MessageID = sp.MessageID
	ack.ReturnCodes = make([]byte, len(sp.Topics)) // QoS=0 only
	if err := ack.Write(client.Conn); err != nil {
		return fmt.Errorf("failed to write suback packet: %w", err)
	}

	return nil
}

// handleDisconnect handles DISCONNECT packets
func (s *Server) handleDisconnect(client *Client, dp *packets.DisconnectPacket) error {
	s.removeClient(client)
	return nil
}

func (s *Server) addClient(client *Client) {
	s.clientsMux.Lock()
	s.clients[client.ID] = client
	s.clientsMux.Unlock()
}

func (s *Server) removeClient(client *Client) {
	s.clientsMux.Lock()
	delete(s.clients, client.ID)
	s.clientsMux.Unlock()
}
