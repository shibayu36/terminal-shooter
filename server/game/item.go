package game

type ItemType string

const (
	ItemTypeBullet   ItemType = "bullet"
	ItemTypeBomb     ItemType = "bomb"
	ItemTypeBombFire ItemType = "bomb_fire"
)

type Item interface {
	collidable
	ID() ItemID
	Type() ItemType
	Position() Position
	Update(provider gameOperationProvider) bool
}
