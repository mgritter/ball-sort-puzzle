package main

import (
	"bytes"
	"container/heap"
	"fmt"
	"sort"
	"strings"
)

type Tube [4]byte

type Position struct {
	Tubes []Tube
}

// Start positions have some number of empty tubes, which
// in canonical order will be in front.  (In the game they
// are typically at the end.)
func (p *Position) IsStartPosition(numEmpty int) bool {
	emptyTube := Tube{0, 0, 0, 0}
	for i := 0; i < numEmpty; i++ {
		if p.Tubes[i] != emptyTube {
			return false
		}
	}
	return true
}

const displayRunes = " ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func (p *Position) String() string {
	var buf strings.Builder

	for _, t := range p.Tubes {
		buf.WriteString("|")
		for _, val := range t {
			buf.WriteByte(displayRunes[val])
		}
		buf.WriteString("|\n")
	}
	return buf.String()
}

func (p *Position) TopBall(tubeIndex int) int {
	t := p.Tubes[tubeIndex]
	switch {
	case t[3] != 0:
		return 3
	case t[2] != 0:
		return 2
	case t[1] != 0:
		return 1
	case t[0] != 0:
		return 0
	default:
		return -1
	}
}

func (p *Position) ReverseMove(from int, to int, topBall int) *Position {
	c := &Position{
		Tubes: append([]Tube{}, p.Tubes...),
	}
	color := c.Tubes[to][topBall]
	c.Tubes[to][topBall] = 0
	pos := p.TopBall(from) + 1
	c.Tubes[from][pos] = color
	return c
}

func (p *Position) Predecessors() []*Position {
	preds := make([]*Position, 0, 8)
	// The top ball in each tube can have been moved there only if
	// it is in position 0, or if the ball below it is the same color.
	// It can come from any tube that isn't full
	for to, toTube := range p.Tubes {
		b := p.TopBall(to)
		if b == -1 {
			// No ball
			continue
		}
		color := toTube[b]
		if b > 0 {
			if toTube[b-1] != color {
				// No match
				continue
			}
		}
		for from, fromTube := range p.Tubes {
			if from == to {
				continue
			}
			if fromTube[3] == 0 {
				preds = append(preds, p.ReverseMove(from, to, b))
			}
		}
	}
	return preds
}

func (p *Position) CanonicalPredecessors(numColors int) []*Position {
	allPreds := p.Predecessors()
	for _, pp := range allPreds {
		pp.MakeCanonical(numColors)
	}
	return allPreds
}

func (p *Position) Key() string {
	var buf bytes.Buffer
	for _, t := range p.Tubes {
		buf.Write(t[:])
	}
	return buf.String()
}

func NewEndPosition(colors int, spare int) *Position {
	tubes := make([]Tube, colors+spare)
	for i := 0; i < spare; i++ {
		tubes[i] = Tube{0, 0, 0, 0}
	}
	for i := 0; i < colors; i++ {
		bv := byte(i + 1)
		tubes[i+spare] = Tube{bv, bv, bv, bv}
	}
	return &Position{
		Tubes: tubes,
	}
}

// The canonical form of a position sorts the tubes in alphabetical order,
// (permuting the order of the tubes) and permutes the colors to be the
// alphabetically-lowest string.

// Thus, any empty tubes come first: ____
// Followed by A___, AA__, AAA_, AAAA
// Next-best is those that have to use a second letter:
// AAAB, AAB_, AABA, AABB, AB__, ABA_, ABAA, ABAB
// etc.
// within each "Type" we'd like to tiebreak by later words.
// So, I did this as a search algorithm because I couldn't come up with
// heuristics that I believed in.
//
// For example, if we have
// A___
// B___
// Bxxx
// Bxxx
// Bxxx
// Then we should remap B to A, to move the three other rows
// up in the order.
//

type LowerBound struct {
	Bytes []byte
}

type Mapping struct {
	Colors    map[byte]byte
	NextColor byte
	Bound     LowerBound
}

type MappingQueue struct {
	Elements []*Mapping
}

func NewMappingQueue() *MappingQueue {
	return &MappingQueue{
		Elements: []*Mapping{},
	}
}

// sort interface for LowerBound
func (b *LowerBound) Len() int {
	return len(b.Bytes) / 4
}

func (b *LowerBound) Swap(i, j int) {
	b.Bytes[i*4+0], b.Bytes[j*4+0] = b.Bytes[j*4+0], b.Bytes[i*4+0]
	b.Bytes[i*4+1], b.Bytes[j*4+1] = b.Bytes[j*4+1], b.Bytes[i*4+1]
	b.Bytes[i*4+2], b.Bytes[j*4+2] = b.Bytes[j*4+2], b.Bytes[i*4+2]
	b.Bytes[i*4+3], b.Bytes[j*4+3] = b.Bytes[j*4+3], b.Bytes[i*4+3]
}

func (b *LowerBound) Less(i, j int) bool {
	for k := 0; k < 4; k++ {
		if b.Bytes[i*4+k] < b.Bytes[j*4+k] {
			return true
		}
		if b.Bytes[i*4+k] > b.Bytes[j*4+k] {
			return false
		}
	}
	return false
}

func NewMapping() *Mapping {
	return &Mapping{
		Colors:    map[byte]byte{0: 0},
		NextColor: 1,
		Bound: LowerBound{
			Bytes: []byte{},
		},
	}
}

func ExtendMapping(m *Mapping, p *Position, toAssign byte) *Mapping {
	mapCopy := make(map[byte]byte)
	for k, v := range m.Colors {
		if k == toAssign {
			return nil
		}
		mapCopy[k] = v
	}

	mapCopy[toAssign] = m.NextColor

	lb := []byte{}
	for _, t := range p.Tubes {
		for k := 0; k < 4; k++ {
			origColor := t[k]
			mappedColor, ok := mapCopy[origColor]
			if !ok {
				mappedColor = m.NextColor + 1
			}
			lb = append(lb, mappedColor)
		}
	}

	nm := &Mapping{
		Colors:    mapCopy,
		NextColor: m.NextColor + 1,
		Bound: LowerBound{
			Bytes: lb,
		},
	}
	sort.Sort(&nm.Bound)
	return nm
}

// sort and heap interfaces for MappingQueue
func (q *MappingQueue) Push(x interface{}) {
	q.Elements = append(q.Elements, x.(*Mapping))
}

func (q *MappingQueue) Pop() interface{} {
	ret := q.Elements[len(q.Elements)-1]
	q.Elements = q.Elements[:len(q.Elements)-1]
	return ret
}

func (q *MappingQueue) Len() int {
	return len(q.Elements)
}

func (q *MappingQueue) Less(i, j int) bool {
	bound_i := q.Elements[i].Bound.Bytes
	bound_j := q.Elements[j].Bound.Bytes
	length_i := len(bound_i)
	length_j := len(bound_j)
	min_length := length_i
	if length_j < min_length {
		min_length = length_j
	}
	for k := 0; k < min_length; k++ {
		if bound_i[k] < bound_j[k] {
			return true
		}
		if bound_i[k] > bound_j[k] {
			return false
		}
	}
	// Shortest wins
	if length_i == length_j {
		return false
	}
	if length_i < length_j {
		return true
	}
	return false
}

func (q *MappingQueue) Swap(i, j int) {
	q.Elements[i], q.Elements[j] = q.Elements[j], q.Elements[i]
}

const debugPriorityQueue = false

func (p *Position) MakeCanonical(numColors int) {
	queue := NewMappingQueue()

	heap.Push(queue, NewMapping())
	var bestMap *Mapping

	for queue.Len() > 0 {
		bestMap = heap.Pop(queue).(*Mapping)

		if debugPriorityQueue {
			fmt.Printf("Best mapping: %v Lower bound:\n%v\n", bestMap.Colors, bestMap.Bound.Bytes)
		}

		if int(bestMap.NextColor) > numColors {
			break
		}

		// Try each of the unassigned colors as the
		// next one to remap.
		for c := 1; c <= numColors; c++ {
			newMap := ExtendMapping(bestMap, p, byte(c))
			if newMap != nil {
				heap.Push(queue, newMap)
				if debugPriorityQueue {
					fmt.Printf("New mapping: %v Lower bound:\n%v\n", newMap.Colors, newMap.Bound.Bytes)
				}
			}
		}

	}

	// The mappig already has the position in sorted
	// order, so copy it rather than resorting.
	flat := bestMap.Bound.Bytes
	for i := range p.Tubes {
		p.Tubes[i][0] = flat[i*4+0]
		p.Tubes[i][1] = flat[i*4+1]
		p.Tubes[i][2] = flat[i*4+2]
		p.Tubes[i][3] = flat[i*4+3]
	}
}