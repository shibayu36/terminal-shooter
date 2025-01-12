package main

import (
	"sync"
	"time"
)

// MessageStats はメッセージの配信レートを計算・管理する構造体
type MessageStats struct {
	count    int
	lastTime time.Time
	rate     float64

	mu sync.RWMutex `exhaustruct:"optional"`
}

// メッセージを受信したことを記録する
func (m *MessageStats) RecordMessage() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.count++
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
		m.count = 0
		m.lastTime = now
	}
}

// 現在のレートを取得する
func (m *MessageStats) Rate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rate
}
