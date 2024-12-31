package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

type Hooker interface {
	OnConnected(client *Client, packet *packets.ConnectPacket) error
	OnPublished(client *Client, packet *packets.PublishPacket) error
	OnSubscribed(client *Client, packet *packets.SubscribePacket) error
	OnDisconnected(client *Client, packet *packets.DisconnectPacket) error
}

// Server represents the MQTT server
type Server struct {
	listener net.Listener
	hook     Hooker

	// クライアント管理
	broker *Broker

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

func NewServer(address string, hook Hooker, broker *Broker) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener: listener,
		hook:     hook,
		broker:   broker,
	}, nil
}

func (s *Server) Serve() error {
	slog.Info("MQTT Server listening", "address", s.listener.Addr())

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.inShutdown.Load() {
				return nil
			}

			slog.Error("Failed to accept connection", "error", err)
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
	slog.Info("Shutting down server...")
	s.inShutdown.Store(true)

	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("error closing listener: %w", err)
	}

	// すべての接続を閉じる
	s.broker.CloseAll()

	// タイムアウト付きでgoroutineの終了を待つ
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("Server shutdown complete")
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timed out")
	}

	return nil
}

func (s *Server) handleConnection(client *Client) {
	defer func() {
		s.broker.RemoveClient(client)
		client.Conn.Close()
		s.wg.Done()
	}()

	slog.Info("New client connected", "address", client.Conn.RemoteAddr())

	for {
		packet, err := packets.ReadPacket(client.Conn)
		if err != nil {
			if s.inShutdown.Load() {
				return
			}

			if err == io.EOF {
				slog.Info("Client disconnected", "address", client.Conn.RemoteAddr())
				return
			}

			slog.Error("Error reading packet", "error", err)
			return
		}

		if err := s.handlePacket(client, packet); err != nil {
			slog.Error("Error handling packet", "error", err)
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

	s.broker.AddClient(client)

	if err := s.hook.OnConnected(client, cp); err != nil {
		return err
	}

	return nil
}

// handlePublish handles PUBLISH packets
func (s *Server) handlePublish(client *Client, pp *packets.PublishPacket) error {
	slog.Info("Received publish packet", "topic", pp.TopicName)

	if err := s.hook.OnPublished(client, pp); err != nil {
		return err
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

	if err := s.hook.OnSubscribed(client, sp); err != nil {
		return err
	}

	return nil
}

// handleDisconnect handles DISCONNECT packets
func (s *Server) handleDisconnect(client *Client, dp *packets.DisconnectPacket) error {
	s.broker.RemoveClient(client)

	if err := s.hook.OnDisconnected(client, dp); err != nil {
		return err
	}

	return nil
}
