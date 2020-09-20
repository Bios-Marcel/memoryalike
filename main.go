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

const (
	fullBlock = '█'
	checkMark = '✓'
)

type gameState int

const (
	ongoing = iota
	gameOver
	victory
)

type sessionState struct {
	mutex *sync.Mutex

	currentGameState gameState
	score            int

	gameBoard     []rune
	indicesToHide []int
	runePositions map[rune]int

	cellsHorizontal int
	cellsVertical   int
}

func main() {
	screen, screenCreationError := createScreen()
	if screenCreationError != nil {
		panic(screenCreationError)
	}

	//Cleans up the terminal buffer and returns it to the shell.
	defer screen.Fini()

	width, height := screen.Size()
	currentSessionState := newSessionState(width, height, difficulty)

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
					screen.Fini()
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
	boardWidth := (currentSessionState.cellsHorizontal / 2 * (horizontalSpacing + 1))
	boardHeight := (currentSessionState.cellsVertical / 2 * (verticalSpacing + 1))

	//Gameloop; We always draw and then check for buffered key-inputs.
	//We do the buffering in order to be able to constantly listen for
	//new keysstrokes. This should avoid lag and such.

	//One frame each 1/60 of a second. E.g. we want 60 FPS.
	//TODO Is there a better and more modern approach for this?
	gameLoopTicker := time.NewTicker(1 * time.Second / 60)

	for {
		<-gameLoopTicker.C
		currentSessionState.mutex.Lock()

		switch currentSessionState.currentGameState {
		case ongoing:
			//Draw gameBoard to screen. This block contains no game-logic.
			nextY := height/2 - boardHeight
			for y := 0; y < currentSessionState.cellsVertical; y++ {
				nextX := width/2 - boardWidth
				for x := 0; x < currentSessionState.cellsHorizontal; x++ {
					cellRune := currentSessionState.gameBoard[x+(currentSessionState.cellsHorizontal*y)]
					screen.SetCell(nextX, nextY, tcell.StyleDefault, cellRune)
					nextX += horizontalSpacing + 1
				}
				nextY += verticalSpacing + 1
			}
		case victory:
			screen.Clear()
		case gameOver:
			screen.Clear()
		}
		screen.Show()

		if currentSessionState.currentGameState == ongoing {
			keyEventMutex.Lock()
			currentSessionState.applyKeyEvents(keyEvents)
			//Clearing, in order to prevent, that the user can just hit all keys
			//beforehand and wait in order to win.
			keyEvents = keyEvents[:0]
			keyEventMutex.Unlock()

			currentSessionState.updateGameState()
		}

		currentSessionState.mutex.Unlock()
	}
}

// newSessionState produces a ready-to-use session state. The ticker that
// hides cell contents is started on construction.
func newSessionState(width, height, difficulty int) *sessionState {
	var cellsHorizontal int
	var cellsVertical int
	var hideTimes time.Duration
	var useDigits bool
	var useLowercaseChars bool
	var useUppercaseChars bool

	switch difficulty {
	//easy
	case 0:
		cellsHorizontal = 3
		cellsVertical = 3
		//Speed is higher, as we only got digits.
		hideTimes = 1250 * time.Millisecond
		useDigits = true
		useLowercaseChars = false
		useUppercaseChars = false
		//normal
	case 1:
		cellsHorizontal = 4
		cellsVertical = 3
		hideTimes = 2500 * time.Millisecond
		useDigits = true
		useLowercaseChars = true
		useUppercaseChars = false
	//hard
	case 2:
		cellsHorizontal = 4
		cellsVertical = 4
		hideTimes = 2000 * time.Millisecond
		useDigits = true
		useLowercaseChars = true
		useUppercaseChars = false
	//extreme
	case 3:
		cellsHorizontal = 4
		cellsVertical = 4
		hideTimes = 1750 * time.Millisecond
		useDigits = true
		useLowercaseChars = true
		useUppercaseChars = true
	}

	gameBoard, charSetError := getCharacterSet(cellsHorizontal*cellsVertical, useDigits, useLowercaseChars, useUppercaseChars)
	if charSetError != nil {
		panic(charSetError)
	}

	//Position cache to avoid performance loss due to iteration.
	runePositions := make(map[rune]int, len(gameBoard))
	for index, char := range gameBoard {
		runePositions[char] = index
	}

	//This decides which cells will be hidden in which order. If this stack
	//is empty, the game is over.
	indicesToHide := make([]int, len(gameBoard), len(gameBoard))
	for i := 0; i < len(indicesToHide); i++ {
		indicesToHide[i] = i
	}
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(indicesToHide), func(a, b int) {
		indicesToHide[a], indicesToHide[b] = indicesToHide[b], indicesToHide[a]
	})

	newSessionState := &sessionState{
		mutex:            &sync.Mutex{},
		indicesToHide:    indicesToHide,
		runePositions:    runePositions,
		gameBoard:        gameBoard,
		currentGameState: ongoing,
		cellsHorizontal:  cellsHorizontal,
		cellsVertical:    cellsVertical,
	}

	//This hides characters according to the timeframes decided
	//by the difficulty level.
	characterHideTicker := time.NewTicker(hideTimes)
	go func() {
		for {
			<-characterHideTicker.C

			newSessionState.mutex.Lock()

			index := len(indicesToHide) - 1
			gameBoard[indicesToHide[index]] = fullBlock
			indicesToHide = indicesToHide[:len(indicesToHide)-1]

			newSessionState.mutex.Unlock()

			if len(indicesToHide) == 0 {
				characterHideTicker.Stop()
				break
			}
		}
	}()

	return newSessionState
}

// applyKeyEvents checks the key-events for possible matches and updates the
// sessionState accordingly. Meaning that if a match between a hidden
// cell, it's underlying character and the input rune is found, the player
// gets a point.
func (s *sessionState) applyKeyEvents(keyEvents []*tcell.EventKey) {
	for _, keyEvent := range keyEvents {
		//We assume that we only have KeyRune events here, as they were
		//already pre-checked during the polling.
		runeIndex := s.runePositions[keyEvent.Rune()]

		//Correct match, therefore replace fullBlock with checkmark to
		//mark cell as "correctly guessed".
		if s.gameBoard[runeIndex] == fullBlock {
			s.gameBoard[runeIndex] = checkMark
		}
	}

}

// updateGameState determines whether the game is over and what the players
// score is.
func (s *sessionState) updateGameState() {
	checkMarkCount := 0
	fullBlockCount := 0
	leftOverChars := 0

	for _, char := range s.gameBoard {
		if char == fullBlock {
			fullBlockCount++
		} else if char == checkMark {
			checkMarkCount++
		} else {
			leftOverChars++
		}
	}

	s.score = checkMarkCount

	//if at least 40 percent of the board is fullblocks, the player lost.
	//In case of a normal game for example, this should mean 4 fullBlocks.
	if fullBlockCount != 0 && len(s.gameBoard)/fullBlockCount <= 4 {
		s.currentGameState = gameOver
	} else if leftOverChars == 0 {
		s.currentGameState = victory
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
