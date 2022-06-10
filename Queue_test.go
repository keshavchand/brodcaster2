package main

import (
	"sync"
	"testing"
)

type Counter struct {
	curr int
}

func (c *Counter) next() int {
	defer func() { c.curr++ }()
	return c.curr
}

func TestConcurrent(t *testing.T) {
	queue := NewQueue[int]()
	reader := func(r Reader[int], last *int) {
		var l int
		c := Counter{0}
		for {
			i, end := r.Read()
			if end {
				break
			}
			expected := c.next()
			if i != expected {
				t.Fatalf("Input values not same as expected %v vs %v", i, expected)
			}
		}
		*last = l
	}

	res := make([]int, 10)
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int, r Reader[int]) {
			defer wg.Done()
			reader(r, &res[i])
		}(i, queue.NewReader())
	}
	for i := 0; i < 10; i++ {
		err := queue.Insert(i)
		if err != nil {
			t.Fatalf("inserting %v", err)
		}
	}
	queue.Close()
}
