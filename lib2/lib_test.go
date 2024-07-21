package lib2

import (
	"fmt"
	"reflect"
	"testing"
)

func TestApplyBestMove(t *testing.T) {
	input := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	board := NewBoard()
	steps := 3
	for _, n := range input {
		err, _ := board.ApplyBestMove(n, steps)
		if err != nil {
			fmt.Printf("error applying best move: %v\n", err)
		} else {
			board.PrintOverlays(true)
			// fmt.Printf("best move score: %v\n", score)
		}
	}
}

func TestApplyBestMove_Speed(t *testing.T) {
	type Move struct {
		num      int
		maxScore int
	}
	tests := []struct {
		moves []Move
	}{
		/*
			move:     move:
			.882244   .882664
			.88224.   .88264.
			8822444 > 8822666 >
			8822244   8822666
		*/
		{
			moves: []Move{{num: 2, maxScore: 0}, {num: 4, maxScore: 9}, {num: 8, maxScore: 9}, {num: 6, maxScore: 15}},
		},
	}
	steps := 5
	seenLimit := 1
	for _, tt := range tests {
		R, C := 10, 10
		board := newBoardRC(R, C, seenLimit)
		for _, move := range tt.moves {
			err, maxScore := board.ApplyBestMove(move.num, steps)
			board.PrintOverlays(true)
			if err != nil {
				t.Fatalf("err applying best move: %v", err)
			}
			fmt.Printf("maxScore: %v\n", maxScore)
			// board.printFlat()
			// if maxScore != move.maxScore {
			// 	t.Fatalf("maxScore: want:%v != got:%v", move.maxScore, maxScore)
			// }
		}
	}
}

func TestApplyBestMove_2Steps(t *testing.T) {
	type Move struct {
		num      int
		maxScore int
	}
	tests := []struct {
		moves []Move
	}{
		/*
			move:     move:
			.882244   .882664
			.88224.   .88264.
			8822444 > 8822666 >
			8822244   8822666
		*/
		{
			moves: []Move{{num: 2, maxScore: 0}, {num: 4, maxScore: 9}, {num: 8, maxScore: 9}, {num: 6, maxScore: 15}},
		},
	}
	steps := 2
	seenLimit := 1
	for _, tt := range tests {
		R, C := 4, 7
		board := newBoardRC(R, C, seenLimit)
		for _, move := range tt.moves {
			err, maxScore := board.ApplyBestMove(move.num, steps)
			board.PrintOverlays(true)
			if err != nil {
				t.Fatalf("err applying best move: %v", err)
			}
			if maxScore != move.maxScore {
				t.Fatalf("maxScore: want:%v != got:%v", move.maxScore, maxScore)
			}
		}
	}
}

func TestApplyBestMove_1Step(t *testing.T) {
	type Move struct {
		num      int
		maxScore int
	}
	tests := []struct {
		moves []Move
	}{
		/*
			setup:    next:     next:
			...2244   ...9994   ...9884
			...224.   ...999.   ...988.
			..22444 > ..29944 > ..28844
			..22244   ..29944   ..28844
		*/
		{
			moves: []Move{{num: 2, maxScore: 0}, {num: 4, maxScore: 0}, {num: 9, maxScore: 9}, {num: 8, maxScore: 16}},
		},
	}
	steps := 1
	seenLimit := 1
	for _, tt := range tests {
		R, C := 4, 7
		board := newBoardRC(R, C, seenLimit)
		for _, move := range tt.moves {
			err, maxScore := board.ApplyBestMove(move.num, steps)
			// board.PrintOverlays()
			if err != nil {
				t.Fatalf("err applying best move: %v", err)
			}
			if maxScore != move.maxScore {
				t.Fatalf("maxScore: want:%v != got:%v", move.maxScore, maxScore)
			}
		}
	}
}

func TestFindBestMoveV2_1Setup2Best(t *testing.T) {
	tests := []struct {
		setupNum       int
		setupMove      [2]int
		nextNum0       int
		wantNextMove0  [2]int
		wantNextLevel0 int8
		nextNum1       int
		wantNextMove1  [2]int
		wantNextLevel1 int8
	}{
		/*
			setup:    next:     next:
			.999...   .988...   .98866.
			.999...   .988...   .9886..
			.99....   .88....   .88.666
			.99....   .88....   .88.666
		*/
		{
			setupNum: 9, setupMove: [2]int{0, 1},
			nextNum0: 8, wantNextMove0: [2]int{0, 1}, wantNextLevel0: 1,
			nextNum1: 6, wantNextMove1: [2]int{0, 4}, wantNextLevel1: 0,
		},
		/*
			setup:    next:     next:
			..22...   ..2244.   ..6644.
			..22...   ..224..   ..624..
			.22....   .22444.   .26664.
			.222...   .22244.   .26664.
		*/
		{
			setupNum: 2, setupMove: [2]int{0, 1},
			nextNum0: 4, wantNextMove0: [2]int{0, 3}, wantNextLevel0: 0,
			nextNum1: 6, wantNextMove1: [2]int{0, 2}, wantNextLevel1: 1,
		},
	}
	seenLimit := 1
	for _, tt := range tests {
		R, C := 4, 7
		board := newBoardRC(R, C, seenLimit)
		board.addLayer()
		board.putNumberAtLayer(0, tt.setupNum, tt.setupMove[0], tt.setupMove[1])
		steps := 1
		seenLimit := 1
		// find best move 0
		seen := make([]int, 10)
		br, bc, level, _, err := board.findBestMoveV2(board.flat, seen, seenLimit, tt.nextNum0, steps)
		if err != nil {
			t.Fatalf("err finding best move: %v", err)
		}
		if best, want := [2]int{br, bc}, tt.wantNextMove0; best != want {
			t.Fatalf("best move: want:%v != got:%v", want, best)
		}
		if best, want := level, tt.wantNextLevel0; best != want {
			t.Fatalf("best level: want:%v != got:%v", want, best)
		}
		// apply best move 0
		board.putNumberAtLayer(level, tt.nextNum0, br, bc)
		// find best move 1
		seen = make([]int, 10)
		br, bc, level, _, err = board.findBestMoveV2(board.flat, seen, seenLimit, tt.nextNum1, steps)
		if err != nil {
			t.Fatalf("err finding best move: %v", err)
		}
		if best, want := [2]int{br, bc}, tt.wantNextMove1; best != want {
			t.Fatalf("best move: want:%v != got:%v", want, best)
		}
		if best, want := level, tt.wantNextLevel1; best != want {
			t.Fatalf("best level: want:%v != got:%v", want, best)
		}
	}
}

func TestFindBestMove1Setup2Best(t *testing.T) {
	tests := []struct {
		setupNum       int
		setupMove      [2]int
		nextNum0       int
		wantNextMove0  [2]int
		wantNextLevel0 int8
		nextNum1       int
		wantNextMove1  [2]int
		wantNextLevel1 int8
	}{
		/*
			setup:    next:     next:
			.999...   .988...   .98866.
			.999...   .988...   .9886..
			.99....   .88....   .88.666
			.99....   .88....   .88.666
		*/
		{
			setupNum: 9, setupMove: [2]int{0, 1},
			nextNum0: 8, wantNextMove0: [2]int{0, 1}, wantNextLevel0: 1,
			nextNum1: 6, wantNextMove1: [2]int{0, 4}, wantNextLevel1: 0,
		},
		/*
			setup:    next:     next:
			..22...   ..2244.   ..6644.
			..22...   ..224..   ..624..
			.22....   .22444.   .26664.
			.222...   .22244.   .26664.
		*/
		{
			setupNum: 2, setupMove: [2]int{0, 1},
			nextNum0: 4, wantNextMove0: [2]int{0, 3}, wantNextLevel0: 0,
			nextNum1: 6, wantNextMove1: [2]int{0, 2}, wantNextLevel1: 1,
		},
	}
	seenLimit := 1
	for _, tt := range tests {
		R, C := 4, 7
		board := newBoardRC(R, C, seenLimit)
		board.addLayer()
		board.putNumberAtLayer(0, tt.setupNum, tt.setupMove[0], tt.setupMove[1])
		// find best move 0
		br, bc, level := board.findBestMove(board.flat, tt.nextNum0)
		if best, want := [2]int{br, bc}, tt.wantNextMove0; best != want {
			t.Fatalf("best move: want:%v != got:%v", want, best)
		}
		if best, want := level, tt.wantNextLevel0; best != want {
			t.Fatalf("best level: want:%v != got:%v", want, best)
		}
		// apply best move 0
		board.putNumberAtLayer(level, tt.nextNum0, br, bc)
		// find best move 1
		br, bc, level = board.findBestMove(board.flat, tt.nextNum1)
		if best, want := [2]int{br, bc}, tt.wantNextMove1; best != want {
			t.Fatalf("best move: want:%v != got:%v", want, best)
		}
		if best, want := level, tt.wantNextLevel1; best != want {
			t.Fatalf("best level: want:%v != got:%v", want, best)
		}
	}
}

func TestFindBestMove2Setups1Layer(t *testing.T) {
	tests := []struct {
		setupNum   int
		setupMove  [2]int
		setupNum1  int
		setupMove1 [2]int
		nextNum    int
		nextMove   [2]int
	}{
		/*
			setup:    next:     next:	  best: 0,3
			.999...   .99988.   .66988.
			.999...   .99988.   .69988.
			.99....   .9988..   .6668..
			.99....   .9988..   .6668..
		*/
		{setupNum: 9, setupMove: [2]int{0, 1}, setupNum1: 8, setupMove1: [2]int{0, 3}, nextNum: 6, nextMove: [2]int{0, 1}},
	}
	for _, tt := range tests {
		R, C := 4, 7
		board := newBoardRC(R, C, 1)
		// move 1
		layer := board.addLayer()
		board.putNumber(layer, tt.setupNum, tt.setupMove[0], tt.setupMove[1])
		flat := board.flat
		flat = board.flatten(flat, layer, 1)
		// move 2
		layer = board.addLayer()
		board.putNumber(layer, tt.setupNum1, tt.setupMove1[0], tt.setupMove1[1])
		flat = board.flatten(flat, layer, 1)
		// printLayer(flat)
		br, bc, _ := board.findBestMove(flat, tt.nextNum)
		best := [2]int{br, bc}
		want := tt.nextMove
		if best != want {
			t.Fatalf("want not equal to got: %v != %v", want, best)
		}
	}
}

func TestFindBestMove2Setups(t *testing.T) {
	tests := []struct {
		setupNum   int
		setupMove  [2]int
		setupNum1  int
		setupMove1 [2]int
		nextNum    int
		nextMove   [2]int
	}{
		/*
			setup:    next:     next:     best: 0,4
			.000...   .011...   .011.22
			.0.0...   .0.1...   .0.1.22
			.0.0...   .0.1...   .0.122.
			.000...   .001...   .001222
		*/
		{setupNum: 0, setupMove: [2]int{0, 1}, setupNum1: 1, setupMove1: [2]int{0, 2}, nextNum: 2, nextMove: [2]int{0, 4}},
		/*
			setup:    next:     next:	  best: 0,3
			...999.   ...988.   777988.
			...999.   ...988.   .7.988.
			...99..   ...88..   77.88..
			...99..   ...88..   7..88..
		*/
		{setupNum: 9, setupMove: [2]int{0, 3}, setupNum1: 8, setupMove1: [2]int{0, 3}, nextNum: 7, nextMove: [2]int{0, 0}},
	}
	for _, tt := range tests {
		R, C := 4, 7
		board := newBoardRC(R, C, 1)
		// move 1
		layer := board.addLayer()
		board.putNumber(layer, tt.setupNum, tt.setupMove[0], tt.setupMove[1])
		flat := board.flat
		flat = board.flatten(flat, layer, 1)
		// move 2
		layer = board.addLayer()
		board.putNumber(layer, tt.setupNum1, tt.setupMove1[0], tt.setupMove1[1])
		flat = board.flatten(flat, layer, 2)
		// printLayer(flat)
		br, bc, _ := board.findBestMove(flat, tt.nextNum)
		best := [2]int{br, bc}
		want := tt.nextMove
		if best != want {
			t.Fatalf("want not equal to got: %v != %v", want, best)
		}
	}
}

func TestFindBestMove1Setup(t *testing.T) {
	tests := []struct {
		setupNum  int
		setupMove [2]int
		nextNum   int
		nextMove  [2]int
	}{
		/*
			setup:    next:    best: 0,3
			..000..   ..011..
			..0.0..   ..0.1..
			..0.0..   ..0.1..
			..000..   ..001..
		*/
		{setupNum: 0, setupMove: [2]int{0, 2}, nextNum: 1, nextMove: [2]int{0, 3}},
		/*
			setup:    next:    best: 0,3
			...999.   ....88.
			...999.   ....88.
			...99..   ...88..
			...99..   ...88..
		*/
		{setupNum: 9, setupMove: [2]int{0, 3}, nextNum: 8, nextMove: [2]int{0, 3}},
		/*
			setup:    next:    best: 0,3
			...999.   ...11..
			...999.   ....1..
			...99..   ....1..
			...99..   ....1..
		*/
		{setupNum: 9, setupMove: [2]int{0, 3}, nextNum: 1, nextMove: [2]int{0, 3}},
	}
	for _, tt := range tests {
		R, C := 4, 7
		board := newBoardRC(R, C, 1)
		layer := board.addLayer()
		board.putNumber(layer, tt.setupNum, tt.setupMove[0], tt.setupMove[1])
		flat := board.flat
		flat = board.flatten(flat, layer, 1)
		br, bc, _ := board.findBestMove(flat, tt.nextNum)
		best := [2]int{br, bc}
		want := tt.nextMove
		if best != want {
			t.Fatalf("want not equal to got: %v != %v", want, best)
		}
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		setupNum  int
		setupMove [2]int
		nextNum   int
		nextMove  [2]int
		wantValid bool
		wantLevel int8
	}{
		/*
			setup:    next:
			.999...   .9988..
			.999...   .9988..
			.99....   .988...
			.99....   .988...
		*/
		{setupNum: 9, setupMove: [2]int{0, 1}, nextNum: 8, nextMove: [2]int{0, 2}, wantValid: false, wantLevel: 0},
		/*
			setup:    next:
			.999...   .99988.
			.999...   .99988.
			.99....   .9988..
			.99....   .9988..
		*/
		{setupNum: 9, setupMove: [2]int{0, 1}, nextNum: 8, nextMove: [2]int{0, 3}, wantValid: true, wantLevel: 0},
		/*
			setup:    next:
			.999...   .988...
			.999...   .988...
			.99....   .88....
			.99....   .88....
		*/
		{setupNum: 9, setupMove: [2]int{0, 1}, nextNum: 8, nextMove: [2]int{0, 1}, wantValid: true, wantLevel: 1},
	}
	for _, tt := range tests {
		R, C := 4, 7
		board := newBoardRC(R, C, 1)
		// move 1
		layer := board.addLayer()
		board.putNumber(layer, tt.setupNum, tt.setupMove[0], tt.setupMove[1])
		flat := board.flat
		flat = board.flatten(flat, layer, 0)

		valid, level := board.isValid(flat, tt.nextNum, tt.nextMove[0], tt.nextMove[1])
		if valid != tt.wantValid {
			t.Fatalf("valid: want not equal to got: %v != %v", tt.wantValid, valid)
		}
		if level != tt.wantLevel {
			t.Fatalf("level: want not equal to got: %v != %v", tt.wantLevel, level)
		}
	}
}

func TestIsInBounds(t *testing.T) {
	R, C := 4, 5
	board := newBoardRC(R, C, 1)
	/*
		should be in bounds:
		01110
		01010
		01010
		01110
	*/
	got := board.isInBounds(0, 1)
	want := true
	if want != got {
		t.Fatalf("want not equal to got: %v != %v", want, got)
	}
	/*
		should be out of bounds:
		00000
		01110
		01010
		01010
	*/
	got = board.isInBounds(1, 1)
	want = false
	if want != got {
		t.Fatalf("want not equal to got: %v != %v", want, got)
	}
	/*
		should be out of bounds:
		00011
		00010
		00010
		00011
	*/
	got = board.isInBounds(0, 3)
	if want != got {
		t.Fatalf("want not equal to got: %v != %v", want, got)
	}
}

func TestFlatten(t *testing.T) {
	R, C := 4, 5
	board := newBoardRC(R, C, 1)
	layer := board.addLayer()
	board.putNumber(layer, 0, 0, 1)
	/*
		should be:
		E000E
		E0E0E
		E0E0E
		E000E
	*/
	flat := makeFlatRC(R, C)
	flat = board.flatten(flat, layer, 0)
	/*
		should be:
		.000.
		.0.0.
		.0.0.
		.000.
	*/
	want := &Layer{cells: []int8{-1, 0, 0, 0, -1, -1, 0, -1, 0, -1, -1, 0, -1, 0, -1, -1, 0, 0, 0, -1}}
	if !reflect.DeepEqual(flat.cells, want.cells) {
		fmt.Println("want:")
		board.printLayer(want)
		fmt.Println("got:")
		board.printLayer(flat)
		t.Fatalf("want not equal to got")
	}
	// add another layer and flatten it
	layer = board.addLayer()
	board.putNumber(layer, 1, 0, 2)
	flat = board.flatten(flat, layer, 1)
	/*
		should be:
		.011.
		.0.1.
		.0.1.
		.001.
	*/
	want = &Layer{cells: []int8{-1, 0, 1, 1, -1, -1, 0, -1, 1, -1, -1, 0, -1, 1, -1, -1, 0, 0, 1, -1}}
	if !reflect.DeepEqual(flat.cells, want.cells) {
		fmt.Println("want:")
		board.printLayer(want)
		fmt.Println("got:")
		board.printLayer(flat)
		t.Fatalf("want not equal to got")
	}
}

func TestPutNumberBoundingBox(t *testing.T) {
	R, C := 12, 10
	board := newBoardRC(R, C, 1)
	layer := board.addLayer()
	/*
		should be:
		.......  .......
		.......  .......
		...22..  ...2244
		...22..  ...224.
		..22...  ..22444
		..222..  ..22244
	*/
	board.putNumber(layer, 2, 2, 2)
	if layer.BB_TL_R != 2 || layer.BB_TL_C != 2 {
		board.printLayer(layer)
		t.Fatalf("bounding box top left: want:%v != got:%v", []int{2, 2}, []int{layer.BB_TL_R, layer.BB_TL_C})
	}
	if layer.BB_BR_R != 5 || layer.BB_BR_C != 4 {
		board.printLayer(layer)
		t.Fatalf("bounding box bottom right: want:%v != got:%v", []int{5, 4}, []int{layer.BB_BR_R, layer.BB_BR_C})
	}
	board.putNumber(layer, 4, 2, 4)
	if layer.BB_TL_R != 2 || layer.BB_TL_C != 2 {
		board.printLayer(layer)
		t.Fatalf("bounding box top left: want:%v != got:%v", []int{2, 2}, []int{layer.BB_TL_R, layer.BB_TL_C})
	}
	if layer.BB_BR_R != 5 || layer.BB_BR_C != 6 {
		board.printLayer(layer)
		t.Fatalf("bounding box bottom right: want:%v != got:%v", []int{5, 6}, []int{layer.BB_BR_R, layer.BB_BR_C})
	}
}

func TestPutNumberSkipEmpty(t *testing.T) {
	R, C := 4, 5
	board := newBoardRC(R, C, 1)
	layer := board.addLayer()
	board.putNumber(layer, 2, 0, 0)
	board.putNumber(layer, 4, 0, 2)
	/*
		should be:
		.22..  .2244
		.22..  .224.
		22...  22444
		222..  22244
	*/
	want := &Layer{cells: []int8{-1, 2, 2, 4, 4, -1, 2, 2, 4, -1, 2, 2, 4, 4, 4, 2, 2, 2, 4, 4}}
	if !reflect.DeepEqual(layer.cells, want.cells) {
		fmt.Println("want:")
		board.printLayer(want)
		fmt.Println("got:")
		board.printLayer(layer)
		t.Fatalf("want not equal to got")
	}
}

func TestPutNumber(t *testing.T) {
	R, C := 4, 5
	// layer := makeLayerRC(R, C)
	board := newBoardRC(R, C, 1)
	layer := board.addLayer()
	board.putNumber(layer, 0, 0, 1)
	/*
		should be:
		E000E
		E0E0E
		E0E0E
		E000E
	*/
	want := &Layer{cells: []int8{-1, 0, 0, 0, -1, -1, 0, -1, 0, -1, -1, 0, -1, 0, -1, -1, 0, 0, 0, -1}}
	if !reflect.DeepEqual(layer.cells, want.cells) {
		fmt.Println("want:")
		board.printLayer(want)
		fmt.Println("got:")
		board.printLayer(layer)
		t.Fatalf("want not equal to got")
	}
}

func TestMakeLayer(t *testing.T) {
	R, C := 10, 10
	layer := makeLayerRC(R, C)
	for i := 0; i < R; i++ {
		for j := 0; j < C; j++ {
			if layer.cells[i*C+j] != -1 {
				t.Fatalf("Cell not initialized to -1: %v", layer.cells[i*C+j])
			}
		}
	}
}

func TestPrintLayers(t *testing.T) {
	// R, C := 10, 10
	// layer1 := makeLayerRC(R, C)
	// layer2 := makeLayerRC(R, C)
	// layers := [][][]int8{layer1, layer2}
	board := newBoardRC(10, 10, 1)
	board.printLayers(board.layers)
}
