package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/eclipse/paho.mqtt.golang/packets"
)

type Hooker interface {
	OnConnected(client Client, packet *packets.ConnectPacket) error
	OnPublished(client Client, packet *packets.PublishPacket) error
	OnSubscribed(client Client, packet *packets.SubscribePacket) error
	OnDisconnected(client Client) error
}

// Server represents the MQTT server
type Server struct {
	listener net.Listener
	hook     Hooker

	// サーバーの終了のため
	activeConn map[net.Conn]struct{}
	inShutdown atomic.Bool
	wg         sync.WaitGroup
}

func NewServer(address string, hook Hooker) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to listen")
	}

	return &Server{
		listener: listener,
		hook:     hook,

		activeConn: make(map[net.Conn]struct{}),
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

			return errors.Wrap(err, "failed to accept connection")
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) Shutdown(timeout time.Duration) error {
	slog.Info("Shutting down server...")
	s.inShutdown.Store(true)

	if err := s.listener.Close(); err != nil {
		return errors.Wrap(err, "error closing listener")
	}

	for conn := range s.activeConn {
		slog.Info("Closing connection", "address", conn.RemoteAddr())
		conn.Close()
	}

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
		return errors.New("shutdown timed out")
	}

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	s.activeConn[conn] = struct{}{}

	client := &client{
		conn: conn,
	}

	defer func() {
		err := s.hook.OnDisconnected(client)
		if err != nil {
			slog.Error("Error on disconnected", "error", fmt.Sprintf("%+v", err))
		}
		conn.Close()
		delete(s.activeConn, conn)
		s.wg.Done()
	}()

	slog.Info("New client connected", "address", conn.RemoteAddr())

	for {
		packet, err := packets.ReadPacket(conn)
		if err != nil {
			if s.inShutdown.Load() {
				return
			}

			if errors.Is(err, io.EOF) {
				slog.Info("Client disconnected", "address", conn.RemoteAddr())
				return
			}

			slog.Error("Error reading packet",
				"error", fmt.Sprintf("%+v", err),
			)
			return
		}

		if err := s.handlePacket(client, packet); err != nil {
			// packet一つのハンドリングを失敗しただけなら、そのパケットを破棄して続ける
			slog.Error("Error handling packet",
				"error", fmt.Sprintf("%+v", err),
			)
		}
	}
}

func (s *Server) handlePacket(client *client, packet packets.ControlPacket) error {
	switch p := packet.(type) {
	case *packets.ConnectPacket:
		return s.handleConnect(client, p)
	case *packets.PublishPacket:
		return s.handlePublish(client, p)
	case *packets.SubscribePacket:
		return s.handleSubscribe(client, p)
	case *packets.PingreqPacket:
		pingresp := packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket)
		return pingresp.Write(client.conn)
	case *packets.DisconnectPacket:
		// 何もしない
		return nil
	default:
		// サポートしていないパケットは無視
		return nil
	}
}

// handleConnect handles CONNECT packets
func (s *Server) handleConnect(client *client, cp *packets.ConnectPacket) error {
	// CONNACK パケットの作成と送信
	connack := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
	connack.ReturnCode = packets.Accepted
	connack.SessionPresent = false

	if err := connack.Write(client.conn); err != nil {
		return errors.Wrap(err, "failed to write CONNACK")
	}

	// クライアントの登録
	client.id = cp.ClientIdentifier

	if err := s.hook.OnConnected(client, cp); err != nil {
		return err
	}

	return nil
}

// handlePublish handles PUBLISH packets
func (s *Server) handlePublish(client *client, pp *packets.PublishPacket) error {
	slog.Info("Received publish packet", "topic", pp.TopicName)

	if err := s.hook.OnPublished(client, pp); err != nil {
		return err
	}

	return nil
}

// handleSubscribe handles SUBSCRIBE packets
func (s *Server) handleSubscribe(client *client, sp *packets.SubscribePacket) error {
	ack := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
	ack.MessageID = sp.MessageID
	ack.ReturnCodes = make([]byte, len(sp.Topics)) // QoS=0 only
	if err := ack.Write(client.conn); err != nil {
		return errors.Wrap(err, "failed to write suback packet")
	}

	if err := s.hook.OnSubscribed(client, sp); err != nil {
		return err
	}

	return nil
}
