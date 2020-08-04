package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

type Config struct {
	NumColors      int
	NumSpares      int
	NumWorkers     int
	CpuProfileName string
	MemProfileName string
}

func (config *Config) WriteMemProfile(suffix int) {
	if config.MemProfileName != "" {
		filename := fmt.Sprintf("%v.%v", config.MemProfileName, suffix)
		f, err := os.Create(filename)
		if err != nil {
			fmt.Printf("could not create memory profile: %v\n", err)
			return
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Printf("could not write memory profile: %v\n", err)
		}
	}
}

func main() {
	var config Config

	flag.IntVar(&config.NumColors, "colors", 4, "number of colors")
	flag.IntVar(&config.NumSpares, "spares", 2, "number of spare locations")
	flag.IntVar(&config.NumWorkers, "workers", 2, "number of parallel workers to use")
	flag.StringVar(&config.CpuProfileName, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&config.MemProfileName, "memprofile", "", "write memory profile to `file`")

	flag.Parse()

	if config.CpuProfileName != "" {
		f, err := os.Create(config.CpuProfileName)
		if err != nil {
			fmt.Printf("could not create CPU profile: %v\n", err)
			return
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Printf("could not start CPU profile: %v\n", err)
			return
		}
		defer pprof.StopCPUProfile()
	}

	enumerateGames(&config)
}
