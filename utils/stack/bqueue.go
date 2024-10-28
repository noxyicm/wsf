package stack

import (
	"bytes"
	"sync/atomic"
	"unsafe"
)

type bnode struct {
	value *bytes.Buffer
	next  *bnode
}

// BQueue is a queue
type BQueue struct {
	head *bnode
	tail *bnode
	len  int64
}

// Enqueue add pointer to queue
func (q *BQueue) Enqueue(x *bytes.Buffer) bool {
	newNode := new(bnode)
	newNode.value = x
	var tail *bnode
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
func (q *BQueue) Dequeue() (*bytes.Buffer, bool) {
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
func (q *BQueue) Len() int64 {
	return atomic.LoadInt64(&q.len)
}

// NewBQueue creates a new queue
func NewBQueue() *BQueue {
	n := new(bnode)
	q := new(BQueue)
	q.head = n
	q.tail = n
	return q
}
