# Memoryalike

Just a tiny terminal game where you have to remember the characters presented
to you and hit the respective key when they are hidden by a box. It's kinda
like memory.

## Rules

The rules are quite simply. Your goal is to guess all characters correctly.
If you can't do that, you don't win. If 40% of the board is hidden, you lose.
So speed does matter. Every incorrect guess will give you minus points.
While achieving a victory might not be easy, you can still get a good loss.

## Controls

You can give up on <kbd>ESC</kbd> and restart on <kbd>Ctrl</kbd> + <kbd>R</kbd>.

If you hit <kbd>ESC</kbd> again while still in the "Game Over" / "Victory"
screen, you'll be taken to the main menu.

## How to use it

It's not fully done yet, so it's the bare minimum required for playing.

You need to download Golang 1.14 or later and either create an executable
with `go build .` or run it directly via `go run .`.

## What's up with the name

That's as far as my imagination goes. If you have suggestions for a better
name, feel free to hit me up.

## Known issues

* Resizing isn't really handled yet
