package main

import (
	"flag"
)

func main() {
	colorPtr := flag.Int("colors", 4, "number of colors")
	sparePtr := flag.Int("spares", 2, "number of spare locations")

	flag.Parse()
	enumerateGames(*colorPtr, *sparePtr)
}
