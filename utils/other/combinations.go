// Package combinations provides an iterator that goes through
// combinations in lexicographic order.
//
// # This code is significantly indebted to
//
//   - https://compprog.wordpress.com/2007/10/17/generating-combinations-1/
//     (Note: This blog post has an array out of bounds error bug!)
//   - https://docs.python.org/3/library/itertools.html#itertools.combinations
package otherUtils

// package combinations

// https://github.com/earthboundkid/go-utils/blob/3990067f140bd1e9878f1b7801df008629ad1d52/combinations/combinations.go

// Iterator iterates through the length K combinations of
// an N sized set. E.g. for K = 2 and N = 3, it sets Comb to {0, 1},
// then {0, 2}, and finally {1, 2}.
type Iterator struct {
	N, K int
	Comb []int
}

// Init initializes Comb as []int{0, 1, ..., k-1}. Automatically called
// on first call to Next().
func (c *Iterator) Init() {
	c.Comb = make([]int, c.K)
	for i := 0; i < c.K; i++ {
		c.Comb[i] = i
	}
}

// Next sets Comb to the next combination in lexicographic order.
// Returns false if Comb is already the last combination in
// lexicographic order. Initializes Comb if it doesn't exist.
func (c *Iterator) Next() bool {
	var (
		i int
	)

	if len(c.Comb) != c.K {
		c.Init()
		return true
	}

	// Combination (n-k, n-k+1, ..., n) reached
	// No more combinations can be generated
	if c.Comb[0] == c.N-c.K {
		return false
	}

	for i = c.K - 1; i >= 0; i-- {
		c.Comb[i]++
		if c.Comb[i] < c.N-c.K+1+i {
			break
		}
	}

	// c.Comb now looks like (..., x, n, n, n, ..., n).
	// Turn it into (..., x, x + 1, x + 2, ...)
	for i = i + 1; i < c.K; i++ {
		c.Comb[i] = c.Comb[i-1] + 1
	}

	return true
}
