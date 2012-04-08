// Workaholic is a light-weight treatment to provide command & control semantics
// for long running goroutines aka workers. Each worker is a simple FSM exposing
// command and control input channels and a very basic instrumentation (status)
// output channels.

package workaholic

import (
//	"log" // TEMP
)

// A Task is a user provided function
// REVU: review this (contexts, names instead of functions .. etc.)
type Task func()

// Task fault in case of panic
type Fault struct {
	task  Task
	fault interface{}
}

type interrupt_code int8

const (
	_zerovalue interrupt_code = iota // SPEC: must be 0

	// interrupt code issued to worker to initiate,
	// or to resume, work if previously paused or faulted.
	Work

	// interrupt code issued to worker to commend working
	Pause

	// interrupt code issued to worker to stop working
	// Worker's can resume work subsequently via interrupt Work.
	Report

	// interrupt code issued to worker to quit (terminate).
	Quit
)

// Mainly used by tests and a pseudo enum for the end user
var Interrupts = [...]interrupt_code{Work, Pause, Report, Quit}

type status_code int8

const (
	_ status_code = iota - 3

	// A status code.
	// Faulted indicates that the worker (itself) has faulted
	// and has stopped working. A worker can attempt to recover
	// from Faulted state if given interrupt to Work.
	Faulted

	// A status code.
	// Interrupted indicated that the worker was interrupted while
	// performing a task.  (Not yet supported)

	Interrupted
	// A status code.
	// Idle indicates that the worker is ready and able to work.
	// In this state, worker will respond to Work, Report, and Quit.
	Idle

	// A status code.
	// Busy indicates the worker is busy doing work.
	// Worker will respond to Pause, Report and Quit.
	// REVU: broken -- task is not interruptible.
	Busy

	// A status code.
	// Terminated indicates that the worker has entered a terminal
	// state.  Provided channels to worker are closed at this point.
	Terminated
)

// Mainly used by tests and a pseudo enum for the end user
var Statuses = [...]status_code{Faulted, Interrupted, Idle, Busy, Terminated}

// Pretty print for status codes
func (s status_code) String() string {
	switch s {
	case Faulted:
		return "Faulted"
	case Interrupted:
		return "Interrupted"
	case Idle:
		return "Idle"
	case Busy:
		return "Busy"
	case Terminated:
		return "Terminated"
	}
	return "not-coded"
}

/* used by tests */
var criticalStats = 2

/* used by tests */

// status_code channels convey status_code (signals) from worker to clients.
// define general, write only, and, real-only types

//type statusCh chan status_code
type statusOut chan<- status_code

// ----------------------------------------------------------------------------
// Worker
// ----------------------------------------------------------------------------

// Command is a bounded send-only channel of Task
type Command chan<- Task

// Control is a bounded send-only channel of interrupt_code
type Control chan<- interrupt_code

// Status is a bounded receive-only channel of status_code
type Status <-chan status_code

// Faults is a bounded receive-only channel of Fault
type Faults <-chan Fault

// worker struct is a handle for a worker
type Worker struct {

	/* identity */

	Id   int
	Name string

	/* internals */

	commandCh chan Task
	statusCh  chan status_code
	controlCh chan interrupt_code
	faultsCh  chan Fault
	fsmstat   status_code

	/* Worker interface for use by Worker's client */

	// Channel for issuing Task(s) to the Worker
	Command

	// Channel for issuing interrupts to the Worker
	Control

	// Status is a len 1 read only channel of status_code
	// used by worker to send status in response to Report requests
	Status

	// used by worker to send status to user
	Faults
}

func NewWorker(name string, id int, qlen int) *Worker {
	w := new(Worker)
	// initialize
	w.commandCh = make(chan Task, qlen)
	w.Command = (chan<- Task)(w.commandCh)

	w.controlCh = make(chan interrupt_code, 1)
	w.Control = (chan<- interrupt_code)(w.controlCh)

	w.statusCh = make(chan status_code, 1) // REVU: the 1 bothers me .. a bit
	w.Status = (<-chan status_code)(w.statusCh)

	w.faultsCh = make(chan Fault, qlen) // REVU: hate these qlen knobs
	w.Faults = (<-chan Fault)(w.faultsCh)

	go w.fsm()

	return w
}

func (w *Worker) fsm() {

	var controller = (<-chan interrupt_code)(w.controlCh)
	var commander = (<-chan Task)(w.commandCh)
	var statout = (chan<- status_code)(w.statusCh)

	var interrupt, lastinterrupt interrupt_code

	w.fsmstat = Idle // in this state only at startup and after a Pause command

_await_signal:
	lastinterrupt = interrupt
	interrupt = <-controller

_interrupted:
	//	w.fsmstat = Interrupted // REVU: this is wrong for various reasons.

	switch interrupt {
	case Work:
		goto _work
	case Pause:
		goto _pause
	case Report:
		goto _report
	case Quit:
		goto _shutdown
	default:
		goto _await_signal
	}
	panic("SHOULD NOT BE REACHED")

_work:
	w.fsmstat = Busy
	lastinterrupt = interrupt
	select {
	case interrupt = <-controller:
		goto _interrupted
	case task := <-commander:
		w.perform(task)
		//		if taskInterrupt := w.perform(task); taskInterrupt == _zerovalue {
		//			interrupt = taskInterrupt
		//			goto _interrupted
		//		}
		goto _work
	}

_pause:
	w.fsmstat = Idle

	goto _await_signal

_report:
	statout <- w.fsmstat
	interrupt = lastinterrupt
	goto _interrupted

	//Faulted:
	////	w.statusCh <- workerStatus{id, fault, taskstatus, &controller}
	//	w.statusCh <- Faulted
	//	goto _await_signal

_shutdown:
	w.fsmstat = Terminated

	statout <- w.fsmstat
	// TODO: add shutdown hook for worker
}

func (w *Worker) perform(task Task) {
	defer func() {
		if p := recover(); p != nil {
			w.faultsCh <- Fault{task, p}
			//			w.statusCh <- Faulted
		}
	}()

	task()
}
