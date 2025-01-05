package game

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/shibayu36/terminal-shooter/shared"
)

type (
	//nolint:revive
	GameID   string
	PlayerID string
	ItemID   string
)

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

// 方向を、dxとdyのベクトルに変換する
func (d Direction) ToVector() (dx int, dy int) {
	switch d {
	case DirectionUp:
		dx, dy = 0, -1
	case DirectionDown:
		dx, dy = 0, 1
	case DirectionLeft:
		dx, dy = -1, 0
	case DirectionRight:
		dx, dy = 1, 0
	}
	return
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
