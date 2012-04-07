// Workaholic is a light-weight treatment to provide command & control semantics
// for long running goroutines aka workers. Each worker is a simple FSM exposing
// command and control input channels and a very basic instrumentation (status)
// output channels.

package workaholic

import (
	"log" // TEMP
)

// ----------------------------------------------------------------------------
// Tasks
// ----------------------------------------------------------------------------

// A Task is a function
// REVU: review this (contexts, names instead of functions .. etc.)
type Task func()

type control_code int8
const (
	_ control_code = iota // Note: fsm depends on 0 here.
	Work
	Pause
	Report
	Quit
)
/* used by tests */
var Interrupts = [...]control_code { Work, Pause, Report, Quit}
/* used by tests */


type Status_code int8
const (
	_ Status_code = iota - 3
	faulted
	interrupted
	idle
	busy
	terminated
)
func (s Status_code) String() string {
	switch s {
	case faulted:
		return "faulted"
	case interrupted:
		return "interrupted"
	}
	return "not-coded"
}
/* used by tests */
var criticalStats = 2
var StatusCodes = [...]Status_code {faulted, interrupted, idle, busy}
/* used by tests */

// Status_code channels convey Status_code (signals) from worker to clients.
// define general, write only, and, real-only types

//type statusCh chan Status_code
type statusOut chan<- Status_code

// ----------------------------------------------------------------------------
// Worker
// ----------------------------------------------------------------------------

type Command chan<- Task
type Control chan<- control_code
type Status <-chan Status_code
type Faults <-chan interface{}

// worker struct is a handle for a worker
type Worker struct {

	/* identity */

	Id        int
	Name      string

	/* internals */

	commandCh chan Task
	statusCh  chan Status_code
	controlCh chan control_code
	fsmstat Status_code

	/* ui */

	// Command is a buffered channel of Task
	// used to issue work for worker
	Command Command

	// Control is a len 1 write only channel of control_code
	// used to issue interrupts to the worker
	Control Control

	// Status is a len 1 read only channel of Status_code
	// used by worker to send status to user
	Status  Status
}

func NewWorker(name string, id int, qlen int) *Worker {
	w := new(Worker)
	// initialize
	w.commandCh = make(chan Task, qlen)
	w.Command = (chan<- Task)(w.commandCh)

	w.controlCh = make(chan control_code, 1)
	w.Control = (chan<- control_code)(w.controlCh)

	w.statusCh = make(chan Status_code, 1)         // REVU: the 1 bothers me .. a bit
	w.Status = (<-chan Status_code)(w.statusCh)

	go w.fsm()

	return w
}

func (w *Worker) fsm()  {

	var controller = (<-chan control_code)(w.controlCh)
	var commander = (<-chan Task)(w.commandCh)
    var statout = (chan<- Status_code)(w.statusCh)

	w.fsmstat = idle // in this state only at startup and after a Pause command

await_signal:
	interrupt := <-controller

interrupted:
	w.fsmstat = interrupted

	switch interrupt {
	case Work:
		goto Work
	case Pause:
		goto Pause
	case Report:
		goto Report
	case Quit:
		goto shutdown
	}
panic("SHOULD NOT BE REACHED")

Work:
	w.fsmstat = busy

	select {
	case interrupt = <-controller:
		goto interrupted
	case task := <-commander:
		w.perform(task)
		goto Work
	}

Pause:
	w.fsmstat = idle

	goto await_signal

Report:
	// TODO: do the Report
	// REVU: do we even need this?
	goto await_signal

//faulted:
////	w.statusCh <- workerStatus{id, fault, taskstatus, &controller}
//	w.statusCh <- faulted
//	goto await_signal

shutdown:
	w.fsmstat = terminated

	statout<- terminated
	log.Println("shutdown ..")
	// TODO: add shutdown hook for worker
}

func (w *Worker) perform(task Task) {
	defer func() {
		if p := recover(); p != nil {
		// REVU: get rid of status and use faults
//			w.faultsCh <- p
			w.statusCh <- faulted
		}
	}()

	task()
}
