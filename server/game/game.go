package game

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shibayu36/terminal-shooter/server/stats"
	"github.com/shibayu36/terminal-shooter/shared"
)

// 1つのゲーム内の状態を管理する
type Game struct {
	Width  int
	Height int

	Players map[PlayerID]*Player
	Items   map[ItemID]Item

	// 削除されたアイテムを管理する
	RemovedItems map[ItemID]Item

	mu sync.RWMutex `exhaustruct:"optional"`
}

func NewGame(width, height int) *Game {
	return &Game{
		Width:        width,
		Height:       height,
		Players:      make(map[PlayerID]*Player),
		Items:        make(map[ItemID]Item),
		RemovedItems: make(map[ItemID]Item),
	}
}

// ゲーム状態を更新するループを開始する
// アイテムが何らか更新されたことを通知するチャネルを返す
func (g *Game) StartUpdateLoop(ctx context.Context) <-chan struct{} {
	itemsUpdatedCh := make(chan struct{})

	go func() {
		defer close(itemsUpdatedCh)

		ticker := time.NewTicker(16700 * time.Microsecond) // 16.7ms
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				start := time.Now()
				g.update(itemsUpdatedCh)
				stats.GameLoopDuration.Observe(time.Since(start).Seconds())
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

	g.checkCollisions()

	if len(updatedItems) > 0 {
		updatedItemsCh <- struct{}{}
	}
}

func (g *Game) checkCollisions() {
	// プレイヤーとアイテムが衝突しているかどうかをチェックする
	itemPosMap := make(map[Position][]Item)
	for _, item := range g.GetItems() {
		itemPosMap[item.Position()] = append(itemPosMap[item.Position()], item)
	}

	for _, player := range g.GetPlayers() {
		for _, item := range itemPosMap[player.Position] {
			if bullet, ok := item.(*Bullet); ok {
				g.UpdatePlayerStatus(player.PlayerID, PlayerStatusDead)
				g.RemoveItem(bullet.ID())
			}
		}
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
	g.Players[playerID] = &Player{
		PlayerID:  playerID,
		Position:  Position{X: 0, Y: 0},
		Direction: DirectionUp,
		Status:    PlayerStatusAlive,
	}
}

// プレイヤーを削除する
func (g *Game) RemovePlayer(playerID PlayerID) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.Players, playerID)
}

// プレイヤーの位置を更新する
func (g *Game) MovePlayer(playerID PlayerID, position Position, direction Direction) *Player {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Players[playerID].Position = position
	g.Players[playerID].Direction = direction
	return g.Players[playerID]
}

// プレイヤーのステータスを更新する
func (g *Game) UpdatePlayerStatus(playerID PlayerID, status PlayerStatus) *Player {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Players[playerID].Status = status
	return g.Players[playerID]
}

// プレイヤー一覧を取得する
func (g *Game) GetPlayers() map[PlayerID]*Player {
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

// あるプレイヤーから弾を発射する
// TODO: 追加した時に更新通知する必要がある
func (g *Game) ShootBullet(playerID PlayerID) ItemID {
	g.mu.Lock()
	defer g.mu.Unlock()

	player, ok := g.Players[playerID]
	if !ok {
		return ItemID("")
	}

	// プレイヤーの前方に発射する
	position := player.FowardPosition()
	direction := player.Direction

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
