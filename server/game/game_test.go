package game

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Game(t *testing.T) {
	t.Run("プレイヤーを追加できる", func(t *testing.T) {
		game := NewGame(30, 30)

		// player1を追加
		game.AddPlayer("player1")
		assert.Equal(t, 0, game.GetPlayers()["player1"].Position().X)
		assert.Equal(t, 0, game.GetPlayers()["player1"].Position().Y)
		assert.Equal(t, DirectionUp, game.GetPlayers()["player1"].Direction())

		// player1の位置を更新
		game.MovePlayer("player1", Position{X: 2, Y: 8}, DirectionRight)
		assert.Equal(t, 2, game.GetPlayers()["player1"].Position().X)
		assert.Equal(t, 8, game.GetPlayers()["player1"].Position().Y)
		assert.Equal(t, DirectionRight, game.GetPlayers()["player1"].Direction())

		// player2を追加
		game.AddPlayer("player2")
		assert.Len(t, game.GetPlayers(), 2)
		assert.Equal(t, 0, game.GetPlayers()["player2"].Position().X)
		assert.Equal(t, 0, game.GetPlayers()["player2"].Position().Y)
		assert.Equal(t, DirectionUp, game.GetPlayers()["player2"].Direction())

		// player1を削除
		game.RemovePlayer("player1")
		assert.Len(t, game.GetPlayers(), 1)
		assert.Equal(t, 0, game.GetPlayers()["player2"].Position().X)
	})

	t.Run("弾を追加できる", func(t *testing.T) {
		game := NewGame(30, 30)

		itemID1 := game.AddBullet(Position{X: 3, Y: 8}, DirectionRight)
		assert.Len(t, game.Items, 1)
		assert.Equal(t, ItemTypeBullet, game.Items[itemID1].Type())
		assert.Equal(t, Position{X: 3, Y: 8}, game.Items[itemID1].Position())

		itemID2 := game.AddBullet(Position{X: 1, Y: 2}, DirectionRight)
		assert.Len(t, game.Items, 2)
		assert.Equal(t, ItemTypeBullet, game.Items[itemID2].Type())
		assert.Equal(t, Position{X: 1, Y: 2}, game.Items[itemID2].Position())
	})
}

func Test_Game_StartUpdateLoop(t *testing.T) {
	t.Run("updateが定期的に実行される", func(t *testing.T) {
		game := NewGame(30, 30)

		playerID := PlayerID("player1")
		game.AddPlayer(playerID)
		game.MovePlayer(playerID, Position{X: 0, Y: 0}, DirectionRight)
		bulletID := game.ShootBullet(playerID)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		updatedCh := game.StartUpdateLoop(ctx)

		// 通知が詰まらないようにチャネルから取り出しておく
		go func() {
			for range updatedCh {
			}
		}()

		time.Sleep(560 * time.Millisecond) // 約33回のtickが発生する時間

		pos := game.Items[bulletID].Position()
		assert.Equal(t, 2, pos.X, "弾が移動していること")
	})

	t.Run("contextのキャンセルでループが終了する", func(t *testing.T) {
		game := NewGame(30, 30)
		ctx, cancel := context.WithCancel(context.Background())

		bulletID := game.AddBullet(Position{X: 0, Y: 0}, DirectionRight)

		updatedCh := game.StartUpdateLoop(ctx)

		// 通知が詰まらないようにチャネルから取り出しておく
		go func() {
			for range updatedCh {
			}
		}()

		// キャンセル実行
		cancel()

		time.Sleep(560 * time.Millisecond) // 約33回のtickが発生する時間

		assert.Equal(t, 0, game.Items[bulletID].Position().X)
	})
}

func Test_Game_update(t *testing.T) {
	t.Run("弾が動く", func(t *testing.T) {
		updatedCh := make(chan UpdatedResult, 10)
		game := NewGame(30, 30)

		playerID := PlayerID("player1")
		game.AddPlayer(playerID)
		game.MovePlayer(playerID, Position{X: 4, Y: 8}, DirectionLeft)

		// 弾を追加
		bulletID1 := game.ShootBullet(playerID)
		// 2回動かす
		game.update(updatedCh)
		game.update(updatedCh)
		assert.Len(t, updatedCh, 1, "弾の追加が通知されている")

		// 弾をもう一つ追加
		game.MovePlayer(playerID, Position{X: 1, Y: 3}, DirectionUp)
		bulletID2 := game.ShootBullet(playerID)
		game.update(updatedCh)
		assert.Len(t, updatedCh, 2, "弾の追加が通知されている")

		// あと27回動かすと、bullet1だけ動く
		for range 27 {
			game.update(updatedCh)
		}
		assert.Equal(t, Position{X: 2, Y: 8}, game.Items[bulletID1].Position())
		assert.Equal(t, Position{X: 1, Y: 2}, game.Items[bulletID2].Position())
		assert.Len(t, updatedCh, 3, "弾の追加x2 + 1回の更新で3件")

		// さらに2回動かすと、bullet2が動く
		game.update(updatedCh)
		game.update(updatedCh)
		assert.Equal(t, Position{X: 2, Y: 8}, game.Items[bulletID1].Position())
		assert.Equal(t, Position{X: 1, Y: 1}, game.Items[bulletID2].Position())
		assert.Len(t, updatedCh, 4, "弾の追加x2 + 2回の更新で4件")
	})

	t.Run("ボムを配置して180回更新すると爆発する", func(t *testing.T) {
		updatedCh := make(chan UpdatedResult, 10)
		game := NewGame(30, 30)

		playerID := PlayerID("player1")
		game.AddPlayer(playerID)
		game.MovePlayer(playerID, Position{X: 5, Y: 8}, DirectionRight)

		bombID := game.PlaceBomb(playerID)
		assert.NotEmpty(t, bombID)

		for range 179 {
			game.update(updatedCh)
		}
		// まだ爆発前
		bomb, ok := game.Items[bombID].(*Bomb)
		assert.True(t, ok)
		assert.Equal(t, Position{X: 5, Y: 8}, bomb.Position())

		// 180回目に爆発する
		game.update(updatedCh)
		game.update(updatedCh)
		assert.Len(t, game.GetItems(), 17, "爆発したので17個のBombFireが残っている")
		assert.Len(t, game.GetRemovedItems(), 1)
		assert.NotEmpty(t, game.GetRemovedItems()[bombID])

		// 火の位置にプレイヤーがいるのでプレイヤーは死ぬ
		assert.Equal(t, PlayerStatusDead, game.GetPlayers()[playerID].Status())
	})

	t.Run("アイテムが盤面外に出たら削除される", func(t *testing.T) {
		updatedCh := make(chan UpdatedResult, 10)
		game := NewGame(30, 30)

		bulletID := game.AddBullet(Position{X: 1, Y: 0}, DirectionLeft)

		// 30回更新したタイミングではまだ盤面上
		for range 30 {
			game.update(updatedCh)
		}
		assert.Len(t, game.GetItems(), 1)
		assert.Equal(t, Position{X: 0, Y: 0}, game.GetItems()[bulletID].Position())

		// さらに30回更新したら盤面外に出るので削除される
		for range 30 {
			game.update(updatedCh)
		}
		assert.Empty(t, game.GetItems())
		assert.Len(t, game.GetRemovedItems(), 1)
		assert.NotEmpty(t, game.GetRemovedItems()[bulletID])
	})

	t.Run("プレイヤーと弾が衝突するとプレイヤーがdeadになり、弾は消え、更新が通知される", func(t *testing.T) {
		updatedCh := make(chan UpdatedResult, 10)

		game := NewGame(30, 30)

		playerID := PlayerID("player1")
		game.AddPlayer(playerID)
		game.MovePlayer(playerID, Position{X: 2, Y: 3}, DirectionRight)
		bulletID := game.ShootBullet(playerID)
		// 当たるように移動しておく
		game.MovePlayer(playerID, Position{X: 4, Y: 3}, DirectionRight)

		game.update(updatedCh)

		// まだ衝突していない
		assert.Equal(t, PlayerStatusAlive, game.GetPlayers()[playerID].Status())
		assert.Len(t, game.GetItems(), 1)
		assert.Empty(t, game.GetRemovedItems())
		assert.Len(t, updatedCh, 1, "弾の追加のみが通知される")

		// 29回動くと弾が当たる
		for range 29 {
			game.update(updatedCh)
		}
		assert.Equal(t, PlayerStatusDead, game.GetPlayers()[playerID].Status())
		assert.Empty(t, game.GetItems())
		assert.Len(t, game.GetRemovedItems(), 1)
		assert.NotEmpty(t, game.GetRemovedItems()[bulletID])
		assert.Len(t, updatedCh, 3, "弾の追加、弾の更新とプレイヤーの更新の3件が通知される")
	})
}

func Test_Game_update_checkCollisions(t *testing.T) {
	t.Run("弾がプレイヤーに当たったらプレイヤーがdeadになり、弾は消える", func(t *testing.T) {
		updatedCh := make(chan UpdatedResult, 10)

		game := NewGame(30, 30)

		playerID := PlayerID("player1")
		game.AddPlayer(playerID)
		game.MovePlayer(playerID, Position{X: 2, Y: 3}, DirectionRight)
		bulletID := game.AddBullet(Position{X: 1, Y: 3}, DirectionRight)

		game.update(updatedCh)

		// まだ衝突していない
		assert.Equal(t, PlayerStatusAlive, game.GetPlayers()[playerID].Status())
		assert.Len(t, game.GetItems(), 1)
		assert.Empty(t, game.GetRemovedItems())

		// 29回動くと弾が当たる
		for range 29 {
			game.update(updatedCh)
		}
		assert.Equal(t, PlayerStatusDead, game.GetPlayers()[playerID].Status())
		assert.Empty(t, game.GetItems())
		assert.Len(t, game.GetRemovedItems(), 1)
		assert.NotEmpty(t, game.GetRemovedItems()[bulletID])
	})
}

func Test_Game_detectCollisions(t *testing.T) {
	t.Run("プレイヤーと弾の衝突を検出できる", func(t *testing.T) {
		game := NewGame(30, 30)

		collidedPlayerID1 := PlayerID("player1")
		game.AddPlayer(collidedPlayerID1)
		game.MovePlayer(collidedPlayerID1, Position{X: 2, Y: 3}, DirectionRight)

		collidedPlayerID2 := PlayerID("player2")
		game.AddPlayer(collidedPlayerID2)
		game.MovePlayer(collidedPlayerID2, Position{X: 1, Y: 4}, DirectionRight)

		otherPlayerID := PlayerID("player3")
		game.AddPlayer(otherPlayerID)
		game.MovePlayer(otherPlayerID, Position{X: 1, Y: 3}, DirectionRight)

		collidedBulletID1 := game.AddBullet(Position{X: 2, Y: 3}, DirectionRight)
		collidedBulletID2 := game.AddBullet(Position{X: 1, Y: 4}, DirectionRight)

		game.AddBullet(Position{X: 5, Y: 5}, DirectionRight)

		collisions := game.detectCollisions()
		assert.Len(t, collisions, 2)

		expected := map[PlayerID]ItemID{
			collidedPlayerID1: collidedBulletID1,
			collidedPlayerID2: collidedBulletID2,
		}

		for _, collision := range collisions {
			assert.Equal(t, expected[collision.Player.PlayerID], collision.Item.ID())
		}
	})

	t.Run("どれも衝突していない", func(t *testing.T) {
		game := NewGame(30, 30)

		playerID := PlayerID("player1")
		game.AddPlayer(playerID)
		game.MovePlayer(playerID, Position{X: 2, Y: 3}, DirectionRight)

		game.AddBullet(Position{X: 1, Y: 4}, DirectionRight)

		collisions := game.detectCollisions()
		assert.Empty(t, collisions)
	})
}

func Test_Game_isWithinBounds(t *testing.T) {
	testCases := []struct {
		name     string
		pos      Position
		expected bool
	}{
		{name: "盤面内にある", pos: Position{X: 15, Y: 15}, expected: true},
		{name: "盤面内にある", pos: Position{X: 0, Y: 0}, expected: true},
		{name: "盤面内にある", pos: Position{X: 29, Y: 29}, expected: true},
		{name: "X < 0で盤面外にある", pos: Position{X: -1, Y: 15}, expected: false},
		{name: "X >= 30で盤面外にある", pos: Position{X: 30, Y: 15}, expected: false},
		{name: "Y < 0で盤面外にある", pos: Position{X: 15, Y: -1}, expected: false},
		{name: "Y >= 30で盤面外にある", pos: Position{X: 15, Y: 30}, expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			game := NewGame(30, 30)
			bullet := NewBullet("bullet1", tc.pos, DirectionUp)
			assert.Equal(t, tc.expected, game.isWithinBounds(bullet))
		})
	}
}

func Test_Game_ItemsOperation(t *testing.T) {
	game := NewGame(30, 30)

	bulletID1 := game.AddBullet(Position{X: 3, Y: 8}, DirectionLeft)
	bullet1 := NewBullet(bulletID1, Position{X: 3, Y: 8}, DirectionLeft)
	bulletID2 := game.AddBullet(Position{X: 1, Y: 2}, DirectionUp)
	bullet2 := NewBullet(bulletID2, Position{X: 1, Y: 2}, DirectionUp)
	bulletID3 := game.AddBullet(Position{X: 2, Y: 3}, DirectionRight)
	bullet3 := NewBullet(bulletID3, Position{X: 2, Y: 3}, DirectionRight)

	items := game.GetItems()
	assert.Len(t, items, 3)
	assert.Equal(t, map[ItemID]Item{
		bulletID1: bullet1,
		bulletID2: bullet2,
		bulletID3: bullet3,
	}, items)

	// bulletID1と3を削除
	game.RemoveItem(bulletID1)
	game.RemoveItem(bulletID3)

	// Itemsにはbullet2のみ残っている
	items = game.GetItems()
	assert.Len(t, items, 1)
	assert.Equal(t, map[ItemID]Item{
		bulletID2: bullet2,
	}, items)

	// RemovedItemsにはbullet1とbullet3が残っている
	removedItems := game.GetRemovedItems()
	assert.Len(t, removedItems, 2)
	assert.Equal(t, map[ItemID]Item{
		bulletID1: bullet1,
		bulletID3: bullet3,
	}, removedItems)

	// ClearRemovedItemでbullet1のみ削除する
	game.ClearRemovedItem(bulletID1)

	// RemovedItemsにはbullet1のみ残っている
	removedItems = game.GetRemovedItems()
	assert.Len(t, removedItems, 1)
	assert.Equal(t, map[ItemID]Item{
		bulletID3: bullet3,
	}, removedItems)
}

func Test_Game_MovePlayer(t *testing.T) {
	t.Run("プレイヤーを指定した位置、方向に移動できる", func(t *testing.T) {
		game := NewGame(30, 30)
		playerID := PlayerID("player1")
		game.AddPlayer(playerID)

		game.MovePlayer(playerID, Position{X: 2, Y: 3}, DirectionRight)
		assert.Equal(t, Position{X: 2, Y: 3}, game.GetPlayers()[playerID].Position())
		assert.Equal(t, DirectionRight, game.GetPlayers()[playerID].Direction())
	})

	t.Run("プレイヤーが死んでいる場合は位置を更新できない", func(t *testing.T) {
		game := NewGame(30, 30)
		playerID := PlayerID("player1")
		game.AddPlayer(playerID)

		game.MovePlayer(playerID, Position{X: 2, Y: 3}, DirectionRight)
		game.UpdatePlayerStatus(playerID, PlayerStatusDead)
		game.MovePlayer(playerID, Position{X: 2, Y: 3}, DirectionRight)
		assert.Equal(t, Position{X: 2, Y: 3}, game.GetPlayers()[playerID].Position())
	})
}

func Test_Game_UpdatePlayerStatus(t *testing.T) {
	game := NewGame(30, 30)
	playerID := PlayerID("player1")
	game.AddPlayer(playerID)

	game.UpdatePlayerStatus(playerID, PlayerStatusDead)
	assert.Equal(t, PlayerStatusDead, game.GetPlayers()[playerID].Status())

	// 一度deadになったらaliveに戻せない
	game.UpdatePlayerStatus(playerID, PlayerStatusAlive)
	assert.Equal(t, PlayerStatusDead, game.GetPlayers()[playerID].Status())
}

func Test_Game_ShootBullet(t *testing.T) {
	t.Run("プレイヤーが弾を発射できる", func(t *testing.T) {
		game := NewGame(30, 30)

		playerID1 := PlayerID("player1")
		playerID2 := PlayerID("player2")

		game.AddPlayer(playerID1)
		game.AddPlayer(playerID2)

		game.MovePlayer(playerID1, Position{X: 3, Y: 8}, DirectionRight)

		bulletID := game.ShootBullet(playerID1)
		assert.NotEmpty(t, bulletID)

		// player1の前方に弾が発射されていること
		bullet, ok := (game.GetItems()[bulletID]).(*Bullet)
		assert.True(t, ok)
		assert.Equal(t, Position{X: 4, Y: 8}, bullet.Position())
		assert.Equal(t, DirectionRight, bullet.direction)
	})

	t.Run("存在しないプレイヤーから弾を発射すると何もしない", func(t *testing.T) {
		game := NewGame(30, 30)
		bulletID := game.ShootBullet(PlayerID("player1"))
		assert.Empty(t, bulletID)
		assert.Empty(t, game.GetItems())
	})

	t.Run("deadのプレイヤーから弾を発射すると何もしない", func(t *testing.T) {
		game := NewGame(30, 30)
		playerID := PlayerID("player1")
		game.AddPlayer(playerID)
		game.UpdatePlayerStatus(playerID, PlayerStatusDead)
		bulletID := game.ShootBullet(playerID)
		assert.Empty(t, bulletID)
	})
}

func Test_Game_PlaceBomb(t *testing.T) {
	t.Run("プレイヤーがボムを設置できる", func(t *testing.T) {
		game := NewGame(30, 30)
		playerID := PlayerID("player1")
		game.AddPlayer(playerID)
		game.MovePlayer(playerID, Position{X: 3, Y: 8}, DirectionRight)

		bombID := game.PlaceBomb(playerID)
		assert.NotEmpty(t, bombID)

		// player1と同じ位置にボムが設置されていること
		bomb, ok := (game.GetItems()[bombID]).(*Bomb)
		assert.True(t, ok)
		assert.Equal(t, Position{X: 3, Y: 8}, bomb.Position())
	})

	t.Run("存在しないプレイヤーからボムを設置すると何もしない", func(t *testing.T) {
		game := NewGame(30, 30)
		bombID := game.PlaceBomb(PlayerID("player1"))
		assert.Empty(t, bombID)
	})

	t.Run("deadのプレイヤーからボムを設置すると何もしない", func(t *testing.T) {
		game := NewGame(30, 30)
		playerID := PlayerID("player1")
		game.AddPlayer(playerID)
		game.UpdatePlayerStatus(playerID, PlayerStatusDead)
		bombID := game.PlaceBomb(playerID)
		assert.Empty(t, bombID)
	})
}
