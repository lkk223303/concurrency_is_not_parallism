package main

import (
	"container/heap"
	"time"
)

func main() {

}

type Request struct {
	fn func() int // The operation to perform
	c  chan int   // The channel to return the result
}

func requester(work chan<- Request) {
	c := make(chan int)
	for {
		// kill some time (fake loading)
		time.Sleep(2 * time.Second)
		work <- Request{workFn, c} // send request
		result := <-c
		furtherProcess(result)
	}
}

type Worker struct {
	requests chan Request // work to do (buffered channel)
	pending  int          // counts of pending tasks
	index    int          // index in the heap
}

func (w *Worker) work(done chan *Worker) {
	for {
		req := <-w.requests //get Request from balancer
		req.c <- req.fn()   // cal fn and send result
		done <- w           //we've finish the request
	}
}

// Balancer
type Pool []*Worker
type Balancer struct {
	pool Pool
	done chan *Worker
}

// need to implement dispatch and completed
func (b *Balancer) balance(work chan Request) {
	for {
		select {
		case req := <-work: // receive a Request
			b.dispatch(req) //so send it to the worker
		case w := <-b.done: // a worker has finished the request
			b.completed(w) // so update its info
		}
	}
}

func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending

}
func (p Pool) Len() int {
	return len(p)
}
func (p Pool) Pop() (out interface{}) {
	out = p[len(p)-1]
	return
}
func (p Pool) Push(out interface{}) {
	p = append(p, out.(*Worker))
}

// Send Request to worker
func (b *Balancer) dispatch(req Request) {
	// Grab the least loaded worker...
	w := heap.Pop(&b.pool).(*Worker)
	// ...send it the task.
	w.requests <- req
	// One more in its work queue.
	w.pending++
	// Put it into its place on the heap.
	heap.Push(&b.pool, w)
}

// jos is completed update the heap
func (b *Balancer) completed(w *Worker) {
	// one fewer in the queue
	w.pending--
	// remove from the heap
	heap.Remove(&b.pool, w.index)
	// put it into its place on the heap
	heap.Push(&b.pool, w)

}
