package queue


// Queue is a generic queue that can hold elements of any type
type Queue[T any] struct {
	elements []T
}

// Enqueue adds an element to the end of the queue
func (q *Queue[T]) Enqueue(value T) {
	q.elements = append(q.elements, value)
}

// Dequeue removes and returns the element at the front of the queue
func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.elements) == 0 {
		var zero T // Zero value of T
		return zero, false
	}
	front := q.elements[0]
	q.elements = q.elements[1:]
	return front, true
}

// IsEmpty checks if the queue is empty
func (q *Queue[T]) IsEmpty() bool {
	return len(q.elements) == 0
}

// Front returns the element at the front of the queue without removing it
func (q *Queue[T]) Front() (T, bool) {
	if len(q.elements) == 0 {
		var zero T // Zero value of T
		return zero, false
	}
	return q.elements[0], true
}
