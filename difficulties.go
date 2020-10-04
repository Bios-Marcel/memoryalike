package main

import "time"

type difficulty struct {
	visibleName string

	startDelay time.Duration
	hideTimes  time.Duration

	rowCount    int
	columnCount int
	runePools   [][]rune
}

var difficulties = []*difficulty{
	{
		visibleName: "easy",
		rowCount:    3,
		columnCount: 2,
		startDelay:  750 * time.Millisecond,
		hideTimes:   1250 * time.Millisecond,
		runePools: [][]rune{
			runeRange('1', '6'),
		},
	}, {
		visibleName: "normal",
		rowCount:    3,
		columnCount: 3,
		startDelay:  1500 * time.Millisecond,
		hideTimes:   1250 * time.Millisecond,
		runePools: [][]rune{
			runeRange('0', '9'),
		},
	}, {
		visibleName: "hard",
		rowCount:    3,
		columnCount: 3,
		startDelay:  1500 * time.Millisecond,
		hideTimes:   1500 * time.Millisecond,
		runePools: [][]rune{
			runeRange('a', 'z'),
		},
	}, {
		visibleName: "extreme",
		rowCount:    4,
		columnCount: 3,
		startDelay:  1500 * time.Millisecond,
		hideTimes:   1500 * time.Millisecond,
		runePools: [][]rune{
			runeRange('0', '9'),
			runeRange('a', 'z'),
		},
	}, {
		visibleName: "nightmare",
		rowCount:    5,
		columnCount: 5,
		startDelay:  2500 * time.Millisecond,
		hideTimes:   1500 * time.Millisecond,
		runePools: [][]rune{
			runeRange('0', '9'),
			runeRange('a', 'z'),
		},
	},

	//Easter egg level?!
	//{'!', '"', '§', '$', '%', '&', '/', '(', ')', '0', 'ß'},
}
