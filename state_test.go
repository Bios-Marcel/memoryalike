package main

import "testing"

type guessType int

const (
	none guessType = iota
	correct
	incorrect
)

type stateIteration struct {
	hideRune                  bool
	input                     guessType
	expectedScore             int
	expectedInvalidKeyPresses int
	expectedGameState         gameState
}

func TestState(t *testing.T) {
	t.Run("No input gameover due to 50% hidden fields", func(t *testing.T) {
		iterations := []stateIteration{
			{false, none, 0, 0, ongoing},
			{true, none, 0, 0, ongoing},
			{true, none, 0, 0, ongoing},
			{true, none, 0, 0, gameOver},
		}
		state := newSessionState(make(chan bool, 100), difficulties[0])
		runIterations(t, iterations, state)
	})

	t.Run("invalid niput without hidden fields", func(t *testing.T) {
		iterations := []stateIteration{
			{false, incorrect, -2, 1, ongoing},
			{false, incorrect, -4, 2, ongoing},
			{false, incorrect, -6, 3, ongoing},
			{false, incorrect, -8, 4, gameOver},
		}
		state := newSessionState(make(chan bool, 100), difficulties[0])
		runIterations(t, iterations, state)
	})

	t.Run("all guesses correct", func(t *testing.T) {
		iterations := []stateIteration{
			{true, correct, 5, 0, ongoing},
			{true, correct, 10, 0, ongoing},
			{true, correct, 15, 0, ongoing},
			{true, correct, 20, 0, ongoing},
			{true, correct, 25, 0, ongoing},
			{true, correct, 30, 0, victory},
		}
		state := newSessionState(make(chan bool, 100), difficulties[0])
		runIterations(t, iterations, state)
	})

	t.Run("one incorrect guess no gameover", func(t *testing.T) {
		iterations := []stateIteration{
			{true, correct, 5, 0, ongoing},
			{true, correct, 10, 0, ongoing},
			{true, correct, 15, 0, ongoing},
			{true, incorrect, 13, 1, ongoing},
			{true, correct, 18, 1, ongoing},
			{true, correct, 23, 1, ongoing},
		}
		state := newSessionState(make(chan bool, 100), difficulties[0])
		runIterations(t, iterations, state)
	})
}

func runIterations(t *testing.T, iterations []stateIteration, state *sessionState) {
	for _, iteration := range iterations {
		var hiddenRune rune
		if iteration.hideRune {
			hiddenRune = state.hideRune()
		}

		if iteration.input != none {
			switch iteration.input {
			case correct:
				state.inputRunePress(hiddenRune)
			case incorrect:
				state.inputRunePress('-')
			}
		}

		if iteration.expectedInvalidKeyPresses != state.invalidKeyPresses {
			t.Errorf("Invalid keypresses %d, expected %d", state.invalidKeyPresses, iteration.expectedInvalidKeyPresses)
		}

		if iteration.expectedScore != state.score {
			t.Errorf("score %d, expected %d", state.score, iteration.expectedScore)
		}
	}
}
