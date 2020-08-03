package main

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func Test_Canonical(t *testing.T) {
	p := &Position{
		Tubes: []Tube{
			{1, 1, 1, 1},
			{2, 2, 0, 0},
			{2, 2, 3, 0},
			{0, 0, 0, 0},
		},
	}

	expected := &Position{
		Tubes: []Tube{
			{0, 0, 0, 0},
			{1, 1, 0, 0},
			{1, 1, 2, 0},
			{3, 3, 3, 3},
		},
	}

	t.Log("Input:\n", p)

	p.MakeCanonical(3)

	t.Log("Canonical:\n", p)
	t.Log("Expected:\n", expected)

	for i, tube := range p.Tubes {
		if tube != expected.Tubes[i] {
			t.Error("Mismatch at offet", i)
		}
	}
}

// Courtesy of https://gist.github.com/quux00/8258425
func ShuffleInts(rand *rand.Rand, slice []int) {
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func (p *Position) Generate(rand *rand.Rand, _ int) reflect.Value {
	size := 6
	colors := 4

	ret := &Position{
		Tubes: make([]Tube, size),
	}
	ballsAndBlanks := make([]int, size*4)
	for i := 0; i < colors; i++ {
		ballsAndBlanks[i*4+0] = i + 1
		ballsAndBlanks[i*4+1] = i + 1
		ballsAndBlanks[i*4+2] = i + 1
		ballsAndBlanks[i*4+3] = i + 1
	}
	ShuffleInts(rand, ballsAndBlanks)
	for i := 0; i < size; i++ {
		ret.Tubes[i][0] = byte(ballsAndBlanks[i*4+0])
		ret.Tubes[i][1] = byte(ballsAndBlanks[i*4+1])
		ret.Tubes[i][2] = byte(ballsAndBlanks[i*4+2])
		ret.Tubes[i][3] = byte(ballsAndBlanks[i*4+3])

		top := 3
		for top > 0 && ret.Tubes[i][top] == 0 {
			top--
		}
		for j := 0; j < top; j++ {
			if ret.Tubes[i][j] == 0 {
				ret.Tubes[i][j] = ret.Tubes[i][top]
				ret.Tubes[i][top] = 0
				top--
			}
		}
	}
	return reflect.ValueOf(ret)
}

// Courtesy of https://stackoverflow.com/questions/30226438/generate-all-permutations-in-go
func permutations(arr []int) [][]int {
	var helper func([]int, int)
	res := [][]int{}

	helper = func(arr []int, n int) {
		if n == 1 {
			tmp := make([]int, len(arr))
			copy(tmp, arr)
			res = append(res, tmp)
		} else {
			for i := 0; i < n; i++ {
				helper(arr, n-1)
				if n%2 == 1 {
					tmp := arr[i]
					arr[i] = arr[n-1]
					arr[n-1] = tmp
				} else {
					tmp := arr[0]
					arr[0] = arr[n-1]
					arr[n-1] = tmp
				}
			}
		}
	}
	helper(arr, len(arr))
	return res
}

func (p *Position) LessThanOrEqual(other *Position) bool {
	if len(p.Tubes) != len(other.Tubes) {
		return false
	}
	for i := range p.Tubes {
		for k := 0; k < 4; k++ {
			if p.Tubes[i][k] < other.Tubes[i][k] {
				return true
			}
			if p.Tubes[i][k] > other.Tubes[i][k] {
				return false
			}
		}
	}
	return true
}

func (p *Position) SlowCanonical(out chan<- *Position) {
	tubePerms := permutations([]int{0, 1, 2, 3, 4, 5})
	colorPerms := permutations([]int{1, 2, 3, 4})

	for _, t := range tubePerms {
		for _, c := range colorPerms {
			// 0 always maps to 0
			cp := append([]int{0}, c...)

			other := &Position{
				Tubes: append([]Tube{}, p.Tubes...),
			}
			for i := range other.Tubes {
				src := p.Tubes[t[i]]
				other.Tubes[i][0] = byte(cp[src[0]])
				other.Tubes[i][1] = byte(cp[src[1]])
				other.Tubes[i][2] = byte(cp[src[2]])
				other.Tubes[i][3] = byte(cp[src[3]])
			}
			out <- other
		}
	}
	close(out)

}

func Test_Canonical_Rand(t *testing.T) {
	f := func(p *Position) bool {
		p.MakeCanonical(4)
		c := make(chan *Position, 1)
		go p.SlowCanonical(c)
		for other := range c {
			if !p.LessThanOrEqual(other) {
				t.Log("Canonical:\n", p)
				t.Log("Not less than:\n", other)
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
