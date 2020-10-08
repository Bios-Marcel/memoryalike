package main

import (
	"os"

	"github.com/gdamore/tcell"
)

func main() {
	screen, screenCreationError := createScreen()
	if screenCreationError != nil {
		panic(screenCreationError)
	}

	//Cleans up the terminal buffer and returns it to the shell.
	defer screen.Fini()

	//renderer used for drawing the board and the menu.
	renderer := newRenderer()
	//menuState is reused throughout the runtime of the app. This allows
	//us to remember the selection inbetween sessions.
	menuState := newMenuState()

	//blocks till it's closed.
	openMenu(menuState, screen, renderer)

	renderNotificationChannel := make(chan bool)
	gameSession := newGameSession(renderNotificationChannel, menuState.getDiffculty())
	gameSession.startRuneHidingCoroutine()

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
					oldGameSession := gameSession
					oldGameSession.mutex.Lock()

					//When hitting ESC twice, e.g. when already in the
					//end-screen, we want to go to the menu instead.
					if oldGameSession.state != ongoing {
						openMenu(menuState, screen, renderer)
						//We have to reset the state, as it's still in the
						//"game over" state.
						gameSession = newGameSession(renderNotificationChannel,
							menuState.getDiffculty())
						gameSession.startRuneHidingCoroutine()
					} else {
						oldGameSession.state = gameOver
					}
					oldGameSession.mutex.Unlock()
					renderNotificationChannel <- true
				} else if event.Key() == tcell.KeyCtrlR {
					//RESTART!
					//Remove previous game over message and such and create
					//a fresh state, as we needn't save any information for
					//the next session.
					oldGameSession := gameSession
					oldGameSession.mutex.Lock()

					//Make sure the state knows it's supposed to be dead.
					oldGameSession.state = gameOver
					screen.Clear()
					gameSession = newGameSession(renderNotificationChannel,
						menuState.getDiffculty())
					gameSession.startRuneHidingCoroutine()
					gameSession.mutex.Lock()

					oldGameSession.mutex.Unlock()
					gameSession.mutex.Unlock()
					renderNotificationChannel <- true

				} else if event.Key() == tcell.KeyRune {
					gameSession.mutex.Lock()
					gameSession.inputRunePress(event.Rune())
					gameSession.mutex.Unlock()
				}
			case *tcell.EventResize:
				gameSession.mutex.Lock()
				screen.Clear()
				gameSession.mutex.Unlock()
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
		gameSession.mutex.Lock()
		renderer.drawGameBoard(screen, gameSession)
		gameSession.mutex.Unlock()

		<-renderNotificationChannel
	}
}

// openMenu draws the game menu and listens for keyboard input.
// This method blocks until a difficulty has been selected.
func openMenu(menuState *menuState, targetScreen tcell.Screen, renderer *renderer) {
MENU_KEY_LOOP:
	for {
		//We draw the menu initially and then once after any event.
		renderer.drawMenu(targetScreen, menuState)

		switch event := targetScreen.PollEvent().(type) {
		case *tcell.EventKey:
			if event.Key() == tcell.KeyDown || event.Rune() == 's' || event.Rune() == 'k' {
				if menuState.selectedDifficulty >= len(difficulties)-1 {
					menuState.selectedDifficulty = 0
				} else {
					menuState.selectedDifficulty++
				}
			} else if event.Key() == tcell.KeyUp || event.Rune() == 'w' || event.Rune() == 'j' {
				if menuState.selectedDifficulty <= 0 {
					menuState.selectedDifficulty = len(difficulties) - 1
				} else {
					menuState.selectedDifficulty--
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
