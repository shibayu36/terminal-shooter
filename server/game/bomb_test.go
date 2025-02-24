package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Bomb(t *testing.T) {
	game := NewGame(30, 30)

	bomb := NewBomb(ItemID("bomb1"), Position{X: 5, Y: 8})
	game.addItem(bomb)

	assert.Equal(t, ItemID("bomb1"), bomb.ID())
	assert.Equal(t, ItemTypeBomb, bomb.Type())
	assert.Equal(t, Position{X: 5, Y: 8}, bomb.Position(), "ボムを設置できた")

	// 179回更新時はBombのまま
	for i := 1; i <= 179; i++ {
		assert.False(t, bomb.Update(game), "更新後もBombである")
	}

	// 180回目に状態更新される
	assert.True(t, bomb.Update(game), "更新後はBombでない")

	// BombFireが4x4で設置されている
	items := game.GetItems()
	assert.Len(t, items, 17)
}
