package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GameState(t *testing.T) {
	t.Run("プレイヤーを追加できる", func(t *testing.T) {
		gameState := NewGameState(30, 30)

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
		gameState := NewGameState(30, 30)

		itemID1 := gameState.AddBullet(&Position{X: 3, Y: 8}, DirectionRight)
		assert.Len(t, gameState.Items, 1)
		assert.Equal(t, ItemTypeBullet, gameState.Items[itemID1].Type())
		assert.Equal(t, Position{X: 3, Y: 8}, gameState.Items[itemID1].Position())

		itemID2 := gameState.AddBullet(&Position{X: 1, Y: 2}, DirectionRight)
		assert.Len(t, gameState.Items, 2)
		assert.Equal(t, ItemTypeBullet, gameState.Items[itemID2].Type())
		assert.Equal(t, Position{X: 1, Y: 2}, gameState.Items[itemID2].Position())
	})
}

func Test_GameState_StartUpdateLoop(t *testing.T) {
	t.Run("updateが定期的に実行される", func(t *testing.T) {
		gameState := NewGameState(30, 30)

		bulletID := gameState.AddBullet(&Position{X: 0, Y: 0}, DirectionRight)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		gameState.StartUpdateLoop(ctx)

		time.Sleep(560 * time.Millisecond) // 約33回のtickが発生する時間

		pos := gameState.Items[bulletID].Position()
		assert.Positive(t, pos.X, "弾が移動していること")
	})

	t.Run("contextのキャンセルでループが終了する", func(t *testing.T) {
		gameState := NewGameState(30, 30)
		ctx, cancel := context.WithCancel(context.Background())

		bulletID := gameState.AddBullet(&Position{X: 0, Y: 0}, DirectionRight)

		gameState.StartUpdateLoop(ctx)

		// キャンセル実行
		cancel()

		time.Sleep(560 * time.Millisecond) // 約33回のtickが発生する時間

		assert.Equal(t, 0, gameState.Items[bulletID].Position().X)
	})
}

func Test_GameState_update(t *testing.T) {
	t.Run("弾が動く", func(t *testing.T) {
		itemsUpdatedCh := make(chan struct{}, 10)
		gameState := NewGameState(30, 30)

		// 弾を追加
		bulletID1 := gameState.AddBullet(&Position{X: 3, Y: 8}, DirectionLeft)
		// 2回動かす
		gameState.update(itemsUpdatedCh)
		gameState.update(itemsUpdatedCh)

		// 弾をもう一つ追加
		bulletID2 := gameState.AddBullet(&Position{X: 1, Y: 2}, DirectionUp)

		// 28回動かすと、bullet1だけ動く
		for range 28 {
			gameState.update(itemsUpdatedCh)
		}
		assert.Equal(t, Position{X: 2, Y: 8}, gameState.Items[bulletID1].Position())
		assert.Equal(t, Position{X: 1, Y: 2}, gameState.Items[bulletID2].Position())
		// bullet1が更新されたので更新件数が1件になる
		assert.Len(t, itemsUpdatedCh, 1)

		// さらに2回動かすと、bullet2が動く
		gameState.update(itemsUpdatedCh)
		gameState.update(itemsUpdatedCh)
		assert.Equal(t, Position{X: 2, Y: 8}, gameState.Items[bulletID1].Position())
		assert.Equal(t, Position{X: 1, Y: 1}, gameState.Items[bulletID2].Position())
		// bullet2が更新されたので更新件数が2件になる
		assert.Len(t, itemsUpdatedCh, 2)
	})

	t.Run("アイテムが盤面外に出たら削除される", func(t *testing.T) {
		itemsUpdatedCh := make(chan struct{}, 10)
		gameState := NewGameState(30, 30)

		bulletID := gameState.AddBullet(&Position{X: 1, Y: 0}, DirectionLeft)

		// 30回更新したタイミングではまだ盤面上
		for range 30 {
			gameState.update(itemsUpdatedCh)
		}
		assert.Len(t, gameState.GetItems(), 1)
		assert.Equal(t, Position{X: 0, Y: 0}, gameState.GetItems()[bulletID].Position())

		// さらに30回更新したら盤面外に出るので削除される
		for range 30 {
			gameState.update(itemsUpdatedCh)
		}
		assert.Empty(t, gameState.GetItems())
		assert.Len(t, gameState.GetRemovedItems(), 1)
		assert.NotEmpty(t, gameState.GetRemovedItems()[bulletID])
	})
}

func Test_GameState_isWithinBounds(t *testing.T) {
	testCases := []struct {
		name     string
		pos      *Position
		expected bool
	}{
		{name: "盤面内にある", pos: &Position{X: 15, Y: 15}, expected: true},
		{name: "盤面内にある", pos: &Position{X: 0, Y: 0}, expected: true},
		{name: "盤面内にある", pos: &Position{X: 29, Y: 29}, expected: true},
		{name: "X < 0で盤面外にある", pos: &Position{X: -1, Y: 15}, expected: false},
		{name: "X >= 30で盤面外にある", pos: &Position{X: 30, Y: 15}, expected: false},
		{name: "Y < 0で盤面外にある", pos: &Position{X: 15, Y: -1}, expected: false},
		{name: "Y >= 30で盤面外にある", pos: &Position{X: 15, Y: 30}, expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gameState := NewGameState(30, 30)
			bullet := NewBullet("bullet1", tc.pos, DirectionUp)
			assert.Equal(t, tc.expected, gameState.isWithinBounds(bullet))
		})
	}
}

func Test_GameState_ItemsOperation(t *testing.T) {
	gameState := NewGameState(30, 30)

	bulletID1 := gameState.AddBullet(&Position{X: 3, Y: 8}, DirectionLeft)
	bullet1 := NewBullet(bulletID1, &Position{X: 3, Y: 8}, DirectionLeft)
	bulletID2 := gameState.AddBullet(&Position{X: 1, Y: 2}, DirectionUp)
	bullet2 := NewBullet(bulletID2, &Position{X: 1, Y: 2}, DirectionUp)
	bulletID3 := gameState.AddBullet(&Position{X: 2, Y: 3}, DirectionRight)
	bullet3 := NewBullet(bulletID3, &Position{X: 2, Y: 3}, DirectionRight)

	items := gameState.GetItems()
	assert.Len(t, items, 3)
	assert.Equal(t, map[ItemID]Item{
		bulletID1: bullet1,
		bulletID2: bullet2,
		bulletID3: bullet3,
	}, items)

	// bulletID1と3を削除
	gameState.removeItem(bulletID1)
	gameState.removeItem(bulletID3)

	// Itemsにはbullet2のみ残っている
	items = gameState.GetItems()
	assert.Len(t, items, 1)
	assert.Equal(t, map[ItemID]Item{
		bulletID2: bullet2,
	}, items)

	// RemovedItemsにはbullet1とbullet3が残っている
	removedItems := gameState.GetRemovedItems()
	assert.Len(t, removedItems, 2)
	assert.Equal(t, map[ItemID]Item{
		bulletID1: bullet1,
		bulletID3: bullet3,
	}, removedItems)

	// ClearRemovedItemsでbullet1のみ削除する
	gameState.ClearRemovedItem(bulletID1)

	// RemovedItemsにはbullet1のみ残っている
	removedItems = gameState.GetRemovedItems()
	assert.Len(t, removedItems, 1)
	assert.Equal(t, map[ItemID]Item{
		bulletID3: bullet3,
	}, removedItems)
}

func Test_Bullet(t *testing.T) {
	bullet := NewBullet("bullet1", &Position{X: 3, Y: 8}, DirectionRight)

	assert.Equal(t, ItemID("bullet1"), bullet.ID())
	assert.Equal(t, ItemTypeBullet, bullet.Type())
	assert.Equal(t, Position{X: 3, Y: 8}, bullet.Position())

	// 1回目は更新されない
	assert.False(t, bullet.Update())

	// その後29回目まで更新されない
	for i := 1; i < 29; i++ {
		assert.False(t, bullet.Update())
	}

	// 30回目は更新される
	assert.True(t, bullet.Update())
	assert.Equal(t, Position{X: 4, Y: 8}, bullet.Position())
}
