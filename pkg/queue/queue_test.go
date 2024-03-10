package queue_test

import (
	"testing"

	"github.com/prxg22/git-drive/pkg/queue"
)

func TestQueueEnqueueDequeue(t *testing.T) {
	q := queue.NewQueue[int](5)

	// Enqueue elements
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	// Dequeue elements
	el, err := q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 1 {
		t.Errorf("Expected element 1, got %v", el)
	}

	el, err = q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 2 {
		t.Errorf("Expected element 2, got %v", el)
	}

	el, err = q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 3 {
		t.Errorf("Expected element 3, got %v", el)
	}

	// Check length
	length := q.Length()
	if length != 0 {
		t.Errorf("Expected length 0, got %v", length)
	}
}

func TestQueueLength(t *testing.T) {
	q := queue.NewQueue[int](5)

	// Check length of an empty queue
	length := q.Length()
	if length != 0 {
		t.Errorf("Expected length 0, got %v", length)
	}

	// Enqueue elements
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	// Check length
	length = q.Length()
	if length != 3 {
		t.Errorf("Expected length 3, got %v", length)
	}

	q.Enqueue(4)
	q.Enqueue(5)
	q.Enqueue(6)

	length = q.Length()
	if length != 5 {
		t.Errorf("Expected length 5, got %v", length)
	}
}

func TestQueueIterate(t *testing.T) {
	q := queue.NewQueue[int](5)

	// Enqueue elements
	q.Enqueue(1, 2, 3, 4, 5, 6)

	// Iterate over the queue
	i := 2
	for n := range q.Iterate() {
		if n != i {
			t.Errorf("Expected element %v, got %v", i, n)
		}
		i++
	}
}
