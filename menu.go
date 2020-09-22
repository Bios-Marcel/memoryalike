package main

type menuState struct {
	selectedDifficulty int
}

func newMenuState() *menuState {
	return &menuState{
		//Default difficulty normal
		selectedDifficulty: 1,
	}
}
