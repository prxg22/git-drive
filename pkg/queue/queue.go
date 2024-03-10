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

func NewQueue[T any](capacity int) *Queue[T] {
	return &Queue[T]{
		capacity: capacity,
		queue:    make([]T, capacity), // Allocate space for 'capacity' elements
		head:     0,
		tail:     0,
		count:    0,
	}
}

func (q *Queue[T]) Length() int {
	return q.count
}

func (q *Queue[T]) IsFull() bool {
	return q.count == q.capacity
}

func (q *Queue[T]) Enqueue(els ...T) {
	for _, el := range els {
		q.enqueue(el)
	}
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

func (q *Queue[T]) Flush(fn func(el T, i int) error) error {
	i := 0

	for q.count > 0 {
		el, err := q.Dequeue()
		if err != nil {
			break
		}

		err = fn(el, i)
		if err != nil {
			return err
		}
		i++
	}

	return nil
}
