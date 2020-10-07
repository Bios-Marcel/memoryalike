package main

import "time"

type difficulty struct {
	visibleName string

	startDelay time.Duration
	hideTimes  time.Duration

	correctGuessPoints      int
	invalidKeyPressPenality int

	rowCount    int
	columnCount int
	runePools   [][]rune
}

var difficulties = []*difficulty{
	{
		visibleName:        "easy",
		correctGuessPoints: 5,
		//You better take easy seriously!
		invalidKeyPressPenality: 4,
		rowCount:                3,
		columnCount:             2,
		startDelay:              750 * time.Millisecond,
		hideTimes:               1250 * time.Millisecond,
		runePools: [][]rune{
			runeRange('1', '6'),
		},
	}, {
		visibleName:             "normal",
		correctGuessPoints:      5,
		invalidKeyPressPenality: 2,
		rowCount:                3,
		columnCount:             3,
		startDelay:              1500 * time.Millisecond,
		hideTimes:               1250 * time.Millisecond,
		runePools: [][]rune{
			runeRange('0', '9'),
		},
	}, {
		visibleName:             "hard",
		correctGuessPoints:      5,
		invalidKeyPressPenality: 5,
		rowCount:                3,
		columnCount:             3,
		startDelay:              1500 * time.Millisecond,
		hideTimes:               1500 * time.Millisecond,
		runePools: [][]rune{
			runeRange('a', 'z'),
		},
	}, {
		visibleName:             "extreme",
		correctGuessPoints:      4,
		invalidKeyPressPenality: 5,
		rowCount:                4,
		columnCount:             3,
		startDelay:              1500 * time.Millisecond,
		hideTimes:               1500 * time.Millisecond,
		runePools: [][]rune{
			runeRange('0', '9'),
			runeRange('a', 'z'),
		},
	}, {
		visibleName:             "nightmare",
		correctGuessPoints:      4,
		invalidKeyPressPenality: 10,
		rowCount:                5,
		columnCount:             5,
		startDelay:              2500 * time.Millisecond,
		hideTimes:               1500 * time.Millisecond,
		runePools: [][]rune{
			runeRange('0', '9'),
			runeRange('a', 'z'),
		},
	},
}
