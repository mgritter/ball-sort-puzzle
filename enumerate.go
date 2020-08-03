package main

import (
	"fmt"
)

func enumerateGames(numColors, numSpares int) {
	currentDepth := 0
	known_positions := make(map[string]bool)

	p := NewEndPosition(numColors, numSpares)

	fmt.Printf("Depth %2v: %8v total %8v starts \n", 0, 1, 1)
	known_positions[p.Key()] = true
	prev_positions := []*Position{p}

	for len(prev_positions) > 0 {
		currentDepth += 1
		next_positions := make([]*Position, 0, len(prev_positions))
		for _, p := range prev_positions {
			predecessors := p.CanonicalPredecessors(numColors)
			for _, np := range predecessors {
				key := np.Key()
				if !known_positions[key] {
					next_positions = append(next_positions, np)
					known_positions[key] = true
				}
			}
		}
		numStartPositions := 0
		var example *Position
		for _, p := range next_positions {
			if p.IsStartPosition(numSpares) {
				numStartPositions += 1
				example = p
			}
		}
		fmt.Printf("Depth %2v: %8v total %8v starts \n", currentDepth, len(next_positions), numStartPositions)
		if example != nil {
			fmt.Printf("Example:\n%v\n", example)
		}
		prev_positions = next_positions
	}
}
