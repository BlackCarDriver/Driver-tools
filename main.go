package main

import (
	"flag"
	"github.com/BlackCarDriver/Driver-tools/function/bossExport"
	c "github.com/BlackCarDriver/Driver-tools/function/common"
	"github.com/BlackCarDriver/Driver-tools/function/example"
	"github.com/BlackCarDriver/Driver-tools/function/flowCounter"
	"github.com/astaxie/beego/logs"
)

var (
	fName string
)

func main() {
	loader := c.NewFuncLoader()
	loader.AddFunc(&example.DriverToolExample{})
	loader.AddFunc(&flowCounter.FlowCounter{})
	loader.AddFunc(&bossExport.MongoExport{})

	flag.StringVar(&fName, "f", "bossExport", loader.HelpDesc) // ./main -f=funcName 来指定功能
	flag.Parse()

	err := loader.Run(fName)
	if err != nil {
		logs.Error("exit with error: err=%v", err)
	}
	return
}
