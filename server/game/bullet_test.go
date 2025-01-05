package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Bullet(t *testing.T) {
	bullet := NewBullet("bullet1", Position{X: 3, Y: 8}, DirectionRight)

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