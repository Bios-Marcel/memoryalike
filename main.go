package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell"
)

const fullBlock = '█'
const checkMark = '✓'

func main() {
	screen, screenCreationError := createScreen()
	if screenCreationError != nil {
		panic(screenCreationError)
	}

	//Cleans up the terminal buffer and returns it to the shell.
	defer screen.Fini()

	width, height := screen.Size()
	//TODO Make configurable later on
	// difficulty := 0
	cellsHorizontal := 6
	cellsVertical := 6

	//TODO make characters connected to difficulty
	gameBoard, charSetError := getCharacterSet(cellsHorizontal*cellsVertical, true, true, false)
	if charSetError != nil {
		panic(charSetError)
	}

	//preserving the original board, in order to which characters lie
	//underneath the crosses and boxes which visually represent the cell state.
	originalGameBoard := make([]rune, len(gameBoard), len(gameBoard))
	copy(originalGameBoard, gameBoard)

	//We use a mutex in order to avoid that we apply game logic while
	//key-pressed are being added to buffer and vice versa
	keyEventMutex := &sync.Mutex{}
	var keyEvents []*tcell.EventKey
	go func() {
		for {
			switch event := screen.PollEvent().(type) {
			case *tcell.EventKey:
				keyEventMutex.Lock()
				if event.Key() == tcell.KeyCtrlC {
					os.Exit(0)
				} else if event.Key() == tcell.KeyRune {
					keyEvents = append(keyEvents, event)
				}
				keyEventMutex.Unlock()
			default:
				//Unsupported or irrelevant event
			}
		}
	}()

	horizontalSpacing := 2
	verticalSpacing := 1
	boardWidth := (cellsHorizontal / 2 * (horizontalSpacing + 1))
	boardHeight := (cellsVertical / 2 * (verticalSpacing + 1))

	//Gameloop; We always draw and then check for buffered key-inputs.
	//We do the buffering in order to be able to constantly listen for
	//new keysstrokes. This should avoid lag and such.
	for {
		nextY := height/2 - boardHeight
		for y := 0; y < cellsVertical; y++ {
			nextX := width/2 - boardWidth
			for x := 0; x < cellsHorizontal; x++ {
				screen.SetCell(nextX, nextY, tcell.StyleDefault, gameBoard[x+(cellsHorizontal*y)])
				nextX += horizontalSpacing + 1
			}
			nextY += verticalSpacing + 1
		}
		screen.Show()

		keyEventMutex.Lock()
		for _, keyEvent := range keyEvents {
			keyEvent.Key()
		}
		keyEventMutex.Unlock()
	}
}

// getCharacterSet creates a unique set of characters to be used for the
// game board. The size must be greater than 0. The character used are
// digits (0 - 10), lowercase latin alphabet (a - z) and uppercase latin
// alphabet (A - Z).
func getCharacterSet(size int, digits, lowercase, uppercase bool) ([]rune, error) {
	//These all have a cell width of one. They'll be used to fill the cells
	//of the playing board.
	var availableCharacters []rune
	if digits {
		availableCharacters = append(availableCharacters,
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9')
	}

	if lowercase {
		availableCharacters = append(availableCharacters,
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z')
	}

	if uppercase {
		availableCharacters = append(availableCharacters,
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z')
	}

	if size > len(availableCharacters) {
		return nil, fmt.Errorf("the characterset can't be bigger than %d; you passed %d", len(availableCharacters), size)
	}

	if size <= 0 {
		return nil, errors.New("the request amount of characters must be greater than 0")
	}

	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(availableCharacters), func(a, b int) {
		availableCharacters[a], availableCharacters[b] = availableCharacters[b], availableCharacters[a]
	})

	return availableCharacters[0:size], nil
}

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
