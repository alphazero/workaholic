//   Copyright 2009-2012 Joubin Houshyar
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

// process defines a simple finite state model for a Process and
// defines the api for a command and control interface for it.
//
// Processes are long-running go routines that sequentially process
// operations from an ops channel.  Overlooking the user-provided op functions,
// a Process is designed to be interruptible and can pause, restart, or
// formally shutdown.
//
// Process is controlled by sending interrupt signals via blocking or non-blocking
// Signal and TrySignal, respectively.
package process

import (
	"fmt"
	"time"
)

type State int

const (
	nil_state State = iota
	s_startup
	s_idle
	s_running
	s_shutdown
)

func (s State) String() string {
	switch s {
	case nil_state:
		return "nil:State"
	case s_startup:
		return "startup:State"
	case s_idle:
		return "idle:State"
	case s_running:
		return "running:State"
	case s_shutdown:
		return "shutdown:State"
	}
	panic(fmt.Errorf("BUG : unreachable in State.String()"))
}

// Interrupts are used to control the process
type Interrupt int

const (
	nil_interrupt Interrupt = iota

	// interrupt code issued to worker to initiate,
	// or to resume, work if previously paused or faulted.
	Start

	// interrupt code issued to worker to commend working
	Stop

	// interrupt code issued to worker to stop working
	// Worker's can resume work subsequently via interrupt Work.
	Status

	// interrupt code issued to worker to quit (terminate).
	Shutdown
)

func (c Interrupt) String() string {
	switch c {
	case nil_interrupt:
		return "nil:Interrupt"
	case Start:
		return "Start"
	case Stop:
		return "Stop"
	case Status:
		return "Status"
	case Shutdown:
		return "Shutdown"
	}
	panic(fmt.Errorf("BUG : unreachable in Interrupt.String()"))
}

// Transition captures the transitional state of the process.
// Process is transition from a state to another. Info provides
// optional additional information.
type Transition struct {
	From, To State
	Info     interface{}
}

func (t Transition) String() string {
	return fmt.Sprintf("Transition:{from:%s to:%s info:'%s'}", t.From, t.To, t.Info)
}

// A process is a lightly regulated go routine conforming to a simple
// finite state model.
type Process interface {
	Signal(Interrupt) *Transition
	TrySignal(Interrupt, time.Duration) (*Transition, bool)
	Ops() chan<- *Op
}

type OpFunc func(context interface{}) (result interface{})
type Op struct {
	opFunc          OpFunc
	context, result interface{}
}

func NewOp(fn OpFunc, context, outargs interface{}) *Op {
	return &Op{fn, context, outargs}
}

type interrupt struct {
	code  Interrupt
	reply chan *Transition
}

func newSignal(code Interrupt) *interrupt {
	return &interrupt{
		code:  code,
		reply: make(chan *Transition, 1),
	}
}

// REVU - process can be beefed up with startup/shutdown hooks and the done.
// TODO - is fault
type process struct {
	tasks   chan *Op
	signals chan *interrupt
}

func New() Process {
	w := &process{
		tasks:   make(chan *Op, 1000),
		signals: make(chan *interrupt, 10),
	}
	go w.fms()

	return w
}

func (w process) TrySignal(s Interrupt, wait time.Duration) (tx *Transition, timeout bool) {
	event := newSignal(s)
	w.signals <- event

	select {
	case tx = <-event.reply:
		return
	case <-time.NewTimer(wait).C:
		timeout = true
		return
	}
	panic(fmt.Errorf("BUG : unreachable in process.TrySignal"))
}

var ShuttingDown = &Transition{s_shutdown, s_shutdown, nil}

func (w process) Signal(s Interrupt) (tx *Transition) {
	defer func() {
		if p := recover(); p != nil {
			tx = ShuttingDown
		}
	}()
	event := newSignal(s)
	w.signals <- event
	return <-event.reply
}

func (w process) Ops() chan<- *Op {
	return w.tasks
}

func (w process) fms() {

	var s0 State
	var event *interrupt
	var stat interface{}
	var tx *Transition

	var s1 State
	var last time.Time
	//	var fault interface{} // TODO

	goto startup

	panic("BUG : unreachable")
Transition:

	switch event.code {
	case nil_interrupt:
		if s0 == s_startup {
			s1 = s_idle
		}
	case Stop:
		s1 = s_idle
	case Start:
		s1 = s_running
	case Status:
		s1 = s0
	case Shutdown:
		s1 = s_shutdown
	default:
		panic(fmt.Errorf(" BUG : unknown interrupt %s", event.code))
	}

	tx = &Transition{s0, s1, stat}
	event.reply <- tx  // REVU - ideally after the transition ..

	switch tx.To {
	case s_idle:
		goto wait
	case s_running:
		goto run
	case s_shutdown:
		goto shutdown
	default:
		panic(fmt.Errorf(" BUG : unexpected to state %s", tx.To))
	}

	panic("BUG : unreachable")
startup:
	// do the startup thing - REVU TODO fn it ..
	s0 = s_startup
	event = newSignal(nil_interrupt)
	stat = fmt.Sprintf("start up - timenow: %d", time.Now())
	goto Transition

	panic("BUG : unreachable")
wait:
	s0 = s_idle
	select {
	case event = <-w.signals:
		stat = fmt.Sprintf("event:%s\n", event)
	}
	goto Transition

	panic("BUG : unreachable")
run:

	s0 = s_running
	select {
	case s := <-w.signals:
		stat = fmt.Sprintf("runing - interrupted - stat: %d", stat)
		event = s
		goto Transition
	case task := <-w.tasks:
		now := time.Now()
		task.result = w.performOp(task.opFunc, task.context)
		if isError(task.result) {
			fmt.Printf("* debug * fsm * FAULTED - %s\n", task.result)
			// TODO faults
		}
		stat = fmt.Sprintf("runing - Op %s - timenow: %s timelast: %d\n", task, now, last)
		last = now
	}
	goto run

	panic("BUG : unreachable")
shutdown:
	s0 = s_shutdown
	close(w.signals)
	// REVU - do the shutdown fn ..
}

func isError(e interface{}) bool {
	if e != nil {
		switch e.(type) {
		case error:
			return true
		}
	}
	return false
}
func (w process) performOp(op OpFunc, context interface{}) (result interface{}) {
	defer func() {
		if p := recover(); p != nil {
			if p != nil {
				result = fmt.Errorf("operation panic - %s", p)
			}
		}
	}()

	return op(context)
}
