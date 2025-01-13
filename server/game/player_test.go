package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Player_FowardPosition(t *testing.T) {
	testcases := []struct {
		position  Position
		direction Direction
		expected  Position
	}{
		{Position{X: 3, Y: 8}, DirectionRight, Position{X: 4, Y: 8}},
		{Position{X: 1, Y: 2}, DirectionLeft, Position{X: 0, Y: 2}},
		{Position{X: 5, Y: 6}, DirectionUp, Position{X: 5, Y: 5}},
		{Position{X: 7, Y: 8}, DirectionDown, Position{X: 7, Y: 9}},
	}

	for _, tc := range testcases {
		player := &Player{Position: tc.position, Direction: tc.direction, Status: PlayerStatusAlive}
		assert.Equal(t, tc.expected, player.FowardPosition())
	}
}
