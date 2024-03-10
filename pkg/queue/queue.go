package queue

import (
	"fmt"
)

type Queue[T any] struct {
	capacity int
	queue    []T
	head     int
	tail     int
	count    int // Added to keep track of the number of elements in the queue
}

// NewQueue creates a new queue with the specified capacity.
// The capacity determines the maximum number of elements that the queue can hold.
// The queue is initialized with zero elements.
// The generic type parameter 'T' represents the type of elements that the queue can hold.
func NewQueue[T any](capacity int) *Queue[T] {
	return &Queue[T]{
		capacity: capacity,
		queue:    make([]T, capacity), // Allocate space for 'capacity' elements
		head:     0,
		tail:     0,
		count:    0,
	}
}

// Length returns the number of elements in the queue.
func (q *Queue[T]) Length() int {
	return q.count
}

// IsFull checks if the queue is full.
// It returns true if the number of elements in the queue is equal to its capacity, and false otherwise.
func (q *Queue[T]) IsFull() bool {
	return q.count == q.capacity
}

// Enqueue adds elements to the queue.
// It takes one or more elements of type T and adds them to the end of the queue.
// If the queue reaches its capacity, Enqueue will overwrite the queue head
func (q *Queue[T]) Enqueue(els ...T) {
	for _, el := range els {
		q.enqueue(el)
	}
}

// Dequeue removes and returns the element at the front of the queue.
// If the queue is empty, it returns an error.
func (q *Queue[T]) Dequeue() (T, error) {
	if q.count == 0 {
		var zero T
		return zero, fmt.Errorf("queue is empty")
	}

	el := q.queue[q.head]
	q.head = (q.head + 1) % q.capacity
	q.count--

	return el, nil
}

// Iterate returns a channel that allows iterating over the elements in the queue.
// The channel will be closed once all elements have been iterated.
func (q *Queue[T]) Iterate() chan T {
	c := make(chan T, q.Length())

	go func() {
		defer close(c)

		for q.count > 0 {
			el, err := q.Dequeue()

			if err != nil {
				break
			}
			c <- el
		}
	}()

	return c
}

func (q *Queue[T]) enqueue(el T) {
	if q.count == q.capacity {
		q.queue[q.tail] = el
		q.head = (q.head + 1) % q.capacity
		q.tail = (q.tail + 1) % q.capacity
	} else {
		q.queue[q.tail] = el
		q.tail = (q.tail + 1) % q.capacity
		q.count++
	}
}
