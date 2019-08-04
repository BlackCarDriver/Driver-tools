package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"./function/common"
	"./function/eventer"
)

var (
	function  *string = flag.String("f", "eventer", "temp")
	workerMap map[string]func(chan<- func()) (int, error)
)

func init() {
	flag.Parse()
	workerMap = make(map[string]func(chan<- func()) (int, error))
	workerMap["eventer"] = eventer.Run
}

func main() {
	worker, ok := workerMap[*function]
	taskBus := make(chan func())
	exitSinal := make(chan os.Signal)
	go destructor(taskBus, exitSinal)
	if ok {
		exitCode, err := worker(taskBus)
		if err != nil {
			fmt.Println(err)
		} else if exitCode == common.NormalReturn { //normally  exit
			exitSinal <- syscall.Signal(0xa)
			time.Sleep(1 * time.Second)
		} else {
			fmt.Println("Exit with unnormal Status: ", exitCode)
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
