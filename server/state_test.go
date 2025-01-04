package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GameState(t *testing.T) {
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
