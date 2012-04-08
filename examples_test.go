package workaholic

import (
	"fmt"
	"math/rand"
)

func ExampleWorker() {
	// channel capacity for bounded channels of Worker
	qlen := 1024 * 1024
	// user assigned worker id
	id := 11
	// user assigned worker name
	name := "hard-worker"

	// create worker -- it is ready to run
	worker := NewWorker(name, id, qlen)

	// you can queue up to qlen tasks
	for _, task := range makeTasks(qlen) {
		worker.Command <- task // REVU: bug exposed - buff len ..
	}

	// Workers will send status in response to
	// Report and Quit control interrupts.
	//
	// We haven't told 'hard-worker' to start work yet,
	// so expecting a status report of Idle
	// Remember: Status is a blocking channel of size 1 and
	// for every Report request you must consume the resultant
	// status (or subsequent requests will block the worker).
	worker.Control <- Report
	fmt.Println(<-worker.Status)

	// Tell it to start on its queue
	//
	worker.Control <- Work

	// At this point expecting worker to be busy
	//
	worker.Control <- Report
	fmt.Println(<-worker.Status)

	// we can pause the worker, and
	// tell it to resume working.
	//
	worker.Control <- Pause
	worker.Control <- Report
	worker.Control <- Work
	fmt.Println(<-worker.Status)

	// we can add more work as long as the
	// Command channel is not full.
	for _, task := range makeTasks(1000) {
		worker.Command <- task // REVU: bug exposed - buff len ..
	}

	worker.Control <- Quit
	fmt.Println(<-worker.Status)

	// Output:
	// Idle
	// Busy
	// Idle
	// Terminated
}

func makeTasks(n int) []Task {
	tasks := make([]Task, n)
	for i, _ := range tasks {
		_i := int32(i)
		tasks[i] = func() {
			a := rand.Int31n(1000000) * _i
			b := rand.Int31n(1000) * _i
			_ = a % b
		}
	}
	return tasks
}
