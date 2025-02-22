package game

// GameCollisionService は衝突時に必要な操作を提供するインターフェース。Gameのメソッドの一部だけを公開する
type GameCollisionService interface {
	RemoveItem(id ItemID)
	UpdatePlayerStatus(playerID PlayerID, status PlayerStatus) *Player
}

// Collidable は衝突判定可能なオブジェクトを表すインターフェース
type Collidable interface {
	Position() Position
	// 衝突時の処理を行う。自身の状態が変更された場合はtrueを返す
	OnCollideWith(other Collidable, svc GameCollisionService) bool
}

// Collision はPlayerとItem間の衝突を表す
type Collision struct {
	Player *Player
	Item   Item
}
