package main

import (
	"fmt"
)

func enumerateGames(config *Config) {
	currentDepth := 0
	known_positions := make(map[string]bool)

	p := NewEndPosition(config.NumColors, config.NumSpares)

	fmt.Printf("Depth %2v: %8v total %8v starts \n", 0, 1, 1)
	known_positions[p.Key(config.NumColors)] = true
	prev_positions := []*Position{p}

	worker := func(prev_positions []*Position, out chan []*Position) {
		local_cache := make(map[string]bool)
		next_positions := make([]*Position, 0, len(prev_positions))
		for _, p := range prev_positions {
			predecessors := p.CanonicalPredecessors(config.NumColors)
			for _, np := range predecessors {
				key := np.Key(config.NumColors)
				if !known_positions[key] && !local_cache[key] {
					next_positions = append(next_positions, np)
					local_cache[key] = true
				}
			}
		}

		out <- next_positions
		close(out)
	}

	for len(prev_positions) > 0 {
		currentDepth += 1
		next_positions := make([]*Position, 0, len(prev_positions))

		num_workers := config.NumWorkers
		if len(prev_positions) < 100 {
			num_workers = 1
		}

		batchSize := len(prev_positions) / num_workers
		workers := make([]struct {
			C       chan []*Position
			Results []*Position
		}, num_workers)

		for i := range workers {
			workers[i].C = make(chan []*Position, 2)
			var batch []*Position
			if i == num_workers-1 {
				batch = prev_positions[i*batchSize:]
			} else {
				batch = prev_positions[i*batchSize : (i+1)*batchSize]
			}
			go worker(batch, workers[i].C)
		}

		numStartPositions := 0
		var example *Position

		// To avoid race conditions, collect all results first
		for i := range workers {
			workers[i].Results = <-workers[i].C
		}

		for i := range workers {
			// Add each to the shared map
			for _, p := range workers[i].Results {
				key := p.Key(config.NumColors)
				if known_positions[key] {
					continue
				}
				known_positions[key] = true
				next_positions = append(next_positions, p)
				if p.IsStartPosition(config.NumSpares) {
					numStartPositions += 1
					example = p
				}
			}
		}

		// This should be the peak?
		config.WriteMemProfile(currentDepth)

		fmt.Printf("Depth %2v: %8v total %8v starts \n", currentDepth, len(next_positions), numStartPositions)
		if example != nil {
			fmt.Printf("Example:\n%v\n", example)
		}
		prev_positions = next_positions
	}
}
