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

	// 全てのアイテムがBombFireである
	for _, item := range items {
		assert.Equal(t, ItemTypeBombFire, item.Type())
	}

	positions := make(map[Position]bool)
	for _, item := range items {
		positions[item.Position()] = true
	}

	// BombFireが4x4で設置されている
	assert.True(t, positions[Position{X: 5, Y: 8}], "BombFireは中心が5,8である")
	// 5,8を中心に4x4の範囲に火が出ている
	// 上
	assert.True(t, positions[Position{X: 5, Y: 9}])
	assert.True(t, positions[Position{X: 5, Y: 10}])
	assert.True(t, positions[Position{X: 5, Y: 11}])
	assert.True(t, positions[Position{X: 5, Y: 12}])
	// 下
	assert.True(t, positions[Position{X: 5, Y: 7}])
	assert.True(t, positions[Position{X: 5, Y: 6}])
	assert.True(t, positions[Position{X: 5, Y: 5}])
	assert.True(t, positions[Position{X: 5, Y: 4}])
	// 左
	assert.True(t, positions[Position{X: 4, Y: 8}])
	assert.True(t, positions[Position{X: 3, Y: 8}])
	assert.True(t, positions[Position{X: 2, Y: 8}])
	assert.True(t, positions[Position{X: 1, Y: 8}])
	// 右
	assert.True(t, positions[Position{X: 6, Y: 8}])
	assert.True(t, positions[Position{X: 7, Y: 8}])
	assert.True(t, positions[Position{X: 8, Y: 8}])
	assert.True(t, positions[Position{X: 9, Y: 8}])

	// その後59TickまではBombFireが維持
	for i := 1; i <= 59; i++ {
		for _, item := range items {
			assert.False(t, item.Update(game))
		}
	}

	// 60Tick目にはBombFireが消える
	for _, item := range items {
		assert.True(t, item.Update(game))
	}
	assert.Empty(t, game.GetItems())
}
