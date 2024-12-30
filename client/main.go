package main

import (
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Game struct {
	screen tcell.Screen
	player struct {
		x, y int
	}
	width, height int
}

func NewGame() (*Game, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	game := &Game{
		screen: screen,
		width:  30,
		height: 30,
	}

	// プレイヤーを中央に配置
	game.player.x = game.width / 2
	game.player.y = game.height / 2

	return game, nil
}

func (g *Game) Run() {
	// 終了時に画面をクリア
	defer g.screen.Fini()

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

func (g *Game) handleEvent(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			return true
		case tcell.KeyLeft:
			if g.player.x > 0 {
				g.player.x--
			}
		case tcell.KeyRight:
			if g.player.x < g.width-1 {
				g.player.x++
			}
		case tcell.KeyUp:
			if g.player.y > 0 {
				g.player.y--
			}
		case tcell.KeyDown:
			if g.player.y < g.height-1 {
				g.player.y++
			}
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
	playerStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	g.screen.SetContent(g.player.x, g.player.y, '◎', nil, playerStyle)

	g.screen.Show()
}

func main() {
	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}

	game.Run()
}
