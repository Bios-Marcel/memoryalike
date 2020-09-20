package main

import "flag"

var difficulty int

func init() {
	flag.IntVar(&difficulty, "difficulty", 0, "Set the difficulty. Must be a value between 0 and 3.")
	flag.Parse()
}
