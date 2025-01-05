package game

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/shibayu36/terminal-shooter/shared"
)

type (
	//nolint:revive
	GameID   string
	PlayerID string
	ItemID   string
)

// 1つのゲーム内の状態を管理する
type Game struct {
	Width  int
	Height int

	Players map[PlayerID]*PlayerState
	Items   map[ItemID]Item

	// 削除されたアイテムを管理する
	RemovedItems map[ItemID]Item

	mu sync.RWMutex `exhaustruct:"optional"`
}

func NewGame(width, height int) *Game {
	return &Game{
		Width:        width,
		Height:       height,
		Players:      make(map[PlayerID]*PlayerState),
		Items:        make(map[ItemID]Item),
		RemovedItems: make(map[ItemID]Item),
	}
}

// ゲーム状態を更新するループを開始する
// アイテムが何らか更新されたことを通知するチャネルを返す
func (g *Game) StartUpdateLoop(ctx context.Context) <-chan struct{} {
	ticker := time.NewTicker(16700 * time.Microsecond) // 16.7ms

	itemsUpdatedCh := make(chan struct{})
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				g.update(itemsUpdatedCh)
			case <-ctx.Done():
				return
			}
		}
	}()

	return itemsUpdatedCh
}

// ゲーム状態を更新する
func (g *Game) update(updatedItemsCh chan<- struct{}) {
	items := g.GetItems()

	updatedItems := []Item{}
	for _, item := range items {
		if item.Update() {
			updatedItems = append(updatedItems, item)
		}
	}
	for _, updatedItem := range updatedItems {
		// 盤面外に出たアイテムを削除する
		if !g.isWithinBounds(updatedItem) {
			g.RemoveItem(updatedItem.ID())
		}
	}

	if len(updatedItems) > 0 {
		updatedItemsCh <- struct{}{}
	}
}

// アイテムが盤面内にあるかどうかを判定する
func (g *Game) isWithinBounds(item Item) bool {
	pos := item.Position()
	return pos.X >= 0 && pos.X < g.Width && pos.Y >= 0 && pos.Y < g.Height
}

// プレイヤーを追加する
// 全てデフォルトで初期化する
func (g *Game) AddPlayer(playerID PlayerID) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Players[playerID] = &PlayerState{
		PlayerID:  playerID,
		Position:  Position{X: 0, Y: 0},
		Direction: DirectionUp,
	}
}

// プレイヤーを削除する
func (g *Game) RemovePlayer(playerID PlayerID) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.Players, playerID)
}

// プレイヤーの位置を更新する
func (g *Game) MovePlayer(playerID PlayerID, position Position, direction Direction) *PlayerState {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Players[playerID].Position = position
	g.Players[playerID].Direction = direction
	return g.Players[playerID]
}

// プレイヤー一覧を取得する
func (g *Game) GetPlayers() map[PlayerID]*PlayerState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return shared.CopyMap(g.Players)
}

// アイテム一覧を取得する
func (g *Game) GetItems() map[ItemID]Item {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return shared.CopyMap(g.Items)
}

// 削除されたアイテム一覧を取得する
func (g *Game) GetRemovedItems() map[ItemID]Item {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return shared.CopyMap(g.RemovedItems)
}

// 削除されたアイテムをクリアする
func (g *Game) ClearRemovedItem(itemID ItemID) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.RemovedItems, itemID)
}

// アイテムを削除する
func (g *Game) RemoveItem(itemID ItemID) {
	g.mu.Lock()
	defer g.mu.Unlock()
	item, ok := g.Items[itemID]
	if !ok {
		return
	}
	delete(g.Items, itemID)
	g.RemovedItems[itemID] = item
}

// 弾を追加する
func (g *Game) AddBullet(position Position, direction Direction) ItemID {
	g.mu.Lock()
	defer g.mu.Unlock()
	bullet := NewBullet(ItemID(uuid.New().String()), position, direction)
	g.Items[bullet.ID()] = bullet
	return bullet.ID()
}

// GetState ゲームの状態をデバッグ用に表示する
func (g *Game) String() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	buf := bytes.NewBufferString("")
	for playerID, player := range g.Players {
		fmt.Fprintf(buf, "Player: %s, Position: %v\n", string(playerID), player.Position)
	}

	return buf.String()
}

type ItemType string

const (
	ItemTypeBullet ItemType = "bullet"
)

type Item interface {
	ID() ItemID
	Type() ItemType
	Position() Position
	Update() (updated bool)
}

type Bullet struct {
	id        ItemID
	position  Position
	direction Direction
	// 何tickで動くか
	moveTick int

	// 現在のtick
	tick int

	mu sync.RWMutex `exhaustruct:"optional"`
}

var _ Item = (*Bullet)(nil)

func NewBullet(id ItemID, position Position, direction Direction) *Bullet {
	return &Bullet{
		id:        id,
		position:  position,
		direction: direction,
		moveTick:  30, // 60fpsで0.5秒
		tick:      0,
	}
}

func (b *Bullet) ID() ItemID {
	return b.id
}

func (b *Bullet) Type() ItemType {
	return ItemTypeBullet
}

func (b *Bullet) Position() Position {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.position
}

func (b *Bullet) Update() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tick++
	if b.tick >= b.moveTick {
		b.tick = 0
		switch b.direction {
		case DirectionUp:
			b.position.Y--
		case DirectionDown:
			b.position.Y++
		case DirectionLeft:
			b.position.X--
		case DirectionRight:
			b.position.X++
		}
		return true
	}
	return false
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
