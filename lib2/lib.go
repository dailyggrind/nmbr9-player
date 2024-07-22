package lib2

import (
	"fmt"
	"math"
	"strings"
)

/*

	lib2 uses flat 1D slice instead of 2D slice for perf reasons

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
var NUMBER = [][]int8{
	// each row is a rasterized 4x3
	// NR,NC = 4,3 = 4rows * 3cols
	// (r,c) = (0,1) = 0*NC + 1 = 1
	// (r,c) = (1,0) = 1*NC + 0 = 3
	// (r,c) = (2,2) = 2*NC + 2 = 8
	{0, 0, 0, 0, -1, 0, 0, -1, 0, 0, 0, 0},
	{1, 1, -1, -1, 1, -1, -1, 1, -1, -1, 1, -1},
	{-1, 2, 2, -1, 2, 2, 2, 2, -1, 2, 2, 2},
	{3, 3, 3, -1, -1, 3, -1, 3, 3, 3, 3, 3},
	{-1, 4, 4, -1, 4, -1, 4, 4, 4, -1, 4, 4},
	{5, 5, 5, 5, 5, 5, -1, -1, 5, 5, 5, 5},
	{6, 6, -1, 6, -1, -1, 6, 6, 6, 6, 6, 6},
	{7, 7, 7, -1, 7, -1, 7, 7, -1, 7, -1, -1},
	{-1, 8, 8, -1, 8, 8, 8, 8, -1, 8, 8, -1},
	{9, 9, 9, 9, 9, 9, 9, 9, -1, 9, 9, -1},
}

type Board struct {
	R         int
	C         int
	layers    []*Layer
	flat      *Layer
	seen      []int
	seenLimit int
	fanout    []int
}

type Layer struct {
	R       int
	C       int
	cells   []int8
	BB_TL_R int
	BB_TL_C int
	BB_BR_R int
	BB_BR_C int
	// BB = bounding box of placed tiles
	// BB_TL = top left of bounding box
	// BB_BR = bottom right of bounding box
	// bounding box used to skip comparing outside for perf
}

func NewBoard() *Board {
	return newBoardRC(12, 12, 2)
}

func newBoardRC(rows, cols int, seenLimit int) *Board {
	return &Board{
		R:         rows,
		C:         cols,
		layers:    make([]*Layer, 0, 20),
		flat:      makeFlatRC(rows, cols),
		seen:      make([]int, 10),
		seenLimit: seenLimit,
		fanout:    make([]int, 10),
	}
}

func (board *Board) addLayer() *Layer {
	layer := makeLayerRC(board.R, board.C)
	board.layers = append(board.layers, layer)
	return layer
}

func (board *Board) putNumberAtLayer(level int8, num, row, col int) {
	if int(level) >= len(board.layers) {
		board.addLayer()
	}
	layer := board.layers[level]
	board.putNumber(layer, num, row, col)
	board.flat = board.flatten(board.flat, layer, level)
	board.seen[num]++
}

func (board *Board) setBaseLayer(num int) {
	// put base layer number in the middle since board should be empty
	NR, NC := getNumberSize()
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
		board.fanout = make([]int, steps)
		bestR, bestC, bestLevel, maxScore, err := board.findBestMoveV2(board.flat, board.seen, board.seenLimit, num, steps)
		if err != nil {
			return err, 0
		}
		board.putNumberAtLayer(bestLevel, num, bestR, bestC)
		return nil, maxScore
	}
}

func (board *Board) PrintOverlays(showBB bool) {
	overlays := make([]*Layer, len(board.layers))
	lay := makeLayerRC(board.R, board.C)
	for i, layer := range board.layers {
		board.overlay(lay, layer)
		overlays[i] = copyLayer(lay)
	}
	board.printLayersBB(overlays, showBB)
}

func (board *Board) printFlat() {
	board.printLayerBB(board.flat)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (board *Board) findBestMoveV2(flat *Layer, seen []int, seenLimit int, num int, steps int) (int, int, int8, int, error) {
	board.fanout[len(board.fanout)-steps]++
	maxScore := -10000
	R, C := board.R, board.C
	NR, NC := getNumberSize()
	hasValid := false
	var bestR, bestC int
	var bestLevel int8
	// for every (r,c) check if num can be placed there in a valid way
	delta := 0
	startR := max(0, flat.BB_TL_R-NR) + delta
	startC := max(0, flat.BB_TL_C-NC) + delta
	endR := min(R-1, flat.BB_BR_R+NR) - delta
	endC := min(C-1, flat.BB_BR_C+NC) - delta
	for r := startR; r <= endR; r++ {
		for c := startC; c <= endC; c++ {
			// for r := 0; r < R; r++ {
			// for c := 0; c < C; c++ {
			if valid, level := board.isValid(flat, num, r, c); valid {
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
					layer := makeLayerRC(board.R, board.C)
					board.putNumber(layer, num, r, c)
					seen[num]++
					// compute new flat
					newFlat := copyFlat(flat)
					board.flatten(newFlat, layer, level)
					// recursively find best move
					for s := 0; s < seenLimit; s++ {
						for i := range seen {
							if seen[i] < seenLimit {
								_, _, _, futureScore, err := board.findBestMoveV2(newFlat, seen, seenLimit, i, steps-1)
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

func (board *Board) findBestMove(flat *Layer, num int) (int, int, int8) {
	R, C := board.R, board.C
	maxScore := -1
	var bestR, bestC int
	var bestLevel int8
	for r := 0; r < R; r++ {
		for c := 0; c < C; c++ {
			if valid, level := board.isValid(flat, num, r, c); valid {
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

func copyLayer(layer *Layer) *Layer {
	return copyFlat(layer)
}

func copyFlat(flat *Layer) *Layer {
	cells := make([]int8, len(flat.cells))
	copy(cells, flat.cells)
	return &Layer{
		R:       flat.R,
		C:       flat.C,
		cells:   cells,
		BB_TL_R: flat.BB_TL_R,
		BB_TL_C: flat.BB_TL_C,
		BB_BR_R: flat.BB_BR_R,
		BB_BR_C: flat.BB_BR_C,
	}
}

func score(num int, level int8) int {
	return num * int(level)
}

// check if num can be placed at (row,col) using flat
func (board *Board) isValid(flat *Layer, num int, row, col int) (bool, int8) {
	if !board.isInBounds(row, col) {
		return false, 0
	}
	// validity requires that every non-empty cell needs to be placed
	// on top of the same level number
	same, level := board.isOnSameLevel(flat, num, row, col)
	if !same {
		return false, 0
	}
	// if we're placing at the bottom layer, validity also requires that num is
	// touching an existing already placed number
	if level == 0 && !board.isTouching(flat, num, row, col) {
		return false, 0
	}
	return true, level
}

func (board *Board) isInBounds(row, col int) bool {
	NR, NC := getNumberSize()
	R, C := board.R, board.C
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

// true if num, when placed at (row,col) is touching an existing number in flat
func (board *Board) isTouching(flat *Layer, num int, row, col int) bool {
	n := NUMBER[num]
	NR, NC := getNumberSize()
	R, C := board.R, board.C
	touching := false
outer:
	for r := row; r < row+NR && r < R; r++ {
		for c := col; c < col+NC && c < C; c++ {
			// number cell is not EMPTY
			if n[((r-row)*NC)+(c-col)] != EMPTY {
				// top neighbor
				if r-1 >= 0 && flat.cells[(r-1)*C+c] >= 0 {
					touching = true
					break outer
				}
				// bottom neighbor
				if r+1 < R && flat.cells[(r+1)*C+c] >= 0 {
					touching = true
					break outer
				}
				// left neighbor
				if c-1 >= 0 && flat.cells[r*C+(c-1)] >= 0 {
					touching = true
					break outer
				}
				// right neighbor
				if c+1 < C && flat.cells[r*C+(c+1)] >= 0 {
					touching = true
					break outer
				}
			}
		}
	}
	return touching
}

func (board *Board) isOnSameLevel(flat *Layer, num int, row, col int) (bool, int8) {
	n := NUMBER[num]
	NR, NC := getNumberSize()
	R, C := board.R, board.C
	unset := true
	var fval int8 = -1
	// validity requires that every non-empty cell needs to be placed
	// on top of the same level number
	for r := row; r < row+NR && r < R; r++ {
		for c := col; c < col+NC && c < C; c++ {
			if n[(r-row)*NC+(c-col)] >= 0 { // if num has non-empty cell...
				if unset {
					fval = flat.cells[r*C+c]
					unset = false
				} else {
					if fval != flat.cells[r*C+c] { // if num crosses levels...
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
func (board *Board) flatten(flat, layer *Layer, level int8) *Layer {
	R, C := board.R, board.C
	for r := 0; r < R; r++ {
		for c := 0; c < C; c++ {
			if layer.cells[r*C+c] != EMPTY {
				flat.cells[r*C+c] = level
			}
		}
	}
	// update bounding box of flat
	if layer.BB_TL_R < flat.BB_TL_R {
		flat.BB_TL_R = layer.BB_TL_R
	}
	if layer.BB_TL_C < flat.BB_TL_C {
		flat.BB_TL_C = layer.BB_TL_C
	}
	if layer.BB_BR_R > flat.BB_BR_R {
		flat.BB_BR_R = layer.BB_BR_R
	}
	if layer.BB_BR_C > flat.BB_BR_C {
		flat.BB_BR_C = layer.BB_BR_C
	}
	return flat
}

// overlay layer2 onto layer1, modifying layer1
func (board *Board) overlay(layer1, layer2 *Layer) {
	R, C := board.R, board.C
	for r := 0; r < R; r++ {
		for c := 0; c < C; c++ {
			if layer2.cells[r*C+c] != EMPTY {
				layer1.cells[r*C+c] = layer2.cells[r*C+c]
			}
		}
	}
	// update bounding box of layer
	if layer2.BB_TL_R < layer1.BB_TL_R {
		layer1.BB_TL_R = layer2.BB_TL_R
	}
	if layer2.BB_TL_C < layer1.BB_TL_C {
		layer1.BB_TL_C = layer2.BB_TL_C
	}
	if layer2.BB_BR_R > layer1.BB_BR_R {
		layer1.BB_BR_R = layer2.BB_BR_R
	}
	if layer2.BB_BR_C > layer1.BB_BR_C {
		layer1.BB_BR_C = layer2.BB_BR_C
	}
}

func (board *Board) putNumber(layer *Layer, num, row, col int) {
	n := NUMBER[num]
	NR, NC := getNumberSize()
	for i := row; i < row+NR; i++ {
		for j := col; j < col+NC; j++ {
			nn := n[(i-row)*NC+(j-col)]
			if nn != EMPTY {
				layer.cells[i*board.C+j] = nn
			}
		}
	}
	// update bounding box of layer
	if row < layer.BB_TL_R {
		layer.BB_TL_R = row
	}
	if col < layer.BB_TL_C {
		layer.BB_TL_C = col
	}
	brr := row + NR - 1
	if brr > layer.BB_BR_R {
		layer.BB_BR_R = brr
	}
	brc := col + NC - 1
	if brc > layer.BB_BR_C {
		layer.BB_BR_C = brc
	}
}

func makeFlatRC(rows, cols int) *Layer {
	return makeLayerRC(rows, cols)
}

func makeLayerRC(rows, cols int) *Layer {
	cells := make([]int8, rows*cols)
	for i := range cells {
		cells[i] = EMPTY
	}
	return &Layer{
		R:       rows,
		C:       cols,
		cells:   cells,
		BB_TL_R: math.MaxInt,
		BB_TL_C: math.MaxInt,
		BB_BR_R: math.MinInt,
		BB_BR_C: math.MinInt,
	}
}

func getNumberSize() (int, int) {
	return 4, 3 // hardcoded
}

func (board *Board) printLayer(layer *Layer) {
	board.printLayers([]*Layer{layer})
}

func (board *Board) printLayerBB(layer *Layer) {
	board.printLayersBB([]*Layer{layer}, true)
}

func (board *Board) printLayers(layers []*Layer) {
	board.printLayersBB(layers, false)
}

func (board *Board) printLayersBB(layers []*Layer, showBB bool) {
	if len(layers) == 0 {
		fmt.Println("==========")
		return
	}
	R, C := board.R, board.C
	fmt.Println(strings.Repeat("=", (C+3)*len(layers)))
	for r := 0; r < R; r++ {
		for _, layer := range layers {
			for c := 0; c < C; c++ {
				if showBB && // show 4 corners of the bounding box
					((r == layer.BB_TL_R && c == layer.BB_TL_C) || (r == layer.BB_BR_R && c == layer.BB_BR_C) ||
						(r == layer.BB_TL_R && c == layer.BB_BR_C) || (r == layer.BB_BR_R && c == layer.BB_TL_C)) {
					fmt.Printf(White + "X" + Reset)
				} else {
					if layer.cells[r*C+c] == EMPTY {
						fmt.Printf(".")
					} else {
						fmt.Printf(color(layer.cells[r*C+c], "%v"), layer.cells[r*C+c])
					}
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
