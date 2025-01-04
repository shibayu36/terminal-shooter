package main

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/shibayu36/terminal-shooter/shared"
)

type (
	GameID   string
	PlayerID string
)

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
// 全てデフォルトで初期化する
func (gs *GameState) AddPlayer(playerID PlayerID) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Players[playerID] = &PlayerState{
		PlayerID:  playerID,
		Position:  &Position{X: 0, Y: 0},
		Direction: DirectionUp,
	}
}

// プレイヤーを削除する
func (gs *GameState) RemovePlayer(playerID PlayerID) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	delete(gs.Players, playerID)
}

// プレイヤーの位置を更新する
func (gs *GameState) MovePlayer(playerID PlayerID, position *Position, direction Direction) *PlayerState {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Players[playerID].Position = position
	gs.Players[playerID].Direction = direction
	return gs.Players[playerID]
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
	PlayerID  PlayerID
	Position  *Position
	Direction Direction
}

// プレイヤーの状態をshared.PlayerStateに変換する
func (ps *PlayerState) ToSharedPlayerState(status shared.Status) *shared.PlayerState {
	return &shared.PlayerState{
		PlayerId: string(ps.PlayerID),
		Position: &shared.Position{
			X: int32(ps.Position.X),
			Y: int32(ps.Position.Y),
		},
		Direction: ps.Direction.ToSharedDirection(),
		Status:    status,
	}
}

// 位置を管理する
type Position struct {
	X int
	Y int
}

// 向き
type Direction string

const (
	DirectionUp    Direction = "up"
	DirectionDown  Direction = "down"
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

// Directionをshared.Directionに変換する
func (d Direction) ToSharedDirection() shared.Direction {
	switch d {
	case DirectionUp:
		return shared.Direction_UP
	case DirectionDown:
		return shared.Direction_DOWN
	case DirectionLeft:
		return shared.Direction_LEFT
	case DirectionRight:
		return shared.Direction_RIGHT
	default:
		panic(fmt.Sprintf("invalid direction: %s", d))
	}
}

// shared.DirectionをDirectionに変換する
func FromSharedDirection(d shared.Direction) (Direction, error) {
	switch d {
	case shared.Direction_UP:
		return DirectionUp, nil
	case shared.Direction_DOWN:
		return DirectionDown, nil
	case shared.Direction_LEFT:
		return DirectionLeft, nil
	case shared.Direction_RIGHT:
		return DirectionRight, nil
	default:
		return "", errors.Newf("invalid direction: %d", d)
	}
}
