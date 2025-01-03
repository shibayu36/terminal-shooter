package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GameState(t *testing.T) {
	gameState := NewGameState()

	// player1を追加
	gameState.AddPlayer("player1", &PlayerState{Position: &Position{X: 0, Y: 0}, Direction: DirectionUp})
	assert.Equal(t, 0, gameState.GetPlayers()["player1"].Position.X)
	assert.Equal(t, 0, gameState.GetPlayers()["player1"].Position.Y)
	assert.Equal(t, DirectionUp, gameState.GetPlayers()["player1"].Direction)

	// player1の位置を更新
	gameState.MovePlayer("player1", &Position{X: 2, Y: 8}, DirectionRight)
	assert.Equal(t, 2, gameState.GetPlayers()["player1"].Position.X)
	assert.Equal(t, 8, gameState.GetPlayers()["player1"].Position.Y)
	assert.Equal(t, DirectionRight, gameState.GetPlayers()["player1"].Direction)

	// player2を追加
	gameState.AddPlayer("player2", &PlayerState{Position: &Position{X: 10, Y: 10}, Direction: DirectionDown})
	assert.Len(t, gameState.GetPlayers(), 2)
	assert.Equal(t, 10, gameState.GetPlayers()["player2"].Position.X)
	assert.Equal(t, 10, gameState.GetPlayers()["player2"].Position.Y)
	assert.Equal(t, DirectionDown, gameState.GetPlayers()["player2"].Direction)

	// player1を削除
	gameState.RemovePlayer("player1")
	assert.Len(t, gameState.GetPlayers(), 1)
	assert.Equal(t, 10, gameState.GetPlayers()["player2"].Position.X)
	assert.Equal(t, 10, gameState.GetPlayers()["player2"].Position.Y)
	assert.Equal(t, DirectionDown, gameState.GetPlayers()["player2"].Direction)
}
