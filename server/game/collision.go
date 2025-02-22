package game

// GameCollisionService は衝突時に必要な操作を提供するインターフェース。Gameのメソッドの一部だけを公開する
type GameCollisionService interface {
	RemoveItem(id ItemID)
	UpdatePlayerStatus(playerID PlayerID, status PlayerStatus) *Player
}

// Collidable は衝突判定可能なオブジェクトを表すインターフェース
type Collidable interface {
	Position() Position
	OnCollideWith(other Collidable, svc GameCollisionService)
}

// Collision は2つのオブジェクト間の衝突を表す
type Collision struct {
	Obj1 Collidable
	Obj2 Collidable
}
