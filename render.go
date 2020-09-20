package main

import "github.com/gdamore/tcell"

const (
	gameOverMessage = "GAME OVER"
	victoryMessage  = "Congratulations! You have won!"
)

// renderer represents a utility object to present a sessionState on a
// terminal screen.
type renderer struct {
	horizontalSpacing int
	verticalSpacing   int
}

// newRenderer creates a new reusable renderer. It can be used for any
// sessionState and any screen. The renderer is stateless.
func newRenderer() *renderer {
	return &renderer{
		horizontalSpacing: 2,
		verticalSpacing:   1,
	}
}

// draw fills the targetScreen with data from the passed sessionState.
func (r *renderer) draw(targetScreen tcell.Screen, session *sessionState) {
	boardWidth := (session.cellsHorizontal / 2 * (r.horizontalSpacing + 1))
	boardHeight := (session.cellsVertical / 2 * (r.verticalSpacing + 1))

	width, height := targetScreen.Size()

	switch session.currentGameState {
	case ongoing:
		//Draw gameBoard to screen. This block contains no game-logic.
		nextY := height/2 - boardHeight
		for y := 0; y < session.cellsVertical; y++ {
			nextX := width/2 - boardWidth
			for x := 0; x < session.cellsHorizontal; x++ {
				cellRune := session.gameBoard[x+(session.cellsHorizontal*y)]
				targetScreen.SetContent(nextX, nextY, cellRune, nil, tcell.StyleDefault)
				nextX += r.horizontalSpacing + 1
			}
			nextY += r.verticalSpacing + 1
		}
	case victory:
		targetScreen.Clear()
		nextX := 0
		for _, char := range victoryMessage {
			targetScreen.SetContent(nextX, 0, char, nil, tcell.StyleDefault)
			nextX++
		}
	case gameOver:
		targetScreen.Clear()
		nextX := 0
		for _, char := range gameOverMessage {
			targetScreen.SetContent(nextX, 0, char, nil, tcell.StyleDefault)
			nextX++
		}
	}
	targetScreen.Show()
}

// createScreen generates a ready to use screen. The screen has
// no cursor and doesn't support mouse eventing.
func createScreen() (tcell.Screen, error) {
	screen, screenCreationError := tcell.NewScreen()
	if screenCreationError != nil {
		return nil, screenCreationError
	}

	screenInitError := screen.Init()
	if screenInitError != nil {
		return nil, screenInitError
	}

	//Make sure it's disable, even though it should be by default.
	screen.DisableMouse()
	//Make sure cursor is hidden by default.
	screen.HideCursor()

	return screen, nil
}
