package game

import (
	"sync"

	"github.com/google/uuid"
)

const (
	BombExplosionTick = 180 // 3秒後に爆発
	BombFireDuration  = 60  // 1秒で消滅
	BombFireRange     = 4   // 爆発の範囲
)

// Bomb ボムを表す
type Bomb struct {
	id       ItemID
	position Position

	// 現在のtick
	tick int

	mu sync.RWMutex `exhaustruct:"optional"`
}

func NewBomb(id ItemID, position Position) *Bomb {
	return &Bomb{
		id:       id,
		position: position,
		tick:     0,
	}
}

func (b *Bomb) ID() ItemID {
	return b.id
}

func (b *Bomb) Type() ItemType {
	return ItemTypeBomb
}

func (b *Bomb) Position() Position {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.position
}

// Update 状態を更新する
func (b *Bomb) Update(provider gameOperationProvider) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tick++

	// 爆発するタイミングになったら
	if b.tick >= BombExplosionTick {
		// 爆発の範囲にBombFireを設置
		pos := b.position
		// 中心
		provider.addItem(NewBombFire(ItemID(uuid.New().String()), pos))

		// 上下左右
		for i := 1; i <= BombFireRange; i++ {
			// 上
			provider.addItem(NewBombFire(ItemID(uuid.New().String()), Position{X: pos.X, Y: pos.Y - i}))
			// 下
			provider.addItem(NewBombFire(ItemID(uuid.New().String()), Position{X: pos.X, Y: pos.Y + i}))
			// 左
			provider.addItem(NewBombFire(ItemID(uuid.New().String()), Position{X: pos.X - i, Y: pos.Y}))
			// 右
			provider.addItem(NewBombFire(ItemID(uuid.New().String()), Position{X: pos.X + i, Y: pos.Y}))
		}

		// ボム自体を削除
		provider.RemoveItem(b.id)
		return true
	}

	return false
}

// BombFire ボムの爆発による火を表す
type BombFire struct {
	id       ItemID
	position Position

	// 現在のtick
	tick int

	mu sync.RWMutex `exhaustruct:"optional"`
}

func NewBombFire(id ItemID, position Position) *BombFire {
	return &BombFire{
		id:       id,
		position: position,
		tick:     0,
	}
}

func (bf *BombFire) ID() ItemID {
	return bf.id
}

func (bf *BombFire) Type() ItemType {
	return ItemTypeBombFire
}

func (bf *BombFire) Position() Position {
	bf.mu.RLock()
	defer bf.mu.RUnlock()
	return bf.position
}

// Update 状態を更新する
func (bf *BombFire) Update(provider gameOperationProvider) bool {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	bf.tick++

	// 一定時間経過したら消滅
	if bf.tick >= BombFireDuration {
		provider.RemoveItem(bf.id)
		return true
	}

	return false
}

// OnCollideWith 他のオブジェクトと衝突した時の処理
func (b *Bomb) OnCollideWith(other collidable, provider gameOperationProvider) bool {
	return false
}

// OnCollideWith 他のオブジェクトと衝突した時の処理
func (bf *BombFire) OnCollideWith(other collidable, provider gameOperationProvider) bool {
	// 何かと当たったとしても何もしない
	return false
}
