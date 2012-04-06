// Workaholic is a light-weight, in-memory, actor framework.  The name of the package is intended to
// be semantically informative: this is not an executor framework and the tasks are coarse grained.
// For example, a typical task may be "process all incoming service x requests". /end hand waving
//
//
// TODO: doc this
package workaholic

import (
//	"time"
)

type TaskContext map[string]interface{}

type Task func(context TaskContext) error

type interrupt int8

const (
	_ interrupt = iota
	work
	pause
	report
	stop
)

type status int8

const (
	_ status = -2
	faulted
	interrupted
	ready
	working
	done
)

// control channels convey interrupt (signals) from clients to workers.
// define general, write only, and, real-only types

type controlCh chan interrupt
type controlIn <-chan interrupt
type Control chan<- interrupt

// status channels convey status (signals) from worker to clients.
// define general, write only, and, real-only types

type statusCh chan status
type statusOut chan<- status
type Status <-chan status

// worker struct is a handle for a worker
type Worker struct {
	/* pkg assigned */
	id        int
	name      string
	controlCh controlCh
	statusCh  statusCh
	/* user assigned */
	task    Task
	context TaskContext
	/* ui */
	Control Control
	Status  Status
}

// REVU: in general, add timed variants for blocking calls

func (w *Worker) Work() {
	w.controlCh <- work
}

func (w *Worker) Puase() {
	w.controlCh <- pause
}

func (w *Worker) Report() {
	w.controlCh <- work
}

func (w *Worker) Stop() {
	w.controlCh <- stop
}

func NewWorker(name string, id int, context TaskContext, task Task) *Worker {
	w := new(Worker)
	// initialize
	w.controlCh = make(controlCh)
	w.Control = (chan<- interrupt)(w.controlCh)
	w.statusCh = make(statusCh)
	w.Status = (<-chan status)(w.statusCh)
	w.task = task
	w.context = context
	return w
}
