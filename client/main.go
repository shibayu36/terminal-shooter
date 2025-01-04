package main

import (
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
	Position  *Position
	Direction shared.Direction
}

type Game struct {
	mqtt mqtt.Client

	screen tcell.Screen

	myPlayerID string
	players    map[string]*Player
	width      int
	height     int
}

func NewGame() (*Game, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new screen")
	}

	if err := screen.Init(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize screen")
	}

	// MQTTクライアントの設定
	clientID := uuid.New().String()
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, errors.Wrap(token.Error(), "failed to connect MQTT broker")
	}

	game := &Game{
		mqtt:       client,
		myPlayerID: clientID,
		screen:     screen,
		width:      30,
		height:     30,
		players:    make(map[string]*Player),
	}

	// プレイヤーをwidthとheightの範囲内でランダムに配置
	game.players[clientID] = &Player{
		ID: clientID,
		//nolint:gosec
		Position:  &Position{X: rand.Intn(game.width), Y: rand.Intn(game.height)},
		Direction: shared.Direction_UP,
	}

	// 全てのtopicをsubscribeする
	token := game.mqtt.Subscribe("#", 0, game.handleMessage)
	if token.Wait() && token.Error() != nil {
		return nil, errors.Wrap(token.Error(), "failed to subscribe to topics")
	}

	// 自分の初期位置を送信
	game.publishMyState()
	return game, nil
}

func (g *Game) Run() {
	// 終了時に接続や画面をクリア
	defer func() {
		g.screen.Fini()
		g.mqtt.Disconnect(250)
	}()

	// イベントチャンネル
	eventChan := make(chan tcell.Event)
	go func() {
		for {
			eventChan <- g.screen.PollEvent()
		}
	}()

	// メインループ
	ticker := time.NewTicker(50 * time.Millisecond)
	for {
		select {
		case event := <-eventChan:
			if g.handleEvent(event) {
				return
			}
		case <-ticker.C:
			g.draw()
		}
	}
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

func (g *Game) movePlayer(dx, dy int) {
	myPlayer := g.getMyPlayer()
	oldX, oldY := myPlayer.Position.X, myPlayer.Position.Y
	oldDirection := myPlayer.Direction

	// 移動方向からdirectionを決定
	var direction shared.Direction
	switch {
	case dx < 0:
		direction = shared.Direction_LEFT
	case dx > 0:
		direction = shared.Direction_RIGHT
	case dy < 0:
		direction = shared.Direction_UP
	case dy > 0:
		direction = shared.Direction_DOWN
	default:
		direction = myPlayer.Direction // 移動がない場合は現在の向きを維持
	}

	if newX := myPlayer.Position.X + dx; newX >= 0 && newX < g.width {
		myPlayer.Position.X = newX
	}
	if newY := myPlayer.Position.Y + dy; newY >= 0 && newY < g.height {
		myPlayer.Position.Y = newY
	}
	myPlayer.Direction = direction

	// 位置か方向が変更されたら自分の状態をサーバーに送る
	if oldX != myPlayer.Position.X || oldY != myPlayer.Position.Y || oldDirection != direction {
		g.publishMyState()
	}
}

func (g *Game) handleEvent(event tcell.Event) bool {
	//nolint:gocritic // ignore singleCaseSwitch
	switch ev := event.(type) {
	case *tcell.EventKey:
		//nolint:exhaustive
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			return true
		case tcell.KeyLeft:
			g.movePlayer(-1, 0)
		case tcell.KeyRight:
			g.movePlayer(1, 0)
		case tcell.KeyUp:
			g.movePlayer(0, -1)
		case tcell.KeyDown:
			g.movePlayer(0, 1)
		}
	}
	return false
}

func getDirectionRune(direction shared.Direction) rune {
	switch direction {
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

func (g *Game) draw() {
	g.screen.Clear()

	// マップの境界を描画
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	for y := range g.height {
		for x := range g.width {
			g.screen.SetContent(x, y, '.', nil, style)
		}
	}

	// プレイヤーを描画
	myPlayerStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	otherPlayerStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	for _, player := range g.players {
		style := otherPlayerStyle
		if player.ID == g.myPlayerID {
			style = myPlayerStyle
		}
		g.screen.SetContent(
			player.Position.X,
			player.Position.Y,
			getDirectionRune(player.Direction),
			nil,
			style,
		)
	}

	g.screen.Show()
}

func (g *Game) getMyPlayer() *Player {
	return g.players[g.myPlayerID]
}

func (g *Game) handleMessage(client mqtt.Client, message mqtt.Message) {
	if message.Topic() == "player_state" {
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

		g.players[playerState.GetPlayerId()] = &Player{
			ID: playerState.GetPlayerId(),
			Position: &Position{
				X: int(playerState.GetPosition().GetX()),
				Y: int(playerState.GetPosition().GetY()),
			},
			Direction: playerState.GetDirection(),
		}
	}
}

func main() {
	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}

	game.Run()
}
