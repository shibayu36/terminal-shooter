package game

import (
	"fmt"

	"github.com/shibayu36/terminal-shooter/shared"
)

type ItemType string

const (
	ItemTypeBullet   ItemType = "bullet"
	ItemTypeBomb     ItemType = "bomb"
	ItemTypeBombFire ItemType = "bomb_fire"
)

// ToSharedItemType ItemTypeをshared.ItemTypeに変換する
func (t ItemType) ToSharedItemType() shared.ItemType {
	switch t {
	case ItemTypeBullet:
		return shared.ItemType_BULLET
	case ItemTypeBomb:
		return shared.ItemType_BOMB
	case ItemTypeBombFire:
		return shared.ItemType_BOMB_FIRE
	default:
		panic(fmt.Sprintf("invalid item type: %s", t))
	}
}

type Item interface {
	collidable
	ID() ItemID
	Type() ItemType
	Position() Position
	Update(provider gameOperationProvider) bool
}
