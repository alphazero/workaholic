package main

import (
	"fmt"
	"time"
	"workaholic/process"
)

func NoOp(context interface{}) interface{} {
	inargs := context.([]int)
	if inargs[1] > 10 {
		panic("TOO BIG!")
	}
	fmt.Printf("* debug * no-op * task %d of %d\n", inargs[1], inargs[0])
	return nil
}

func main() {
	var p process.Process
	p = process.New()

	go func() {
		for i := 0; i < 100; i++ {
			task := process.NewOp(NoOp, []int{100, i}, nil)
			p.Ops() <- task
		}
	}()

	var tx *process.Transition
	tx = p.Signal(process.Status)
	fmt.Printf("* debug * main * tx-ack => %s %s\n", tx)

	tx = p.Signal(process.Start)

	tx = p.Signal(process.Status)
	fmt.Printf("* debug * main * tx-ack => %s %s\n", tx)

	select {
	case <-make(chan bool, 1):
	case <-time.NewTimer(time.Nanosecond * 1).C:
	}

	awhile := 10
	for awhile > 0 {
		var timeout bool
		tx, timeout = p.TrySignal(process.Status, time.Nanosecond*100000)
		fmt.Printf("* debug * main * tx-ack => %s %s\n", tx)
		if timeout {
			fmt.Printf("* debug * main * TrySignal timeout\n")
			continue
		}
		awhile--

		select {
		case <-time.NewTimer(time.Second * 1).C:
			go func() {
				for i := 0; i < 20; i++ {
					task := process.NewOp(NoOp, []int{20, i}, nil)
					p.Ops() <- task
				}
			}()
		}
	}

	tx = p.Signal(process.Status)
	fmt.Printf("* debug * main * tx-ack => %s %s\n", tx)

	tx = p.Signal(process.Stop)

	tx = p.Signal(process.Status)
	fmt.Printf("* debug * main * tx-ack => %s %s\n", tx)

	tx = p.Signal(process.Shutdown)

	tx = p.Signal(process.Status)
	fmt.Printf("* debug * main * tx-ack => %s %s\n", tx)

}
