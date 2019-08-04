package main

import (
	"flag"
	"fmt"

	"./function/eventer"
)

var (
	function  *string = flag.String("f", "eventer", "temp")
	workerMap map[string]func() (int, error)
)

func init() {
	flag.Parse()
	workerMap = make(map[string]func() (int, error))
	workerMap["eventer"] = eventer.Run
}

func main() {
	worker, ok := workerMap[*function]
	if ok {
		res, err := worker()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Return with: %d \n", res)
	}
}

// func destructor() {
// 	c := make(chan os.Signal)
// 	signal.Notify(c) //Monitor all signal
// 	<-c
// 	logs.Info("Recover a interrupt and save enveter data!")
// 	saveState()
// 	os.Exit(1)
// }
