package main

import (
	"os"

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
	renderNotificationChannel := make(chan bool)
	currentSessionState := newSessionState(renderNotificationChannel, width, height, difficulty)
	renderer := newRenderer()

	go func() {
		for {
			switch event := screen.PollEvent().(type) {
			case *tcell.EventKey:
				if event.Key() == tcell.KeyCtrlC {
					screen.Fini()
					os.Exit(0)
				} else if event.Key() == tcell.KeyEscape {
					//SURRENDER!
					currentSessionState.mutex.Lock()
					currentSessionState.currentGameState = gameOver
					currentSessionState.mutex.Unlock()
					renderNotificationChannel <- true
				} else if event.Key() == tcell.KeyCtrlR {
					//RESTART!
					//Remove previous game over message and such and create
					//a fresh state, as we needn't save any information for
					//the next session.
					oldSession := currentSessionState
					oldSession.mutex.Lock()
					screen.Clear()
					currentSessionState = newSessionState(renderNotificationChannel, width, height, difficulty)
					oldSession.mutex.Unlock()
				} else if event.Key() == tcell.KeyRune {
					currentSessionState.mutex.Lock()
					currentSessionState.applyKeyEvent(event)
					currentSessionState.mutex.Unlock()
				}
			case *tcell.EventResize:
				currentSessionState.mutex.Lock()
				screen.Clear()
				currentSessionState.mutex.Unlock()
				renderNotificationChannel <- true
				//TODO Handle resize; Validate session;
			default:
				//Unsupported or irrelevant event
			}
		}
	}()

	//Gameloop; We draw whenever there's a frame-change. This means we
	//don't have any specific frame-rates and it could technically happen
	//that we don't draw for a while. The first frame is drawn without
	//waiting for a change, so that the screen doesn't stay empty.

	for {
		//We start lock before draw in order to avoid drawing crap.
		currentSessionState.mutex.Lock()
		renderer.draw(screen, currentSessionState)
		currentSessionState.mutex.Unlock()

		<-renderNotificationChannel
	}
}
