package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	c "./function/common"
	"./function/coster"
	"./function/eventer"
)

var (
	function  *string = flag.String("f", "coster", "Eventer help you record what happen in every days")
	workerMap map[string]func(chan<- func()) (int, error)
)

func init() {
	flag.Parse()
	workerMap = make(map[string]func(chan<- func()) (int, error))
	workerMap["eventer"] = eventer.Run
	workerMap["coster"] = coster.Run
}

func main() {
	worker, ok := workerMap[*function]
	if !ok {
		c.ColorPrint(c.Light_red, "No such function: %s", *function)
		return
	}
	taskBus := make(chan func())
	exitSinal := make(chan os.Signal)
	go destructor(taskBus, exitSinal)
	for {
		exitCode, err := worker(taskBus)
		if err != nil {
			fmt.Println(err)
		} else {
			switch exitCode {
			case c.NormalReturn:
				exitSinal <- syscall.Signal(0xa)
				time.Sleep(1 * time.Second)
			case c.Eventer:
				worker = eventer.Run
				continue
			case c.Coster:
				worker = coster.Run
				continue
			default:
				fmt.Println("Exit with unnormal Status: ", exitCode)
			}
			c.ClearConsole()
		}
	}
}

//execute task function when receive a interupt signal
func destructor(task <-chan func(), exitSig chan os.Signal) {
	tasklist := make([]func(), 0)
	signal.Notify(exitSig) //Monitor all signal
	for {
		select {
		case t := <-task: //receive a task
			tasklist = append(tasklist, t)
		case s := <-exitSig: //execute task function before exit
			fmt.Printf("\nReceive exit signal: %v\n", s)
			for i := 0; i < len(tasklist); i++ {
				tasklist[i]()
			}
			os.Exit(1)
		}
	}
}
