package game

import "github.com/shibayu36/terminal-shooter/shared"

// プレイヤーの状態を管理する
type Player struct {
	PlayerID  PlayerID
	Position  Position
	Direction Direction
}

// プレイヤーの前方の座標を取得する
func (p *Player) FowardPosition() Position {
	dx, dy := p.Direction.ToVector()
	return Position{X: p.Position.X + dx, Y: p.Position.Y + dy}
}

// プレイヤーの状態をshared.PlayerStateに変換する
func (ps *Player) ToSharedPlayerState(status shared.Status) *shared.PlayerState {
	return &shared.PlayerState{
		PlayerId: string(ps.PlayerID),
		Position: &shared.Position{
			X: int32(ps.Position.X),
			Y: int32(ps.Position.Y),
		},
		Direction: ps.Direction.ToSharedDirection(),
		Status:    status,
	}
}
