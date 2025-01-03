package main

import (
	"bytes"
	"fmt"
	"sync"
)

type GameID string
type PlayerID string

// 1つのゲーム内の状態を管理する
type GameState struct {
	mu sync.RWMutex `exhaustruct:"optional"`

	Players map[PlayerID]*PlayerState
}

func NewGameState() *GameState {
	return &GameState{
		Players: make(map[PlayerID]*PlayerState),
	}
}

// プレイヤーを追加する
func (gs *GameState) AddPlayer(playerID PlayerID, state *PlayerState) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Players[playerID] = state
}

// プレイヤーを削除する
func (gs *GameState) RemovePlayer(playerID PlayerID) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	delete(gs.Players, playerID)
}

// プレイヤーの位置を更新する
func (gs *GameState) UpdatePlayerPosition(playerID PlayerID, position *Position) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Players[playerID].Position = position
}

// プレイヤー一覧を取得する
func (gs *GameState) GetPlayers() map[PlayerID]*PlayerState {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.Players
}

// GetState ゲームの状態をデバッグ用に表示する
func (gs *GameState) String() string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	buf := bytes.NewBufferString("")
	for playerID, player := range gs.Players {
		fmt.Fprintf(buf, "Player: %s, Position: %v\n", string(playerID), player.Position)
	}

	return buf.String()
}

// プレイヤーの状態を管理する
type PlayerState struct {
	Position *Position
}

// 位置を管理する
type Position struct {
	X int
	Y int
}
