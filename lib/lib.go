package lib

import (
	"fmt"
	"strings"
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

var EMPTY int8 = -1
var NUMBER = [][][]int8{
	{
		{0, 0, 0},
		{0, -1, 0},
		{0, -1, 0},
		{0, 0, 0},
	},
	{
		{1, 1, -1},
		{-1, 1, -1},
		{-1, 1, -1},
		{-1, 1, -1},
	},
	{
		{-1, 2, 2},
		{-1, 2, 2},
		{2, 2, -1},
		{2, 2, 2},
	},
	{
		{3, 3, 3},
		{-1, -1, 3},
		{-1, 3, 3},
		{3, 3, 3},
	},
	{
		{-1, 4, 4},
		{-1, 4, -1},
		{4, 4, 4},
		{-1, 4, 4},
	},
	{
		{5, 5, 5},
		{5, 5, 5},
		{-1, -1, 5},
		{5, 5, 5},
	},
	{
		{6, 6, -1},
		{6, -1, -1},
		{6, 6, 6},
		{6, 6, 6},
	},
	{
		{7, 7, 7},
		{-1, 7, -1},
		{7, 7, -1},
		{7, -1, -1},
	},
	{
		{-1, 8, 8},
		{-1, 8, 8},
		{8, 8, -1},
		{8, 8, -1},
	},
	{
		{9, 9, 9},
		{9, 9, 9},
		{9, 9, -1},
		{9, 9, -1},
	},
}

type Board struct {
	R         int
	C         int
	layers    [][][]int8
	flat      [][]int8
	seen      []int
	seenLimit int
}

func NewBoard() *Board {
	return newBoardRC(12, 12, 2)
}

func newBoardRC(rows, cols int, seenLimit int) *Board {
	return &Board{
		R:         rows,
		C:         cols,
		layers:    make([][][]int8, 0, 10),
		flat:      makeFlatRC(rows, cols),
		seen:      make([]int, 10),
		seenLimit: seenLimit,
	}
}

func (board *Board) addLayer() {
	board.layers = append(board.layers, makeLayerRC(board.R, board.C))
}

func (board *Board) putNumberAtLayer(level int8, num, row, col int) {
	if int(level) >= len(board.layers) {
		board.addLayer()
	}
	layer := board.layers[level]
	putNumber(layer, num, row, col)
	board.flat = flatten(board.flat, layer, level)
	board.seen[num]++
}

func (board *Board) setBaseLayer(num int) {
	// put base layer number in the middle since board should be empty
	NR, NC := getNumberSize(NUMBER[num])
	midR := (board.R - NR) / 2
	midC := (board.C - NC) / 2
	board.putNumberAtLayer(0, num, midR, midC)
}

// steps: how many steps to look ahead
// limit: how many steps until game ends
func (board *Board) ApplyBestMove(num int, steps int) (error, int) {
	if len(board.layers) == 0 {
		board.setBaseLayer(num)
		return nil, 0
	} else {
		if board.seen[num] >= board.seenLimit {
			return fmt.Errorf("num:%v has already been seen limit:%v times", num, board.seenLimit), 0
		}
		bestR, bestC, bestLevel, maxScore, err := findBestMoveV2(board.flat, board.seen, board.seenLimit, num, steps)
		if err != nil {
			return err, 0
		}
		board.putNumberAtLayer(bestLevel, num, bestR, bestC)
		return nil, maxScore
	}
}

func (board *Board) PrintOverlays() {
	overlays := make([][][]int8, len(board.layers))
	lay := makeLayerRC(board.R, board.C)
	for i, layer := range board.layers {
		overlay(lay, layer)
		overlays[i] = copyLayer(lay)
	}
	printLayers(overlays)
}

func (board *Board) printFlat() {
	printLayer(board.flat)
}

func findBestMoveV2(flat [][]int8, seen []int, seenLimit int, num int, steps int) (int, int, int8, int, error) {
	maxScore := -10000
	R, C := getLayerSize(flat)
	hasValid := false
	var bestR, bestC int
	var bestLevel int8
	for r := 0; r < R; r++ {
		for c := 0; c < C; c++ {
			if valid, level := isValid(flat, num, r, c); valid {
				hasValid = true
				// score this move
				newScore := score(num, level)
				if newScore > maxScore {
					maxScore = newScore
					bestR, bestC = r, c
					bestLevel = level
				}
				if steps > 1 {
					// apply move
					layer := makeLayerRC(R, C)
					putNumber(layer, num, r, c)
					seen[num]++
					// compute new flat
					newFlat := copyFlat(flat)
					flatten(newFlat, layer, level)
					// recursively find best move
					for s := 0; s < seenLimit; s++ {
						for i := range seen {
							if seen[i] < seenLimit {
								_, _, _, futureScore, err := findBestMoveV2(newFlat, seen, seenLimit, i, steps-1)
								if err == nil && newScore+futureScore > maxScore {
									maxScore = newScore + futureScore
									bestR, bestC = r, c
									bestLevel = level
								}
							}
						}
					}
					// undo move to backtrack
					seen[num]--
				}
			}
		}
	}
	if !hasValid {
		return 0, 0, 0, 0, fmt.Errorf("no valid moves for num: %v", num)
	}
	return bestR, bestC, bestLevel, maxScore, nil
}

func findBestMove(flat [][]int8, num int) (int, int, int8) {
	R, C := getLayerSize(flat)
	maxScore := -1
	var bestR, bestC int
	var bestLevel int8
	for r := 0; r < R; r++ {
		for c := 0; c < C; c++ {
			if valid, level := isValid(flat, num, r, c); valid {
				newScore := score(num, level)
				if newScore > maxScore {
					maxScore = newScore
					bestR, bestC = r, c
					bestLevel = level
				}
			}
		}
	}
	return bestR, bestC, bestLevel
}

func copyLayer(layer [][]int8) [][]int8 {
	return copyFlat(layer)
}

func copyFlat(flat [][]int8) [][]int8 {
	flat2 := make([][]int8, len(flat))
	for i := range flat {
		flat2[i] = make([]int8, len(flat[i]))
		copy(flat2[i], flat[i])
	}
	return flat2
}

func score(num int, level int8) int {
	return num * int(level)
}

// check if num can be placed at (row,col) using flat
func isValid(flat [][]int8, num int, row, col int) (bool, int8) {
	if !isInBounds(flat, num, row, col) {
		return false, 0
	}
	// validity requires that every non-empty cell needs to be placed
	// on top of the same level number
	same, level := isOnSameLevel(flat, num, row, col)
	if !same {
		return false, 0
	}
	// if we're placing at the bottom layer, validity also requires that num is
	// touching an existing already placed number
	if level == 0 && !isTouching(flat, num, row, col) {
		return false, 0
	}
	return true, level
}

func isInBounds(flat [][]int8, num int, row, col int) bool {
	n := NUMBER[num]
	NR, NC := getNumberSize(n)
	R, C := getLayerSize(flat)
	/*
		000
		000 (row,col) = (1,1)
		000 (R,C) = (3,3)

		11
		11 (NR,NC) = (2,2)

		000
		011 return 1+2-1<3 && 1+2-1<3 = true
		011
	*/
	return row+NR-1 < R && col+NC-1 < C
}

func isTouching(flat [][]int8, num int, row, col int) bool {
	n := NUMBER[num]
	NR, NC := getNumberSize(n)
	touching := false
outer:
	for r := row; r < row+NR && r < len(flat); r++ {
		for c := col; c < col+NC && c < len(flat[r]); c++ {
			// number cell is not EMPTY
			if n[r-row][c-col] != EMPTY {
				// top neighbor
				if r-1 >= 0 && flat[r-1][c] >= 0 {
					touching = true
					break outer
				}
				// bottom neighbor
				if r+1 < len(flat) && flat[r+1][c] >= 0 {
					touching = true
					break outer
				}
				// left neighbor
				if c-1 >= 0 && flat[r][c-1] >= 0 {
					touching = true
					break outer
				}
				// right neighbor
				if c+1 < len(flat[r]) && flat[r][c+1] >= 0 {
					touching = true
					break outer
				}
			}
		}
	}
	return touching
}

func isOnSameLevel(flat [][]int8, num int, row, col int) (bool, int8) {
	n := NUMBER[num]
	NR, NC := getNumberSize(n)
	unset := true
	var fval int8 = -1
	// validity requires that every non-empty cell needs to be placed
	// on top of the same level number
	for r := row; r < row+NR && r < len(flat); r++ {
		for c := col; c < col+NC && c < len(flat[r]); c++ {
			if n[r-row][c-col] >= 0 { // if num has non-empty cell...
				if unset {
					fval = flat[r][c]
					unset = false
				} else {
					if fval != flat[r][c] { // if num crosses levels...
						return false, 0
					}
				}
			}
		}
	}
	level := fval + 1
	return true, level
}

// merges layer onto flat, modifying flat
// level is the level of layer
func flatten(flat, layer [][]int8, level int8) [][]int8 {
	R, C := getLayerSize(flat)
	for r := 0; r < R; r++ {
		for c := 0; c < C; c++ {
			if layer[r][c] != EMPTY {
				flat[r][c] = level
			}
		}
	}
	return flat
}

// overlay layer2 onto layer1, modifying layer1
func overlay(layer1, layer2 [][]int8) {
	R, C := getLayerSize(layer1)
	for r := 0; r < R; r++ {
		for c := 0; c < C; c++ {
			if layer2[r][c] != EMPTY {
				layer1[r][c] = layer2[r][c]
			}
		}
	}
}

func putNumber(layer [][]int8, num, row, col int) {
	n := NUMBER[num]
	NR, NC := getNumberSize(n)
	for i := row; i < row+NR; i++ {
		for j := col; j < col+NC; j++ {
			if n[i-row][j-col] != EMPTY {
				layer[i][j] = n[i-row][j-col]
			}
		}
	}
}

func makeFlatRC(rows, cols int) [][]int8 {
	return makeLayerRC(rows, cols)
}

func makeLayerRC(rows, cols int) [][]int8 {
	layer := make([][]int8, rows)
	for i := range layer {
		layer[i] = make([]int8, cols)
		for j := range layer[i] {
			layer[i][j] = -1
		}
	}
	return layer
}

func getNumberSize(number [][]int8) (int, int) {
	return getLayerSize(number)
}

func getLayerSize(layer [][]int8) (int, int) {
	rows := len(layer)
	cols := 0
	if rows > 0 {
		cols = len(layer[0])
	}
	return rows, cols
}

func printLayer(layer [][]int8) {
	printLayers([][][]int8{layer})
}

func printLayers(layers [][][]int8) {
	if len(layers) == 0 {
		fmt.Println("==========")
		return
	}
	R, C := getLayerSize(layers[0])
	fmt.Println(strings.Repeat("=", (C+3)*len(layers)))
	for r := 0; r < R; r++ {
		for _, layer := range layers {
			for c := 0; c < C; c++ {
				if layer[r][c] == EMPTY {
					fmt.Printf(".")
				} else {
					fmt.Printf(color(layer[r][c], "%v"), layer[r][c])
				}
			}
			if r == R/2 {
				fmt.Print(" > ")
			} else {
				fmt.Print("   ")
			}
		}
		fmt.Println()
	}
}

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var Purple = "\033[90m"
var White = "\033[97m"
var Orange = "\033[91m"
var Pink = "\033[35m"

// var Brown = "\033[94m"

var COLOR = []string{
	Gray,
	Gray,
	Orange,
	Yellow,
	Green,
	Cyan,
	Blue,
	Purple,
	Orange,
	Red,
}

func color(num int8, str string) string {
	return COLOR[num] + str + Reset
}
