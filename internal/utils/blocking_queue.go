package utils

import (
	"container/list"
	"sync"
)

type BlockingQueue struct {
	list  *list.List
	mutex *sync.Mutex
	cond  *sync.Cond
}

func NewBlockingQueue() *BlockingQueue {
	mutex := sync.Mutex{}
	return &BlockingQueue{
		list:  list.New(),
		mutex: &mutex,
		cond:  sync.NewCond(&mutex),
	}
}

func (q *BlockingQueue) Enqueue(item any) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.list.PushBack(item)
	q.cond.Broadcast()
}

func (q *BlockingQueue) Dequeue() any {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for q.list.Len() == 0 {
		q.cond.Wait()
	}
	value := any(nil)
	element := q.list.Front()
	if element != nil {
		q.list.Remove(element)
		value = element.Value
	}
	return value
}
