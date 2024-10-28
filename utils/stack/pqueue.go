package stack

import (
	"sync/atomic"
	"unsafe"
)

type pnode struct {
	value unsafe.Pointer
	next  *pnode
}

// PQueue is a queue of pointers
type PQueue struct {
	head *pnode
	tail *pnode
	len  int64
}

// Enqueue add pointer to queue
func (q *PQueue) Enqueue(x unsafe.Pointer) bool {
	newNode := new(pnode)
	newNode.value = x
	var tail *pnode
	for {
		tail = q.tail
		next := tail.next
		if next != nil {
			atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)), unsafe.Pointer(tail), unsafe.Pointer(next))
			continue
		}

		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next)), nil, unsafe.Pointer(newNode)) {
			break
		}
	}

	casok := atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)), unsafe.Pointer(tail), unsafe.Pointer(newNode))
	atomic.AddInt64(&q.len, 1)

	return casok
}

// Dequeue extracts pointer from queue
func (q *PQueue) Dequeue() (unsafe.Pointer, bool) {
	for {
		firstNode := q.head
		lastNode := q.tail
		nextNode := firstNode.next
		if firstNode != q.head {
			continue
		}

		if firstNode == lastNode {
			if nextNode == nil {
				return nil, false
			}

			atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)), unsafe.Pointer(lastNode), unsafe.Pointer(nextNode))
			continue
		}

		x := nextNode.value
		if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.head)), unsafe.Pointer(firstNode), unsafe.Pointer(nextNode)) {
			continue
		}

		atomic.AddInt64(&q.len, -1)
		return x, true
	}
}

// Len returns length of queue
func (q *PQueue) Len() int64 {
	return atomic.LoadInt64(&q.len)
}

// NewPQueue creates a new queue
func NewPQueue() *PQueue {
	n := new(pnode)
	q := new(PQueue)
	q.head = n
	q.tail = n
	return q
}
