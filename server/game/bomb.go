package game

import (
	"sync"
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
func (b *Bomb) Update() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tick++
	return false
}

// OnCollideWith 他のオブジェクトと衝突した時の処理
func (b *Bomb) OnCollideWith(other collidable, service gameCollisionService) bool {
	return false
}
