package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GameState(t *testing.T) {
	gs := NewGameState()

	// player1を追加
	gs.AddPlayer("player1", &PlayerState{Position: &Position{X: 0, Y: 0}})
	assert.Equal(t, 0, gs.GetPlayers()["player1"].Position.X)
	assert.Equal(t, 0, gs.GetPlayers()["player1"].Position.Y)

	// player1の位置を更新
	gs.UpdatePlayerPosition("player1", &Position{X: 2, Y: 8})
	assert.Equal(t, 2, gs.GetPlayers()["player1"].Position.X)
	assert.Equal(t, 8, gs.GetPlayers()["player1"].Position.Y)

	// player2を追加
	gs.AddPlayer("player2", &PlayerState{Position: &Position{X: 10, Y: 10}})
	assert.Len(t, gs.GetPlayers(), 2)
	assert.Equal(t, 10, gs.GetPlayers()["player2"].Position.X)
	assert.Equal(t, 10, gs.GetPlayers()["player2"].Position.Y)

	// player1を削除
	gs.RemovePlayer("player1")
	assert.Len(t, gs.GetPlayers(), 1)
	assert.Equal(t, 10, gs.GetPlayers()["player2"].Position.X)
	assert.Equal(t, 10, gs.GetPlayers()["player2"].Position.Y)
}
