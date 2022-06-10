package main

import (
	"fmt"
	"sync"
)

type Node[T any] struct {
	Item T

	// Reader must wait by reading this Lock
	// and check if the last is read
	// ex
	// someFunc()
	// <-n.Lock
	// if n.Last {
	//  return
	// }
	// NextFunc()
	Lock chan struct{}

	// If It is true then Item
	// is invalid and no other should be written
	Last bool

	Next *Node[T]
}

// One Producer multiple consumer queue
type Queue[T any] struct {
	sync.Mutex
	N *Node[T]
}

type Reader[T any] struct {
	N *Node[T]
}

func NewNode[T any]() *Node[T] {
	return &Node[T]{Lock: make(chan struct{})}
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{N: NewNode[T]()}
}

func (q *Queue[T]) NewReader() Reader[T] {
	q.Lock()
	defer q.Unlock()
	r := Reader[T]{N: q.N}
	return r
}

func (q *Queue[T]) Insert(item T) error {
	if q.N == nil {
		return fmt.Errorf("writing to closed queue")
	}

	q.Lock()
	defer q.Unlock()

	n := NewNode[T]()
	q.N.Item = item
	q.N.Next = n
	// This will unblock all the goroutines listening
	// on the lock
	close(q.N.Lock)
	q.N = n

	return nil
}

func (q *Queue[T]) Close() {
	q.Lock()
	defer q.Unlock()

	q.N.Last = true
	// This will unblock all the goroutines listening
	// on the lock
	close(q.N.Lock)
	q.N = nil
}

func (r *Reader[T]) Read() (T, bool) {
	<-r.N.Lock
	if r.N.Last {
		var val T
		return val, true
	}

	item := r.N.Item
	r.N = r.N.Next
	return item, false
}

func (r *Reader[T]) Wait() chan struct{} {
	return r.N.Lock
}
