package game

type ItemType string

const (
	ItemTypeBullet ItemType = "bullet"
)

type Item interface {
	collidable
	ID() ItemID
	Type() ItemType
	Position() Position
	Update() (updated bool)
}
