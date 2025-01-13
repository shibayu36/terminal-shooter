package game

import (
	"fmt"
	"sync"

	"github.com/shibayu36/terminal-shooter/shared"
)

type PlayerStatus string

const (
	PlayerStatusAlive PlayerStatus = "alive"
	PlayerStatusDead  PlayerStatus = "dead"
)

// プレイヤーの状態を管理する
type Player struct {
	PlayerID  PlayerID
	Position  Position
	Direction Direction
	Status    PlayerStatus

	mu sync.RWMutex `exhaustruct:"optional"`
}

func (p *Player) Move(position Position, direction Direction) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Status == PlayerStatusDead {
		// deadの場合は移動できない
		return
	}
	p.Position = position
	p.Direction = direction
}

func (p *Player) UpdateStatus(status PlayerStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Status == PlayerStatusDead {
		// deadの場合は更新できない
		return
	}
	p.Status = status
}

// プレイヤーの前方の座標を取得する
func (p *Player) FowardPosition() Position {
	p.mu.RLock()
	defer p.mu.RUnlock()
	dx, dy := p.Direction.ToVector()
	return Position{X: p.Position.X + dx, Y: p.Position.Y + dy}
}

// プレイヤーの状態をshared.PlayerStateに変換する
func (p *Player) ToSharedPlayerState() *shared.PlayerState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return &shared.PlayerState{
		PlayerId: string(p.PlayerID),
		Position: &shared.Position{
			X: int32(p.Position.X),
			Y: int32(p.Position.Y),
		},
		Direction: p.Direction.ToSharedDirection(),
		Status:    p.Status.ToSharedStatus(),
	}
}

func (ps PlayerStatus) ToSharedStatus() shared.Status {
	switch ps {
	case PlayerStatusAlive:
		return shared.Status_ALIVE
	case PlayerStatusDead:
		return shared.Status_DEAD
	default:
		panic(fmt.Sprintf("invalid player status: %s", ps))
	}
}
