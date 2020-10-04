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

// getDiffculty returns the diffculty chosen by the user.
func (menuState *menuState) getDiffculty() *difficulty {
	return difficulties[menuState.selectedDifficulty]
}
