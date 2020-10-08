package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type cellState int

const (
	shown cellState = iota
	hidden
	guessed
)

// gameBoardCell represents a single character visible to the user.
type gameBoardCell struct {
	character rune
	state     cellState
}

type gameState int

const (
	ongoing = iota
	gameOver
	victory
)

// gameSession represents all game state for a session. All operations on
// this state should make sure that the state is locked using the internal
// mutex.
type gameSession struct {
	mutex                     *sync.Mutex
	renderNotificationChannel chan bool

	state gameState
	score int
	//invalidKeyPresses counts the invalid keyPresses made by the player.
	//This only tracks runes, not stuff like CTRL, ArrowUp ...
	invalidKeyPresses int

	gameBoard     []*gameBoardCell
	indicesToHide []int

	difficulty *difficulty
}

// newGameSession produces a ready-to-use session state. The ticker that
// hides cell contents is started on construction.
func newGameSession(renderNotificationChannel chan bool, difficulty *difficulty) *gameSession {
	characterSet, charSetError := getCharacterSet(difficulty.rowCount*difficulty.columnCount, difficulty.runePools...)
	if charSetError != nil {
		panic(charSetError)
	}
	gameBoard := make([]*gameBoardCell, 0, len(characterSet))
	for _, char := range characterSet {
		gameBoard = append(gameBoard, &gameBoardCell{char, shown})
	}

	//This decides which cells will be hidden in which order. If this stack
	//is empty, the game is over.
	indicesToHide := make([]int, len(gameBoard))
	for i := 0; i < len(indicesToHide); i++ {
		indicesToHide[i] = i
	}
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(indicesToHide), func(a, b int) {
		indicesToHide[a], indicesToHide[b] = indicesToHide[b], indicesToHide[a]
	})

	return &gameSession{
		mutex:                     &sync.Mutex{},
		renderNotificationChannel: renderNotificationChannel,

		state: ongoing,

		gameBoard:     gameBoard,
		indicesToHide: indicesToHide,

		difficulty: difficulty,
	}
}

// startRuneHidingCoroutine starts a goroutine that hides one rune on the
// gameboard each X milliseconds. X is defined by the hidingTime defined in
// the referenced difficulty of the session. If no more characters can be
// hidden or the game has ended, this coroutine exists.
func (s *gameSession) startRuneHidingCoroutine() {
	go func() {
		<-time.NewTimer(s.difficulty.startDelay).C

		characterHideTicker := time.NewTicker(s.difficulty.hideTimes)
		for {
			<-characterHideTicker.C

			if len(s.indicesToHide) == 0 || s.state != ongoing {
				characterHideTicker.Stop()
				break
			}

			s.mutex.Lock()
			s.hideRune()
			s.mutex.Unlock()
		}
	}()
}

// hideRune hides a rune that's currently visible on the gameboard.
func (s *gameSession) hideRune() {
	nextIndexToHide := len(s.indicesToHide) - 1
	if nextIndexToHide != -1 {
		s.gameBoard[s.indicesToHide[nextIndexToHide]].state = hidden
		s.indicesToHide = s.indicesToHide[:len(s.indicesToHide)-1]
		s.updateGameState()
	}
}

// applyKeyEvents checks the key-events for possible matches and updates the
// gameSession accordingly. Meaning that if a match between a hidden
// cell, it's underlying character and the input rune is found, the player
// gets a point.
func (s *gameSession) inputRunePress(pressed rune) {
	//Game is already over. All further checks are unnecessary.
	if s.state != ongoing {
		return
	}

	for _, cell := range s.gameBoard {
		if cell.character == pressed {
			if cell.state == hidden {
				cell.state = guessed
				s.updateGameState()
				return
			}

			break
		}
	}

	//Pressed rune wasn't hidden or wasn't present, therefore the user gets
	//minus points
	s.invalidKeyPresses++
	s.updateGameState()
}

// updateGameState determines whether the game is over and what the players
// score is.
func (s *gameSession) updateGameState() {
	//Game is already over. All further checks are unnecessary.
	if s.state != ongoing {
		return
	}

	var guessedCellCount, hiddenCellCount, shownCellCount int
	for _, cell := range s.gameBoard {
		if cell.state == hidden {
			hiddenCellCount++
		} else if cell.state == guessed {
			guessedCellCount++
		} else {
			shownCellCount++
		}
	}

	s.score = guessedCellCount*s.difficulty.correctGuessPoints -
		s.invalidKeyPresses*s.difficulty.invalidKeyPressPenality

	//if at least 40 percent of the board is hidden, the player loses.
	//In case of a normal game for example, this should mean 4 hidden cells.
	if hiddenCellCount != 0 && float32(hiddenCellCount)/float32(len(s.gameBoard)) >= 0.4 {
		s.state = gameOver
	} else if shownCellCount == 0 && hiddenCellCount == 0 {
		//The game is only over if all cells have been guessed correctly

		//Even if all cells have been guessed correctly, we deem zero score
		//as a loss, as the player probably smashed his keyboard randomly.
		if s.score <= 0 {
			s.state = gameOver
		} else {
			s.state = victory
		}
	}

	// In order to avoid dead-locking the caller.
	go func() {
		s.renderNotificationChannel <- true
	}()
}

// runeRange creates a new rune array containing all the runes between the
// two passed ones. Both from and to are inclusive.
func runeRange(from, to rune) []rune {
	runes := make([]rune, 0, to-from+1)
	for r := from; r <= to; r++ {
		runes = append(runes, r)
	}
	return runes
}

// getCharacterSet creates a unique set of characters to be used for the
// game board. The size must be greater than 0. For sourcing the
// characters, the rune arrays passed to this method will be used.
func getCharacterSet(size int, pools ...[]rune) ([]rune, error) {
	var availableCharacters []rune
	for _, pool := range pools {
		availableCharacters = append(availableCharacters, pool...)
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
