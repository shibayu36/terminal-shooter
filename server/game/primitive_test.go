package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Direction_ToVector(t *testing.T) {
	testcases := []struct {
		direction Direction
		dx        int
		dy        int
	}{
		{DirectionUp, 0, -1},
		{DirectionDown, 0, 1},
		{DirectionLeft, -1, 0},
		{DirectionRight, 1, 0},
		{Direction("invalid"), 0, 0},
	}

	for _, tc := range testcases {
		t.Run(string(tc.direction), func(t *testing.T) {
			dx, dy := tc.direction.ToVector()
			assert.Equal(t, dx, tc.dx)
			assert.Equal(t, dy, tc.dy)
		})
	}
}
