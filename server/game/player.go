package game

import "github.com/shibayu36/terminal-shooter/shared"

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
}

// プレイヤーの前方の座標を取得する
func (p *Player) FowardPosition() Position {
	dx, dy := p.Direction.ToVector()
	return Position{X: p.Position.X + dx, Y: p.Position.Y + dy}
}

// プレイヤーの状態をshared.PlayerStateに変換する
func (p *Player) ToSharedPlayerState(status shared.Status) *shared.PlayerState {
	return &shared.PlayerState{
		PlayerId: string(p.PlayerID),
		Position: &shared.Position{
			X: int32(p.Position.X),
			Y: int32(p.Position.Y),
		},
		Direction: p.Direction.ToSharedDirection(),
		Status:    status,
	}
}
