package main

import "sync"

// 1つのゲーム内の状態を管理する
type GameState struct {
	mu sync.RWMutex

	Players map[string]*PlayerState
}

func NewGameState() *GameState {
	return &GameState{
		Players: make(map[string]*PlayerState),
	}
}

// プレイヤーを追加する
func (gs *GameState) AddPlayer(playerID string, state *PlayerState) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Players[playerID] = state
}

// プレイヤーを削除する
func (gs *GameState) RemovePlayer(playerID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	delete(gs.Players, playerID)
}

// プレイヤーの位置を更新する
func (gs *GameState) UpdatePlayerPosition(playerID string, position *Position) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Players[playerID].Position = position
}

// プレイヤー一覧を取得する
func (gs *GameState) GetPlayers() map[string]*PlayerState {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.Players
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
