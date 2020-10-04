package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gdamore/tcell"
)

const scorePerGuess = 5

// sessionState represents all game state for a session. All operations on
// this state should make sure that the state is locked using the internal
// mutex.
type sessionState struct {
	mutex                     *sync.Mutex
	renderNotificationChannel chan bool

	currentGameState gameState
	score            int
	//invalidKeyPresses counts the invalid keyPresses made by the player.
	invalidKeyPresses int

	gameBoard     []rune
	indicesToHide []int
	runePositions map[rune]int

	cellsHorizontal int
	cellsVertical   int
}

// newSessionState produces a ready-to-use session state. The ticker that
// hides cell contents is started on construction.
func newSessionState(renderNotificationChannel chan bool,
	width, height int, difficulty *difficulty) *sessionState {
	gameBoard, charSetError := getCharacterSet(difficulty.rowCount*difficulty.columnCount, difficulty.runePools...)
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
		mutex:                     &sync.Mutex{},
		renderNotificationChannel: renderNotificationChannel,

		currentGameState: ongoing,

		gameBoard:     gameBoard,
		indicesToHide: indicesToHide,
		runePositions: runePositions,

		cellsHorizontal: difficulty.rowCount,
		cellsVertical:   difficulty.columnCount,
	}

	//This hides characters according to the timeframes decided
	//by the difficulty level.
	go func() {
		//FIXME Consider whether to make this difficulty dependant.
		//Before we start the actual countdown to hiding characters, we wait
		//for a short while to make it a bit easier on the user.
		<-time.NewTimer(difficulty.startDelay).C

		characterHideTicker := time.NewTicker(difficulty.hideTimes)
		for {
			<-characterHideTicker.C

			if len(indicesToHide) == 0 || newSessionState.currentGameState != ongoing {
				characterHideTicker.Stop()
				break
			}

			newSessionState.mutex.Lock()

			index := len(indicesToHide) - 1
			gameBoard[indicesToHide[index]] = fullBlock
			indicesToHide = indicesToHide[:len(indicesToHide)-1]
			newSessionState.updateGameState()

			newSessionState.mutex.Unlock()
		}
	}()

	return newSessionState
}

// applyKeyEvents checks the key-events for possible matches and updates the
// sessionState accordingly. Meaning that if a match between a hidden
// cell, it's underlying character and the input rune is found, the player
// gets a point.
func (s *sessionState) applyKeyEvent(keyEvent *tcell.EventKey) {
	//Game is already over. All further checks are unnecessary.
	if s.currentGameState != ongoing {
		return
	}

	//We assume that we only have KeyRune events here, as they were
	//already pre-checked during the polling.
	runeIndex := s.runePositions[keyEvent.Rune()]

	//Correct match, therefore replace fullBlock with checkmark to
	//mark cell as "correctly guessed".
	if s.gameBoard[runeIndex] == fullBlock {
		s.gameBoard[runeIndex] = checkMark
	} else {
		s.invalidKeyPresses++
	}
	s.updateGameState()
}

// updateGameState determines whether the game is over and what the players
// score is.
func (s *sessionState) updateGameState() {
	//Game is already over. All further checks are unnecessary.
	if s.currentGameState != ongoing {
		return
	}

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

	s.score = checkMarkCount*scorePerGuess - s.invalidKeyPresses*2

	//if at least 40 percent of the board is fullblocks, the player lost.
	//In case of a normal game for example, this should mean 4 fullBlocks.
	lengthFloat := float32(len(s.gameBoard))
	fullBlockCountFloat := float32(fullBlockCount)
	if fullBlockCount != 0 && fullBlockCountFloat/lengthFloat >= 0.4 {
		s.currentGameState = gameOver
	} else if leftOverChars == 0 && fullBlockCount == 0 {
		//The game is only over, if there are no full blocks left and no
		//unmasked cells left.
		if s.score <= 0 {
			s.currentGameState = gameOver
		} else {
			s.currentGameState = victory
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
