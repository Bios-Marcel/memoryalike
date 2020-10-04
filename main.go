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

// currentMenuState is global in order to remember the menu state between
// sessions. It's on purpose, not by accident.
var currentMenuState = newMenuState()

func main() {
	screen, screenCreationError := createScreen()
	if screenCreationError != nil {
		panic(screenCreationError)
	}

	//Cleans up the terminal buffer and returns it to the shell.
	defer screen.Fini()

	//renderer used for drawing the board and the menu.
	renderer := newRenderer()

	//blocks till it's closed.
	openMenu(screen, renderer)

	renderNotificationChannel := make(chan bool)
	currentSessionState := newSessionState(renderNotificationChannel, currentMenuState.getDiffculty())

	//Listen for key input on the gameboard.
	go func() {
		for {
			switch event := screen.PollEvent().(type) {
			case *tcell.EventKey:
				if event.Key() == tcell.KeyCtrlC {
					screen.Fini()
					os.Exit(0)
				} else if event.Key() == tcell.KeyEscape {
					//SURRENDER!
					oldSession := currentSessionState
					oldSession.mutex.Lock()

					//When hitting ESC twice, e.g. when already in the
					//end-screen, we want to go to the menu instead.
					if oldSession.currentGameState != ongoing {
						openMenu(screen, renderer)
						//We have to reset the state, as it's still in the
						//"game over" state.
						currentSessionState = newSessionState(renderNotificationChannel,
							currentMenuState.getDiffculty())
					} else {
						oldSession.currentGameState = gameOver
					}
					oldSession.mutex.Unlock()
					renderNotificationChannel <- true
				} else if event.Key() == tcell.KeyCtrlR {
					//RESTART!
					//Remove previous game over message and such and create
					//a fresh state, as we needn't save any information for
					//the next session.
					oldSession := currentSessionState
					oldSession.mutex.Lock()
					screen.Clear()
					currentSessionState = newSessionState(renderNotificationChannel,
						currentMenuState.getDiffculty())
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

// openMenu draws the game menu and listens for keyboard input.
// This method blocks until a difficulty has been selected.
func openMenu(targetScreen tcell.Screen, renderer *renderer) {
MENU_KEY_LOOP:
	for {
		//We draw the menu initially and then once after any event.
		renderer.drawMenu(targetScreen, currentMenuState)

		switch event := targetScreen.PollEvent().(type) {
		case *tcell.EventKey:
			if event.Key() == tcell.KeyDown || event.Rune() == 's' || event.Rune() == 'k' {
				if currentMenuState.selectedDifficulty >= len(difficulties)-1 {
					currentMenuState.selectedDifficulty = 0
				} else {
					currentMenuState.selectedDifficulty++
				}
			} else if event.Key() == tcell.KeyUp || event.Rune() == 'w' || event.Rune() == 'j' {
				if currentMenuState.selectedDifficulty <= 0 {
					currentMenuState.selectedDifficulty = len(difficulties) - 1
				} else {
					currentMenuState.selectedDifficulty--
				}
			} else if event.Key() == tcell.KeyEnter {
				//We clear in order to get rid of the menu for sure.
				targetScreen.Clear()
				break MENU_KEY_LOOP
				//Implicitly proceed.
			} else if event.Key() == tcell.KeyCtrlC {
				targetScreen.Fini()
				os.Exit(0)
			}
		default:
			//Unsupported or irrelevant event
		}
	}
}
