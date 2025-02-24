package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Bomb(t *testing.T) {
	bomb := NewBomb(ItemID("bomb1"), Position{X: 1, Y: 1})

	assert.Equal(t, ItemID("bomb1"), bomb.ID())
	assert.Equal(t, ItemTypeBomb, bomb.Type())
	assert.Equal(t, Position{X: 1, Y: 1}, bomb.Position())
}
