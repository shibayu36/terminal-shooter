package game

// gameCollisionService は衝突時に必要な操作を提供するインターフェース。Gameのメソッドの一部だけを公開する
type gameCollisionService interface {
	RemoveItem(id ItemID)
	UpdatePlayerStatus(playerID PlayerID, status PlayerStatus) *Player
}

// collidable は衝突判定可能なオブジェクトを表すインターフェース
type collidable interface {
	Position() Position
	// 衝突時の処理を行う。自身の状態が変更された場合はtrueを返す
	OnCollideWith(other collidable, svc gameCollisionService) bool
}

// collision はPlayerとItem間の衝突を表す
type collision struct {
	Player *Player
	Item   Item
}
