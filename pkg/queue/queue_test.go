package queue_test

import (
	"testing"

	"github.com/prxg22/git-drive/pkg/queue"
)

func TestQueueEnqueueDequeue(t *testing.T) {
	q := queue.NewQueue[int](5)

	_, err := q.Dequeue()
	if err == nil {
		t.Errorf("Expected error empty queue!")
	}
	// Enqueue elements
	q.Enqueue(1, 2, 3)

	// Dequeue elements
	el, err := q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if el != 1 {
		t.Errorf("Expected 1, got %v", el)
	}

	el, err = q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 2 {
		t.Errorf("Expected 2, got %v", el)
	}

	el, err = q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 3 {
		t.Errorf("Expected 3, got %v", el)
	}

	// Fill the queue to its capacity
	q.Enqueue(1, 2, 3, 4, 5)

	// Enqueue additional elements
	q.Enqueue(6, 7, 8)

	// Dequeue additional elements
	el, err = q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 4 {
		t.Errorf("Expected 4, got %v", el)
	}

	el, err = q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 5 {
		t.Errorf("Expected 5, got %v", el)
	}

	el, err = q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 6 {
		t.Errorf("Expected 6, got %v", el)
	}

	el, err = q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 7 {
		t.Errorf("Expected 7, got %v", el)
	}

	el, err = q.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if el != 8 {
		t.Errorf("Expected 8, got %v", el)
	}

	_, err = q.Dequeue()
	if err == nil {
		t.Errorf("Expected error empty queue!")
	}

}

func TestQueueLength(t *testing.T) {
	q := queue.NewQueue[int](5)
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

func TestQueueFlush(t *testing.T) {
	q := queue.NewQueue[int](5)

	// Enqueue elements
	q.Enqueue(1, 2, 3)

	// Define a function to flush the queue
	flushFn := func(el int, i int) error {
		// Perform some action with the element
		// In this example, we'll just print the element
		t.Logf("Flushing element %v at index %v", el, i)
		return nil
	}

	// Flush the queue
	q.Flush(flushFn)

	// Check if the queue is empty after flushing
	length := q.Length()
	if length != 0 {
		t.Errorf("Expected length 0 after flushing, got %v", length)
	}
}
