// Workaholic is a light-weight, in-memory, actor framework.  The name of the package is intended to
// be semantically informative: this is not an executor framework and the tasks are coarse grained.
// For example, a typical task may be "process all incoming service x requests". /end hand waving
//
//
// TODO: doc this
package workaholic

import (
)

//type TaskContext map[string]interface{}
type TaskContext struct {
	parent  *TaskContext
	bindings map[string]interface{}
	ichan <-chan bool
}

// A task defines the signature of a contextual function.
// Both input arguments (for the task) and any side-effecting
// and/or results are also bound to the task context.  The details
// are task specifics.
//
// Workers are dumb and simply invoke tasks and are not concerned about
// the result of the work.  But they do have 2 distinct concerns when
// running tasks:
//
// 1 - the task signals a fault.  A fault indicates that progress can no
// longer be made.  Worker pauses.
//
// 2 - the task signals completion.  Job is done and worker pauses.
//
// 3 - the task was interrupted (if applicable).
//

type taskStatus int8

type Task func(context TaskContext) (interrupt_code, error)

type interrupt_code int8

const (
	_ interrupt_code = iota // Note: fsm depends on 0 here.
	work
	pause
	report
	quit
)
var Interrupts = [...]interrupt_code { work, pause, report, quit}

type status_code int8

const (
	_ status_code = iota - 2
	faulted
	interrupted
	ready
	working
	done
)
var ErrorStatusCodes = [...]status_code {faulted, interrupted, ready, working, done}

// control channels convey interrupt_code (signals) from clients to workers.
// define general, write only, and, real-only types

type controlCh chan interrupt_code
type controlIn <-chan interrupt_code
type Control chan<- interrupt_code

// status_code channels convey status_code (signals) from worker to clients.
// define general, write only, and, real-only types

type statusCh chan status_code
type statusOut chan<- status_code
type Status <-chan status_code

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

func (w *Worker) Pause() {
	w.controlCh <- pause
}

func (w *Worker) Report() {
	w.controlCh <- work
}

func (w *Worker) Stop() {
	w.controlCh <- quit
}

/* TODO:
 * - setup the worker go fms
 * - worker init and return
 */
func NewWorker(name string, id int, context TaskContext, task Task) *Worker {
	w := new(Worker)
	// initialize
	w.controlCh = make(controlCh)
	w.Control = (chan<- interrupt_code)(w.controlCh)
	w.statusCh = make(statusCh)
	w.Status = (<-chan status_code)(w.statusCh)
	w.task = task
	w.context = context

//	go w.fsm()

	return w
}
