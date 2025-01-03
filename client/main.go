package main

import (
	"log"
	"math/rand"
	"time"

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
	ID       string
	Position *Position
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
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	// MQTTクライアントの設定
	clientID := uuid.New().String()
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
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
		Position: &Position{X: rand.Intn(game.width), Y: rand.Intn(game.height)},
	}

	// 全てのtopicをsubscribeする
	token := game.mqtt.Subscribe("#", 0, game.handleMessage)
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
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

func (g *Game) handleEvent(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventKey:
		myPlayer := g.getMyPlayer()
		oldX, oldY := myPlayer.Position.X, myPlayer.Position.Y

		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			return true
		case tcell.KeyLeft:
			if myPlayer.Position.X > 0 {
				myPlayer.Position.X--
			}
		case tcell.KeyRight:
			if myPlayer.Position.X < g.width-1 {
				myPlayer.Position.X++
			}
		case tcell.KeyUp:
			if myPlayer.Position.Y > 0 {
				myPlayer.Position.Y--
			}
		case tcell.KeyDown:
			if myPlayer.Position.Y < g.height-1 {
				myPlayer.Position.Y++
			}
		}

		// 位置が変更されたら自分の位置をサーバーに送る
		if oldX != myPlayer.Position.X || oldY != myPlayer.Position.Y {
			g.publishMyState()
		}
	}
	return false
}

func (g *Game) draw() {
	g.screen.Clear()

	// マップの境界を描画
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			g.screen.SetContent(x, y, '.', nil, style)
		}
	}

	// プレイヤーを描画
	myPlayerStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	otherPlayerStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	for _, player := range g.players {
		if player.ID == g.myPlayerID {
			g.screen.SetContent(player.Position.X, player.Position.Y, '◎', nil, myPlayerStyle)
		} else {
			g.screen.SetContent(player.Position.X, player.Position.Y, '◎', nil, otherPlayerStyle)
		}
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

		if playerState.Status == shared.Status_DISCONNECTED {
			delete(g.players, playerState.PlayerId)
			return
		}

		g.players[playerState.PlayerId] = &Player{
			ID:       playerState.PlayerId,
			Position: &Position{X: int(playerState.Position.X), Y: int(playerState.Position.Y)},
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
