package game

type ItemType string

const (
	ItemTypeBullet ItemType = "bullet"
)

type Item interface {
	Collidable
	ID() ItemID
	Type() ItemType
	Position() Position
	Update() (updated bool)
}
