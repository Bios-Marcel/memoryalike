package main

import (
	"fmt"

	"github.com/gdamore/tcell"
)

const (
	chooseDifficultyText = "Choose difficulty"
	gameOverMessage      = "GAME OVER"
	victoryMessage       = "Congratulations! You have won!"
	restartMessage       = "Hit 'Ctrl R' to restart or 'ESC' to show the menu."

	fullBlock = '█'
	checkMark = '✓'
)

var titleStyle = tcell.StyleDefault.Bold(true)

// renderer represents a utility object to present a gameSession on a
// terminal screen.
type renderer struct {
	horizontalSpacing int
	verticalSpacing   int
}

// newRenderer creates a new reusable renderer. It can be used for any
// gameSession and any screen. It is also able to draw the game menu.
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

	unselectedStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault.Reverse(true)

	determineStyle := func(difficulty int) tcell.Style {
		if sourceMenuState.selectedDifficulty == difficulty {
			return selectedStyle
		}

		return unselectedStyle
	}

	screenWidth, _ := targetScreen.Size()

	//Draw "Choose difficulties text"
	r.printStyledLine(targetScreen, chooseDifficultyText, titleStyle,
		getHorizontalCenterForText(screenWidth, chooseDifficultyText), 2)

	//Draw difficulties into menu.
	nextY := 4
	for diffIndex, diff := range difficulties {
		r.printStyledLine(targetScreen, diff.visibleName, determineStyle(diffIndex),
			getHorizontalCenterForText(screenWidth, diff.visibleName), nextY)
		nextY += 2
	}

	targetScreen.Show()
}

// getHorizontalCenterForText returns the x-coordinate at which the caller must
// start drawing in order to horizontally center given text. Note that this
// function doesn't take rune-width into count, as it is currently irrelevant.
func getHorizontalCenterForText(screenWidth int, text string) int {
	return screenWidth/2 - len(text)/2
}

// drawGameBoard fills the targetScreen with data from the passed gameSession.
func (r *renderer) drawGameBoard(targetScreen tcell.Screen, session *gameSession) {
	boardWidth := (session.difficulty.rowCount / 2 * (r.horizontalSpacing + 1))
	boardHeight := (session.difficulty.columnCount / 2 * (r.verticalSpacing + 1))

	width, height := targetScreen.Size()

	//Draw gameBoard to screen. This block contains no game-logic.
	//We draw this regardless of the game state, since the player
	//wouldn't be able to see the effect of their last move otherwise.
	nextY := height/2 - boardHeight
	for y := 0; y < session.difficulty.columnCount; y++ {
		nextX := width/2 - boardWidth
		for x := 0; x < session.difficulty.rowCount; x++ {
			var renderRune rune
			boardCell := session.gameBoard[x+(session.difficulty.rowCount*y)]
			switch boardCell.state {
			case shown:
				renderRune = boardCell.character
			case hidden:
				renderRune = fullBlock
			case guessed:
				renderRune = checkMark
			}

			targetScreen.SetContent(nextX, nextY, renderRune, nil, tcell.StyleDefault)
			nextX += r.horizontalSpacing + 1
		}
		nextY += r.verticalSpacing + 1
	}

	switch session.state {
	case victory:
		r.printStyledLine(targetScreen, victoryMessage, titleStyle, width/2-len(victoryMessage)/2, 2)
	case gameOver:
		r.printStyledLine(targetScreen, gameOverMessage, titleStyle, width/2-len(gameOverMessage)/2, 2)
	}

	if session.state != ongoing {
		r.printGameResults(width, targetScreen, session)
	}

	targetScreen.Show()
}

// printGameResults prints the score, amount of invalid key presses and
// information on how to restart or get to the menu.
func (r *renderer) printGameResults(width int, targetScreen tcell.Screen, session *gameSession) {
	scoreMessage := r.createScoreMessage(session)
	r.printLine(targetScreen, scoreMessage, width/2-len(scoreMessage)/2, 4)
	invalidKeyPressesMessage := r.createInvalidKeyPressesMessage(session)
	r.printLine(targetScreen, invalidKeyPressesMessage, width/2-len(invalidKeyPressesMessage)/2, 5)
	r.printLine(targetScreen, restartMessage, width/2-len(restartMessage)/2, 7)
}

func (r *renderer) createInvalidKeyPressesMessage(session *gameSession) string {
	return fmt.Sprintf("Amount of invalid key presses: %d", session.invalidKeyPresses)
}

func (r *renderer) createScoreMessage(session *gameSession) string {
	return fmt.Sprintf("Your score is %d out of possible %d",
		session.score, len(session.gameBoard)*session.difficulty.correctGuessPoints)
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
