package main

import (
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MessageStats はメッセージの配信レートとバイト数を計算・管理する構造体
type MessageStats struct {
	count     int
	bytes     int64
	lastTime  time.Time
	rate      float64
	bytesRate float64

	mu sync.RWMutex `exhaustruct:"optional"`
}

// 新しいMessageStats構造体を作成
func NewMessageStats() *MessageStats {
	return &MessageStats{
		count:     0,
		bytes:     0,
		lastTime:  time.Now(),
		rate:      0,
		bytesRate: 0,
	}
}

// メッセージを受信したことを記録する
func (m *MessageStats) RecordMessage(msg mqtt.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.count++
	m.bytes += int64(len(msg.Payload()))
}

// メッセージレートを計算する
func (m *MessageStats) Calculate() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	if m.lastTime.IsZero() {
		m.lastTime = now
		return
	}

	duration := now.Sub(m.lastTime).Seconds()
	if duration >= 1.0 { // 1秒以上経過している場合に計算
		m.rate = float64(m.count) / duration
		m.bytesRate = float64(m.bytes) / duration
		m.count = 0
		m.bytes = 0
		m.lastTime = now
	}
}

// 現在のメッセージレートを取得する
func (m *MessageStats) Rate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rate
}

// 現在のバイトレートを取得する（bytes/sec）
func (m *MessageStats) BytesRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bytesRate
}
