package main

import (
	"container/heap"
	"fmt"
	"math/rand"
	"time"
)

const nRequester = 100
const nWorker = 10

// This is a load balancer practice with round-robin and  least loading models
// reference link: https://gist.github.com/chenlujjj/f2cc6b75e5276e41bf82b5d561fcf28f
func main() {
	work := make(chan Request)
	for i := 0; i < nRequester; i++ {
		go requester(work, Workfn)
	}
	NewBalancer().balance(work)

}

type Request struct {
	fn func() int // The operation to perform
	c  chan int   // The channel to return the result
}

func requester(work chan<- Request, Fn func() int) {
	c := make(chan int)
	for {
		// kill some time (fake loading)
		time.Sleep(2 * time.Second)
		work <- Request{Fn, c} // send request
		result := <-c
		furtherProcess(result)
	}
}

// Simulate some work : sleep for a while and return how long
func Workfn() int {
	n := rand.Int63n(int64(time.Second))
	// time.Sleep(time.Duration(nWorker * n))
	time.Sleep(time.Second)
	return int(n)
}
func furtherProcess(result int) {
	fmt.Println("Work time: ", result)
}

type Worker struct {
	requests chan Request // work to do (buffered channel)
	pending  int          // counts of pending tasks
	index    int          // index in the heap
}

func (w *Worker) work(done chan *Worker) {
	for {
		select {
		case req := <-w.requests: //get Request from balancer
			req.c <- req.fn() // cal fn and send result
			done <- w         //we've finish the request
		default:
			continue
		}

	}

}

// Balancer worker pool
type Pool []*Worker

// less pending worker gets work
func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending
}
func (p Pool) Len() int {
	return len(p)
}
func (p *Pool) Swap(i, j int) {
	a := *p
	a[i], a[j] = a[j], a[i]
	a[i].index = i
	a[j].index = j
}
func (p *Pool) Pop() interface{} {
	old := *p
	tmp := old[len(*p)-1]
	*p = old[0 : len(*p)-1]
	tmp.index = -1
	return tmp
}
func (p *Pool) Push(in interface{}) {
	// Push 和 Pop 要透過指標傳值，因為會更動到 slice 的長度而不只是內容
	// 解釋：
	// append 過後的切片可能已經不是本來的切片
	// 所以要透過指標更改 h 的值
	in.(*Worker).index = p.Len()
	*p = append(*p, in.(*Worker))
}

type Balancer struct {
	pool Pool
	done chan *Worker
	i    int  // Round-robin mode, worker index
	rr   bool // Round-robin mode switch
}

func NewBalancer() *Balancer {
	done := make(chan *Worker, nWorker)
	b := &Balancer{make(Pool, 0, nWorker), done, 0, false}
	for i := 0; i < nWorker; i++ {
		w := &Worker{requests: make(chan Request, nRequester)}
		heap.Push(&b.pool, w)
		go w.work(b.done)
	}
	return b
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
		b.print()
	}
}

// print out Wokers in pool and wokers pending numbers, avg and std deviation.
func (b *Balancer) print() {
	sum := 0
	sumq := 0
	fmt.Println("Workers in pool: ", b.pool.Len())

	for _, w := range b.pool {
		fmt.Printf("worker index:%d, has %d pendings\n", w.index, w.pending)
		sum += w.pending
		sumq += w.pending * w.pending
	}
	avg := float64(sum) / float64(b.pool.Len())
	variance := float64(sumq)/float64(b.pool.Len()) - avg*avg
	fmt.Printf("Avg and variance %.2f %.2f\n", avg, variance)

}

// Send Request to worker
func (b *Balancer) dispatch(req Request) {
	// if runnig in round robin mode
	if b.rr {
		w := b.pool[b.i]  // get the worker
		w.requests <- req // send the request
		w.pending++       // increment the pending
		b.i++             // next worker index
		if b.i >= b.pool.Len() {
			b.i = 0 // reset worker index
		}
		return
	}
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
	if b.rr {
		w.pending--
		return
	}
	// one fewer in the queue
	w.pending--
	// remove from the heap
	heap.Remove(&b.pool, w.index)
	// put it into its place on the heap
	heap.Push(&b.pool, w)

}
