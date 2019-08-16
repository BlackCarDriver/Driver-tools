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
	"./function/killer"
	"./function/learner"
)

var (
	function  *string = flag.String("f", "eventer", "Eventer help you record what happen in every days")
	workerMap map[string]func(chan<- func()) (int, error)
)

func init() {
	flag.Parse()
	workerMap = make(map[string]func(chan<- func()) (int, error))
	workerMap["eventer"] = eventer.Run
	workerMap["coster"] = coster.Run
	workerMap["killer"] = killer.Run
	workerMap["learner"] = learner.Run
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
			case c.Coster:
				worker = coster.Run
			case c.Killer:
				worker = killer.Run
			case c.Learner:
				worker = learner.Run
			case c.KillBlacklist: //kill all pid in black list
				killer.KillBlackList()
			default:
				fmt.Println("Exit with unnormal Status: ", exitCode)
			}
			exitSinal <- syscall.Signal(0xc)
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
			if s == syscall.Signal(0xc) { //just save state but not exit
				for i := 0; i < len(tasklist); i++ {
					tasklist[i]()
				}
				tasklist = make([]func(), 0)
			} else { //interrupt
				for i := 0; i < len(tasklist); i++ {
					tasklist[i]()
				}
				os.Exit(1)
			}
		}
	}
}

//#######################  tools function  #####################
