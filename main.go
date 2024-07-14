package main

import (
	"fmt"
	"nmbr9/lib"
)

/*

	board = all layers + a flat
	layer = 2D layer where each cell value equals the number placed there
			empty cells have a value of -1 and are printed as '.'
			.....	.....
			..999	..119
			..99. > ..91.
			..99.	..91.
			.....	.....
	cell = 	each individual cell in a layer, indexed by (row,col)
	level = index of layer in the board, 0 based
	number = game piece shaped like a number
	flat = 	a single layer representation of all layers in the board flattened to 1 layer
			where each cell value equals the level that cell is placed at
			.....	.....
			..000	..110
			..00. > ..01.
			..00.	..01.
			.....	.....

	notes:
		* value of number is needed during scoring

*/

func main() {
	// input := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	board := lib.NewBoard()
	var num int
	for {
		fmt.Print("enter a number: ")
		_, err := fmt.Scanf("%d\n", &num)
		if err != nil {
			fmt.Printf("error scanning input: %v\n", err)
			break
		}
		err, _ = board.ApplyBestMove(num, 4)
		if err != nil {
			fmt.Printf("error applying best move: %v\n", err)
			break
		} else {
			board.PrintOverlays()
			// fmt.Printf("best move score: %v\n", score)
		}
	}
}
