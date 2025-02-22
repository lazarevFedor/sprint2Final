package pkg

import "sync"

type Queue struct {
	queue []interface{}
	mutex sync.Mutex
}

func (cq *Queue) Enqueue(element interface{}) {
	cq.mutex.Lock()
	cq.queue = append(cq.queue, element)
	cq.mutex.Unlock()
}

func (cq *Queue) Dequeue() interface{} {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	if len(cq.queue) == 0 {
		return nil
	}

	element := cq.queue[0]
	cq.queue = cq.queue[1:]
	return element
}

func (cq *Queue) Peek() interface{} {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	if len(cq.queue) == 0 {
		return nil
	}

	return cq.queue[0]
}

func (cq *Queue) IsEmpty() bool {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	return len(cq.queue) == 0
}
