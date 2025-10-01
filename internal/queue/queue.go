package queue

type Queue struct {
    ch chan interface{}
}

func NewQueue() *Queue {
    return &Queue{ch: make(chan interface{}, 100)}
}

func (q *Queue) Push(item interface{}) {
    q.ch <- item
}

func (q *Queue) Pop() <-chan interface{} {
    return q.ch
}
