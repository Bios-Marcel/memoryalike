package main

import (
	"os"
	"sync"

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

func main() {
	screen, screenCreationError := createScreen()
	if screenCreationError != nil {
		panic(screenCreationError)
	}

	//Cleans up the terminal buffer and returns it to the shell.
	defer screen.Fini()

	width, height := screen.Size()
	currentSessionState := newSessionState(width, height, difficulty)
	renderer := newRenderer()

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
				} else if event.Key() == tcell.KeyEscape {
					//SURRENDER!
					currentSessionState.currentGameState = gameOver
				} else if event.Key() == tcell.KeyCtrlR {
					//RESTART!

					//We lock everything in order to avoid any clashing expectations.
					keyEventMutex.Lock()
					oldSession := currentSessionState
					oldSession.mutex.Lock()

					//Make sure there's no invalid key events in the
					//queue to avoid faulty point loss.
					keyEvents = keyEvents[:0]
					//Remove previous game over message and such.
					screen.Clear()
					currentSessionState = newSessionState(width, height, difficulty)

					//Unlock old session to resume execution in gameloop.
					oldSession.mutex.Unlock()
					keyEventMutex.Unlock()
				} else if event.Key() == tcell.KeyRune {
					keyEvents = append(keyEvents, event)
				}
				keyEventMutex.Unlock()
			case *tcell.EventResize:
				//TODO Handle resize; Validate session;
			default:
				//Unsupported or irrelevant event
			}
		}
	}()

	//Gameloop; We always draw and then check for buffered key-inputs.
	//We do the buffering in order to be able to constantly listen for
	//new keysstrokes. This should avoid lag and such.

	for {
		//We start lock before draw in order to avoid drawing crap.
		currentSessionState.mutex.Lock()

		renderer.draw(screen, currentSessionState)

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
