package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GameState(t *testing.T) {
	t.Run("プレイヤーを追加できる", func(t *testing.T) {
		gameState := NewGameState()

		// player1を追加
		gameState.AddPlayer("player1")
		assert.Equal(t, 0, gameState.GetPlayers()["player1"].Position.X)
		assert.Equal(t, 0, gameState.GetPlayers()["player1"].Position.Y)
		assert.Equal(t, DirectionUp, gameState.GetPlayers()["player1"].Direction)

		// player1の位置を更新
		gameState.MovePlayer("player1", &Position{X: 2, Y: 8}, DirectionRight)
		assert.Equal(t, 2, gameState.GetPlayers()["player1"].Position.X)
		assert.Equal(t, 8, gameState.GetPlayers()["player1"].Position.Y)
		assert.Equal(t, DirectionRight, gameState.GetPlayers()["player1"].Direction)

		// player2を追加
		gameState.AddPlayer("player2")
		assert.Len(t, gameState.GetPlayers(), 2)
		assert.Equal(t, 0, gameState.GetPlayers()["player2"].Position.X)
		assert.Equal(t, 0, gameState.GetPlayers()["player2"].Position.Y)
		assert.Equal(t, DirectionUp, gameState.GetPlayers()["player2"].Direction)

		// player1を削除
		gameState.RemovePlayer("player1")
		assert.Len(t, gameState.GetPlayers(), 1)
		assert.Equal(t, 0, gameState.GetPlayers()["player2"].Position.X)
	})

	t.Run("弾を追加できる", func(t *testing.T) {
		gameState := NewGameState()

		itemID1 := gameState.AddBullet(&Position{X: 3, Y: 8}, DirectionRight)
		assert.Equal(t, 1, len(gameState.Items))
		assert.Equal(t, ItemTypeBullet, gameState.Items[itemID1].Type())
		assert.Equal(t, &Position{X: 3, Y: 8}, gameState.Items[itemID1].Position())

		itemID2 := gameState.AddBullet(&Position{X: 1, Y: 2}, DirectionRight)
		assert.Equal(t, 2, len(gameState.Items))
		assert.Equal(t, ItemTypeBullet, gameState.Items[itemID2].Type())
		assert.Equal(t, &Position{X: 1, Y: 2}, gameState.Items[itemID2].Position())
	})
}

func Test_GameState_update(t *testing.T) {
	gameState := NewGameState()

	// 弾を追加
	bulletID1 := gameState.AddBullet(&Position{X: 3, Y: 8}, DirectionLeft)
	// 2回動かす
	gameState.update()
	gameState.update()

	// 弾をもう一つ追加
	bulletID2 := gameState.AddBullet(&Position{X: 1, Y: 2}, DirectionUp)

	// 28回動かすと、bullet1だけ動く
	for i := 0; i < 28; i++ {
		gameState.update()
	}
	assert.Equal(t, &Position{X: 2, Y: 8}, gameState.Items[bulletID1].Position())
	assert.Equal(t, &Position{X: 1, Y: 2}, gameState.Items[bulletID2].Position())

	// さらに2回動かすと、bullet2が動く
	gameState.update()
	gameState.update()
	assert.Equal(t, &Position{X: 2, Y: 8}, gameState.Items[bulletID1].Position())
	assert.Equal(t, &Position{X: 1, Y: 3}, gameState.Items[bulletID2].Position())
}

func Test_Bullet(t *testing.T) {
	bullet := NewBullet("bullet1", &Position{X: 3, Y: 8}, DirectionRight)

	assert.Equal(t, ItemID("bullet1"), bullet.ID())
	assert.Equal(t, ItemTypeBullet, bullet.Type())
	assert.Equal(t, &Position{X: 3, Y: 8}, bullet.Position())

	// 1回目は更新されない
	assert.False(t, bullet.Update())

	// その後29回目まで更新されない
	for i := 1; i < 29; i++ {
		assert.False(t, bullet.Update())
	}

	// 30回目は更新される
	assert.True(t, bullet.Update())
	assert.Equal(t, &Position{X: 4, Y: 8}, bullet.Position())
}
