package game

type ItemType string

const (
	ItemTypeBullet ItemType = "bullet"
	ItemTypeBomb   ItemType = "bomb"
)

type Item interface {
	collidable
	ID() ItemID
	Type() ItemType
	Position() Position
	Update(provider gameOperationProvider) bool
}
