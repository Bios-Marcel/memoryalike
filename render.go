package main

import (
	"fmt"

	"github.com/gdamore/tcell"
)

const (
	gameOverMessage = "GAME OVER"
	victoryMessage  = "Congratulations! You have won!"
	restartMessage  = "Hit 'Ctrl R' to restart or 'ESC' to show the menu."

	chooseDifficultyText    = "Choose difficulty"
	easyDifficultyText      = "easy"
	normalDifficultyText    = "normal"
	hardDifficultyText      = "hard"
	extremeDifficultyText   = "extreme"
	nightmareDifficultyText = "nightmare"
)

// renderer represents a utility object to present a sessionState on a
// terminal screen.
type renderer struct {
	horizontalSpacing int
	verticalSpacing   int
}

// newRenderer creates a new reusable renderer. It can be used for any
// sessionState and any screen. It is also able to draw the game menu.
// The renderer itself is stateless, which is why it can be used for
// multiple sessions and screens. Technically, you could draw on multiple
// screens at once.
func newRenderer() *renderer {
	return &renderer{
		horizontalSpacing: 2,
		verticalSpacing:   1,
	}
}

// drawMenu draws the main menu of the game. It allows for selecting the
// difficulty. Selected menu entries are rendered with the reverse attribute
// activated.
func (r *renderer) drawMenu(targetScreen tcell.Screen, sourceMenuState *menuState) {
	targetScreen.Clear()

	instructionStyle := tcell.StyleDefault.Bold(true)
	unselectedStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault.Reverse(true)

	determineStyle := func(difficulty int) tcell.Style {
		if sourceMenuState.selectedDifficulty == difficulty {
			return selectedStyle
		}

		return unselectedStyle
	}

	screenWidth, _ := targetScreen.Size()
	r.printStyledLine(targetScreen, chooseDifficultyText, instructionStyle,
		getHorizontalCenterForText(screenWidth, chooseDifficultyText), 2)
	r.printStyledLine(targetScreen, easyDifficultyText, determineStyle(0),
		getHorizontalCenterForText(screenWidth, easyDifficultyText), 4)
	r.printStyledLine(targetScreen, normalDifficultyText, determineStyle(1),
		getHorizontalCenterForText(screenWidth, normalDifficultyText), 6)
	r.printStyledLine(targetScreen, hardDifficultyText, determineStyle(2),
		getHorizontalCenterForText(screenWidth, hardDifficultyText), 8)
	r.printStyledLine(targetScreen, extremeDifficultyText, determineStyle(3),
		getHorizontalCenterForText(screenWidth, extremeDifficultyText), 10)
	r.printStyledLine(targetScreen, nightmareDifficultyText, determineStyle(4),
		getHorizontalCenterForText(screenWidth, nightmareDifficultyText), 12)

	targetScreen.Show()
}

// getHorizontalCenterForText returns the x-coordinate at which teh caller must
// start drawing in order to horizontally center given text. Note that this
// function doesn't take rune-width into count, as it is currently irrelevant.
func getHorizontalCenterForText(screenWidth int, text string) int {
	return screenWidth/2 - len(text)/2
}

// drawGameBoard fills the targetScreen with data from the passed sessionState.
func (r *renderer) drawGameBoard(targetScreen tcell.Screen, session *sessionState) {
	boardWidth := (session.cellsHorizontal / 2 * (r.horizontalSpacing + 1))
	boardHeight := (session.cellsVertical / 2 * (r.verticalSpacing + 1))

	width, height := targetScreen.Size()

	//Draw gameBoard to screen. This block contains no game-logic.
	//We draw this regardless of the game state, since the player
	//wouldn't be able to see the effect of their last move otherwise.
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

	switch session.currentGameState {
	case victory:
		r.printLine(targetScreen, victoryMessage, width/2-len(victoryMessage)/2, 2)
		scoreMessage := r.createScoreMessage(session)
		r.printLine(targetScreen, scoreMessage, width/2-len(scoreMessage)/2, 4)
		invalidKeyPressesMessage := r.createInvalidKeyPressesMessage(session)
		r.printLine(targetScreen, invalidKeyPressesMessage, width/2-len(invalidKeyPressesMessage)/2, 5)
		r.printLine(targetScreen, restartMessage, width/2-len(restartMessage)/2, 7)
	case gameOver:
		r.printLine(targetScreen, gameOverMessage, width/2-len(gameOverMessage)/2, 2)
		scoreMessage := r.createScoreMessage(session)
		r.printLine(targetScreen, scoreMessage, width/2-len(scoreMessage)/2, 4)
		invalidKeyPressesMessage := r.createInvalidKeyPressesMessage(session)
		r.printLine(targetScreen, invalidKeyPressesMessage, width/2-len(invalidKeyPressesMessage)/2, 5)
		r.printLine(targetScreen, restartMessage, width/2-len(restartMessage)/2, 7)
	}
	targetScreen.Show()
}

func (r *renderer) createInvalidKeyPressesMessage(session *sessionState) string {
	return fmt.Sprintf("Amount of invalid key presses: %d", session.invalidKeyPresses)
}

func (r *renderer) createScoreMessage(session *sessionState) string {
	return fmt.Sprintf("Your score is %d out of possible %d",
		session.score, len(session.gameBoard)*scorePerGuess)
}

// printLine draws the given text at the desired position. The text will be
// drawn using the default style (tcell.StyleDefault).
func (r *renderer) printLine(targetScreen tcell.Screen, message string, x, y int) {
	r.printStyledLine(targetScreen, message, tcell.StyleDefault, x, y)
}

// printStyledLine is the same as printLine, but you can override the default
// text style.
func (r *renderer) printStyledLine(targetScreen tcell.Screen, message string, style tcell.Style, x, y int) {
	nextX := x
	for _, char := range message {
		targetScreen.SetContent(nextX, y, char, nil, style)
		nextX++
	}
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
