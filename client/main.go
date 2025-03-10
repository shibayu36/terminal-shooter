package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/cockroachdb/errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/shibayu36/terminal-shooter/shared"
	"google.golang.org/protobuf/proto"
)

type Position struct {
	X int
	Y int
}

type Player struct {
	ID        string
	Position  Position
	Direction shared.Direction
	Status    shared.Status
}

type Item struct {
	ID       string
	Type     shared.ItemType
	Position Position
}

type Game struct {
	mqtt mqtt.Client

	screen tcell.Screen

	myPlayerID string
	players    map[string]Player
	items      map[string]Item
	width      int
	height     int

	messageStats *MessageStats
}

func (g *Game) publishMyState() {
	myPlayer := g.getMyPlayer()

	state := &shared.PlayerState{
		PlayerId: g.myPlayerID,
		Position: &shared.Position{
			X: int32(myPlayer.Position.X),
			Y: int32(myPlayer.Position.Y),
		},
		Direction: myPlayer.Direction,
	}

	data, err := proto.Marshal(state)
	if err != nil {
		log.Printf("Failed to encode position: %v", err)
		return
	}

	token := g.mqtt.Publish("player_state", 0, false, data)
	if token.Wait() && token.Error() != nil {
		log.Printf("Failed to publish position: %v", token.Error())
		return
	}
}

func (g *Game) movePlayer(direction shared.Direction) {
	myPlayer := g.getMyPlayer()
	oldX, oldY := myPlayer.Position.X, myPlayer.Position.Y
	oldDirection := myPlayer.Direction

	// directionから移動量を決定
	var dx, dy int
	switch direction {
	case shared.Direction_LEFT:
		dx = -1
	case shared.Direction_RIGHT:
		dx = 1
	case shared.Direction_UP:
		dy = -1
	case shared.Direction_DOWN:
		dy = 1
	}

	if newX := myPlayer.Position.X + dx; newX >= 0 && newX < g.width {
		myPlayer.Position.X = newX
	}
	if newY := myPlayer.Position.Y + dy; newY >= 0 && newY < g.height {
		myPlayer.Position.Y = newY
	}
	myPlayer.Direction = direction

	g.players[g.myPlayerID] = myPlayer

	// 位置か方向が変更されたら自分の状態をサーバーに送る
	if oldX != myPlayer.Position.X || oldY != myPlayer.Position.Y || oldDirection != direction {
		g.publishMyState()
	}
}

func (g *Game) handleEvent(event tcell.Event) bool {
	//nolint:gocritic,varnamelen // ignore singleCaseSwitch
	switch ev := event.(type) {
	case *tcell.EventKey:
		//nolint:exhaustive
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			return true
		case tcell.KeyLeft:
			g.movePlayer(shared.Direction_LEFT)
		case tcell.KeyRight:
			g.movePlayer(shared.Direction_RIGHT)
		case tcell.KeyUp:
			g.movePlayer(shared.Direction_UP)
		case tcell.KeyDown:
			g.movePlayer(shared.Direction_DOWN)
		case tcell.KeyRune:
			if ev.Rune() == ' ' {
				g.shootBullet()
			} else if ev.Rune() == 'b' {
				g.placeBomb()
			}
		}
	}
	return false
}

func (g *Game) shootBullet() {
	req := &shared.PlayerActionRequest{
		Type: shared.ActionType_SHOOT_BULLET,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		log.Printf("Failed to encode player action request: %v", err)
		return
	}

	token := g.mqtt.Publish("player_action", 0, false, data)
	if token.Wait() && token.Error() != nil {
		log.Printf("Failed to publish player action request: %v", token.Error())
		return
	}
}

func (g *Game) placeBomb() {
	actionReq := &shared.PlayerActionRequest{
		Type: shared.ActionType_PLACE_BOMB,
	}
	payload, err := proto.Marshal(actionReq)
	if err != nil {
		log.Printf("Failed to marshal player action request: %v", err)
		return
	}

	if token := g.mqtt.Publish("player_action", 0, false, payload); token.Wait() && token.Error() != nil {
		log.Printf("Failed to publish player action: %v", token.Error())
	}
}

func getPlayerRune(player Player) rune {
	if player.Status == shared.Status_DEAD {
		return 'x'
	}

	switch player.Direction {
	case shared.Direction_UP:
		return '^'
	case shared.Direction_DOWN:
		return 'v'
	case shared.Direction_LEFT:
		return '<'
	case shared.Direction_RIGHT:
		return '>'
	default:
		return '^'
	}
}

const (
	bgColor          = tcell.Color232
	mapColor         = tcell.Color255
	myPlayerColor    = tcell.Color46
	otherPlayerColor = tcell.Color196
	itemColor        = tcell.Color226
	bombColor        = tcell.Color208
	fireColor        = tcell.Color196
)

//nolint:funlen
func (g *Game) draw() {
	g.screen.Clear()

	defaultStyle := tcell.StyleDefault.
		Background(bgColor).
		Foreground(mapColor)

	// マップを描画
	for y := range g.height {
		for x := range g.width {
			g.screen.SetContent(x, y, '.', nil, defaultStyle)
		}
	}

	// プレイヤーを描画
	myPlayerStyle := defaultStyle.Foreground(myPlayerColor)
	otherPlayerStyle := defaultStyle.Foreground(otherPlayerColor)
	for _, player := range g.players {
		style := otherPlayerStyle
		if player.ID == g.myPlayerID {
			style = myPlayerStyle
		}
		g.screen.SetContent(
			player.Position.X,
			player.Position.Y,
			getPlayerRune(player),
			nil,
			style,
		)
	}

	// アイテムを描画
	itemStyle := defaultStyle.Foreground(itemColor)
	for _, item := range g.items {
		var itemRune rune
		style := itemStyle

		switch item.Type {
		case shared.ItemType_BULLET:
			itemRune = '*'
		case shared.ItemType_BOMB:
			itemRune = '@'
			style = defaultStyle.Foreground(bombColor)
		case shared.ItemType_BOMB_FIRE:
			itemRune = '#'
			style = defaultStyle.Foreground(fireColor)
		}
		g.screen.SetContent(
			item.Position.X,
			item.Position.Y,
			itemRune,
			nil,
			style,
		)
	}

	// メッセージレートとバイトレートを画面下部に表示
	statsStr := fmt.Sprintf("Msgs: %.1f/s, KB: %.1f/s",
		g.messageStats.Rate(),
		g.messageStats.BytesRate()/1024,
	)
	style := tcell.StyleDefault.
		Foreground(tcell.ColorWhite)

	for i, r := range []rune(statsStr) {
		g.screen.SetContent(i, g.height, r, nil, style)
	}

	g.screen.Show()
}

func (g *Game) getMyPlayer() Player {
	return g.players[g.myPlayerID]
}

func (g *Game) handleMessage(message mqtt.Message) {
	// メッセージの統計情報を記録
	g.messageStats.RecordMessage(message)

	switch message.Topic() {
	case "player_state":
		playerState := &shared.PlayerState{}
		err := proto.Unmarshal(message.Payload(), playerState)
		if err != nil {
			log.Printf("Failed to unmarshal player state: %v", err)
			return
		}

		if playerState.GetStatus() == shared.Status_DISCONNECTED {
			delete(g.players, playerState.GetPlayerId())
			return
		}

		g.players[playerState.GetPlayerId()] = Player{
			ID: playerState.GetPlayerId(),
			Position: Position{
				X: int(playerState.GetPosition().GetX()),
				Y: int(playerState.GetPosition().GetY()),
			},
			Direction: playerState.GetDirection(),
			Status:    playerState.GetStatus(),
		}
	case "item_state":
		itemState := &shared.ItemState{}
		err := proto.Unmarshal(message.Payload(), itemState)
		if err != nil {
			log.Printf("Failed to unmarshal item state: %v", err)
			return
		}

		if itemState.GetStatus() == shared.ItemStatus_REMOVED {
			delete(g.items, itemState.GetItemId())
			return
		}

		g.items[itemState.GetItemId()] = Item{
			ID:   itemState.GetItemId(),
			Type: itemState.GetType(),
			Position: Position{
				X: int(itemState.GetPosition().GetX()),
				Y: int(itemState.GetPosition().GetY()),
			},
		}
	}
}

//nolint:funlen
func Run() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return errors.Wrap(err, "failed to create new screen")
	}
	defer screen.Fini()

	if err := screen.Init(); err != nil {
		return errors.Wrap(err, "failed to initialize screen")
	}

	// MQTTクライアントの設定
	clientID := uuid.New().String()
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return errors.Wrap(token.Error(), "failed to connect MQTT broker")
	}
	defer client.Disconnect(250)

	game := &Game{
		mqtt:         client,
		myPlayerID:   clientID,
		screen:       screen,
		width:        30,
		height:       30,
		players:      make(map[string]Player),
		items:        make(map[string]Item),
		messageStats: NewMessageStats(),
	}

	// プレイヤーをwidthとheightの範囲内でランダムに配置
	game.players[clientID] = Player{
		ID: clientID,
		//nolint:gosec
		Position:  Position{X: rand.Intn(game.width), Y: rand.Intn(game.height)},
		Direction: shared.Direction_UP,
		Status:    shared.Status_ALIVE,
	}

	// screenからのイベントを受け取る
	eventChan := make(chan tcell.Event)
	go func() {
		for {
			eventChan <- screen.PollEvent()
		}
	}()

	// MQTTのメッセージを受け取る
	messageChan := make(chan mqtt.Message)
	handleMessage := func(client mqtt.Client, message mqtt.Message) {
		messageChan <- message
	}
	token := game.mqtt.Subscribe("#", 0, handleMessage)
	if token.Wait() && token.Error() != nil {
		return errors.Wrap(token.Error(), "failed to subscribe to topics")
	}

	// 自分の初期位置を送信
	game.publishMyState()

	// メインループ
	ticker := time.NewTicker(50 * time.Millisecond)
	for {
		select {
		case event := <-eventChan:
			if game.handleEvent(event) {
				return nil
			}
		case message := <-messageChan:
			game.handleMessage(message)
		case <-ticker.C:
			game.messageStats.Calculate()
			game.draw()
		}
	}
}

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}
