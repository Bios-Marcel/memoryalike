package main

import (
	"testing"
)

type guessType int

const (
	none guessType = iota
	anyhiddenRune
	anyShownRune
	nonExistantRune
)

type stateIteration struct {
	hideRune                  bool
	input                     guessType
	expectedScore             int
	expectedInvalidKeyPresses int
	expectedGameState         gameState
}

// TestState tests the gamestate as a whole. E.g. simulating user interaction
// and seeing whether the results are as expected.
func TestState(t *testing.T) {
	testDifficulty := &difficulty{
		visibleName:             "easy",
		correctGuessPoints:      5,
		invalidKeyPressPenality: 2,
		rowCount:                3,
		columnCount:             2,
		runePools: [][]rune{
			runeRange('1', '6'),
		},
	}

	t.Run("No input gameover due to 50% hidden fields", func(t *testing.T) {
		iterations := []stateIteration{
			{false, none, 0, 0, ongoing},
			{true, none, 0, 0, ongoing},
			{true, none, 0, 0, ongoing},
			{true, none, 0, 0, gameOver},
		}
		state := newGameSession(make(chan bool, 100), testDifficulty)
		runIterations(t, iterations, state)
	})

	t.Run("invalid input without hidden fields", func(t *testing.T) {
		iterations := []stateIteration{
			{false, nonExistantRune, -2, 1, ongoing},
			{false, nonExistantRune, -4, 2, ongoing},
			{false, nonExistantRune, -6, 3, ongoing},
			{false, nonExistantRune, -8, 4, ongoing},
		}
		state := newGameSession(make(chan bool, 100), testDifficulty)
		runIterations(t, iterations, state)
	})

	t.Run("valid input without hidden fields", func(t *testing.T) {
		iterations := []stateIteration{
			{false, anyShownRune, -2, 1, ongoing},
			{false, anyShownRune, -4, 2, ongoing},
			{false, anyShownRune, -6, 3, ongoing},
			{false, anyShownRune, -8, 4, ongoing},
		}
		state := newGameSession(make(chan bool, 100), testDifficulty)
		runIterations(t, iterations, state)
	})

	t.Run("all guesses correct", func(t *testing.T) {
		iterations := []stateIteration{
			{true, anyhiddenRune, 5, 0, ongoing},
			{true, anyhiddenRune, 10, 0, ongoing},
			{true, anyhiddenRune, 15, 0, ongoing},
			{true, anyhiddenRune, 20, 0, ongoing},
			{true, anyhiddenRune, 25, 0, ongoing},
			{true, anyhiddenRune, 30, 0, victory},
		}
		state := newGameSession(make(chan bool, 100), testDifficulty)
		runIterations(t, iterations, state)
	})

	t.Run("one incorrect guess no gameover", func(t *testing.T) {
		iterations := []stateIteration{
			{true, anyhiddenRune, 5, 0, ongoing},
			{true, anyhiddenRune, 10, 0, ongoing},
			{true, anyhiddenRune, 15, 0, ongoing},
			{true, nonExistantRune, 13, 1, ongoing},
			{true, anyhiddenRune, 18, 1, ongoing},
			{true, anyhiddenRune, 23, 1, ongoing},
		}
		state := newGameSession(make(chan bool, 100), testDifficulty)
		runIterations(t, iterations, state)
	})

	t.Run("one incorrect guess with victory", func(t *testing.T) {
		iterations := []stateIteration{
			{true, anyhiddenRune, 5, 0, ongoing},
			{true, anyhiddenRune, 10, 0, ongoing},
			{true, anyhiddenRune, 15, 0, ongoing},
			{true, nonExistantRune, 13, 1, ongoing},
			{true, anyhiddenRune, 18, 1, ongoing},
			{true, anyhiddenRune, 23, 1, ongoing},
			{false, anyhiddenRune, 28, 1, victory},
		}
		state := newGameSession(make(chan bool, 100), testDifficulty)
		runIterations(t, iterations, state)
	})
}

func runIterations(t *testing.T, iterations []stateIteration, state *gameSession) {
	for _, iteration := range iterations {
		if iteration.hideRune {
			state.hideRune()
		}

		switch iteration.input {
		case anyhiddenRune:
			for _, cell := range state.gameBoard {
				if cell.state == hidden {
					state.inputRunePress(cell.character)
					break
				}
			}
		case anyShownRune:
			for _, cell := range state.gameBoard {
				if cell.state == shown {
					state.inputRunePress(cell.character)
					break
				}
			}
		case nonExistantRune:
			state.inputRunePress('-')
		}

		if iteration.expectedInvalidKeyPresses != state.invalidKeyPresses {
			t.Errorf("Invalid keypresses %d, expected %d", state.invalidKeyPresses, iteration.expectedInvalidKeyPresses)
		}

		if iteration.expectedScore != state.score {
			t.Errorf("score %d, expected %d", state.score, iteration.expectedScore)
		}

		if iteration.expectedGameState != state.state {
			t.Errorf("gamestate %d, expected %d", state.state, iteration.expectedGameState)
		}
	}
}
