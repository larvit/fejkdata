// Package avoid draws a value made of fixed runs of characters that holds none of a set of strings.
package avoid

import "encoding/binary"

// Run is N characters, each one byte of Alphabet.
type Run struct {
	Alphabet string
	N        int
}

// Runs is every value a builtin can print: its runs, one after another.
type Runs []Run

// Len is the length of each value of rs.
func (rs Runs) Len() int {
	n := 0
	for _, r := range rs {
		n += r.N
	}
	return n
}

// Rand draws a whole number from 0 up to n, n left out.
type Rand interface{ IntN(n int) int }

// Plan draws the values of its runs holding none of its patterns.
type Plan struct {
	runs  Runs
	a     *automaton
	ahead []sequence // ahead[i].at(k): the states from which k more characters of run i, and every later run, can be drawn
}

// Compile plans the draws of the values of rs holding none of patterns; it returns nil where each value
// holds one. A pattern holding a byte no run draws is ignored; an empty one is held by every value.
func Compile(rs Runs, patterns []string) *Plan {
	a := newAutomaton(rs, patterns)
	p := &Plan{runs: rs, a: a, ahead: make([]sequence, len(rs))}
	end := a.live()
	for i := len(rs) - 1; i >= 0; i-- {
		p.ahead[i] = a.sequence(end, rs[i])
		end = p.ahead[i].at(rs[i].N)
	}
	if !end.has(0) {
		return nil
	}
	return p
}

// Draw draws a value, each character evenly among those that still lead to a value holding no pattern.
func (p *Plan) Draw(r Rand) string {
	out := make([]byte, 0, p.runs.Len())
	allowed := make([]byte, 0, 64)
	state := 0
	for i, run := range p.runs {
		for k := run.N - 1; k >= 0; k-- {
			ahead := p.ahead[i].at(k)
			allowed = allowed[:0]
			for _, c := range []byte(run.Alphabet) {
				if ahead.has(p.a.step(state, c)) {
					allowed = append(allowed, c)
				}
			}
			c := allowed[r.IntN(len(allowed))]
			out = append(out, c)
			state = p.a.step(state, c)
		}
	}
	return string(out)
}

// automaton is the Aho–Corasick automaton of a set of patterns over the bytes a set of runs draws: a
// state is the longest end of the text read that begins a pattern, and accepts once the text holds one.
type automaton struct {
	symbol  [256]int // the column of a byte in next, -1 for a byte no run draws
	next    [][]int
	accepts []bool
}

func newAutomaton(rs Runs, patterns []string) *automaton {
	a := &automaton{}
	symbols := 0
	for i := range a.symbol {
		a.symbol[i] = -1
	}
	for _, r := range rs {
		for _, c := range []byte(r.Alphabet) {
			if a.symbol[c] < 0 {
				a.symbol[c] = symbols
				symbols++
			}
		}
	}
	a.addState(symbols)
	for _, p := range patterns {
		a.insert(p, symbols)
	}
	a.link(symbols)
	return a
}

func (a *automaton) addState(symbols int) int {
	row := make([]int, symbols)
	for i := range row {
		row[i] = -1
	}
	a.next = append(a.next, row)
	a.accepts = append(a.accepts, false)
	return len(a.next) - 1
}

// insert adds p to the trie, unless no run draws one of its bytes.
func (a *automaton) insert(p string, symbols int) {
	for i := 0; i < len(p); i++ {
		if a.symbol[p[i]] < 0 {
			return
		}
	}
	s := 0
	for i := 0; i < len(p); i++ {
		col := a.symbol[p[i]]
		if a.next[s][col] < 0 {
			a.next[s][col] = a.addState(symbols)
		}
		s = a.next[s][col]
	}
	a.accepts[s] = true
}

// link fills each missing step with the step of the state's longest proper end in the trie, breadth
// first, and makes a state accept where that end does.
func (a *automaton) link(symbols int) {
	fail := make([]int, len(a.next))
	var queue []int
	for col := range symbols {
		if c := a.next[0][col]; c > 0 {
			queue = append(queue, c)
		} else {
			a.next[0][col] = 0
		}
	}
	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]
		a.accepts[s] = a.accepts[s] || a.accepts[fail[s]]
		for col := range symbols {
			if c := a.next[s][col]; c < 0 {
				a.next[s][col] = a.next[fail[s]][col]
			} else {
				fail[c] = a.next[fail[s]][col]
				queue = append(queue, c)
			}
		}
	}
}

func (a *automaton) step(state int, c byte) int { return a.next[state][a.symbol[c]] }

// live is every state that accepts no pattern.
func (a *automaton) live() bitset {
	b := newBitset(len(a.next))
	for s, accepts := range a.accepts {
		if !accepts {
			b.add(s)
		}
	}
	return b
}

// before is every live state from which one character of alphabet leads into after.
func (a *automaton) before(after bitset, alphabet string) bitset {
	b := newBitset(len(a.next))
	for s, accepts := range a.accepts {
		if accepts {
			continue
		}
		for _, c := range []byte(alphabet) {
			if after.has(a.step(s, c)) {
				b.add(s)
				break
			}
		}
	}
	return b
}

// sequence is the sets of states from which k more characters of a run can be drawn ending in the first
// set: sets[k] up to the first set that repeats an earlier one, sets[loop], from where they cycle.
type sequence struct {
	sets []bitset
	loop int
}

func (q sequence) at(k int) bitset {
	if k < len(q.sets) {
		return q.sets[k]
	}
	return q.sets[q.loop+(k-q.loop)%(len(q.sets)-q.loop)]
}

func (a *automaton) sequence(end bitset, r Run) sequence {
	q := sequence{sets: []bitset{end}}
	seen := map[string]int{end.key(): 0}
	for k := 1; k <= r.N; k++ {
		b := a.before(q.sets[k-1], r.Alphabet)
		if at, repeats := seen[b.key()]; repeats {
			q.loop = at
			return q
		}
		seen[b.key()] = k
		q.sets = append(q.sets, b)
	}
	return q
}

type bitset []uint64

func newBitset(n int) bitset    { return make(bitset, (n+63)/64) }
func (b bitset) has(i int) bool { return b[i/64]&(1<<(i%64)) != 0 }
func (b bitset) add(i int)      { b[i/64] |= 1 << (i % 64) }
func (b bitset) key() (key string) {
	buf := make([]byte, 0, 8*len(b))
	for _, w := range b {
		buf = binary.LittleEndian.AppendUint64(buf, w)
	}
	return string(buf)
}
