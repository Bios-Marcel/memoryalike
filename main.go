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

var currentMenuState = &menuState{}

func main() {
	screen, screenCreationError := createScreen()
	if screenCreationError != nil {
		panic(screenCreationError)
	}

	//Cleans up the terminal buffer and returns it to the shell.
	defer screen.Fini()

	//renderer used for drawing the board and the menu.
	renderer := newRenderer()

	//Draw menu, in order to allow difficulty selection.
	width, height := screen.Size()

MENU_KEY_LOOP:
	for {
		//We draw the menu initially and then once after any event.
		renderer.drawMenu(screen, currentMenuState)

		switch event := screen.PollEvent().(type) {
		case *tcell.EventKey:
			if event.Key() == tcell.KeyDown {
				if currentMenuState.selectedDifficulty >= 3 {
					currentMenuState.selectedDifficulty = 0
				} else {
					currentMenuState.selectedDifficulty++
				}
			} else if event.Key() == tcell.KeyUp {
				if currentMenuState.selectedDifficulty <= 0 {
					currentMenuState.selectedDifficulty = 3
				} else {
					currentMenuState.selectedDifficulty--
				}
			} else if event.Key() == tcell.KeyEnter {
				//We clear in order to get rid of the menu for sure.
				screen.Clear()
				break MENU_KEY_LOOP
				//Implicitly proceed.
			} else if event.Key() == tcell.KeyCtrlC {
				screen.Fini()
				os.Exit(0)
			}
		default:
			//Unsupported or irrelevant event
		}
	}

	renderNotificationChannel := make(chan bool)
	currentSessionState := newSessionState(renderNotificationChannel, width, height, currentMenuState.selectedDifficulty)

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
					currentSessionState = newSessionState(renderNotificationChannel, width, height, currentMenuState.selectedDifficulty)
					currentSessionState.mutex.Lock()
					oldSession.mutex.Unlock()
					currentSessionState.mutex.Unlock()
					renderNotificationChannel <- true

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
		renderer.drawGameBoard(screen, currentSessionState)
		currentSessionState.mutex.Unlock()

		<-renderNotificationChannel
	}
}
