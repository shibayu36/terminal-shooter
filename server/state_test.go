package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GameState(t *testing.T) {
	gs := NewGameState()

	// player1を追加
	gs.AddPlayer("player1", &PlayerState{Position: &Position{X: 0, Y: 0}})
	assert.Equal(t, gs.GetPlayers()["player1"].Position.X, 0)
	assert.Equal(t, gs.GetPlayers()["player1"].Position.Y, 0)

	// player1の位置を更新
	gs.UpdatePlayerPosition("player1", &Position{X: 2, Y: 8})
	assert.Equal(t, gs.GetPlayers()["player1"].Position.X, 2)
	assert.Equal(t, gs.GetPlayers()["player1"].Position.Y, 8)

	// player2を追加
	gs.AddPlayer("player2", &PlayerState{Position: &Position{X: 10, Y: 10}})
	assert.Len(t, gs.GetPlayers(), 2)
	assert.Equal(t, gs.GetPlayers()["player2"].Position.X, 10)
	assert.Equal(t, gs.GetPlayers()["player2"].Position.Y, 10)

	// player1を削除
	gs.RemovePlayer("player1")
	assert.Equal(t, len(gs.GetPlayers()), 1)
	assert.Equal(t, gs.GetPlayers()["player2"].Position.X, 10)
	assert.Equal(t, gs.GetPlayers()["player2"].Position.Y, 10)
}
