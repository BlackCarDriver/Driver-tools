package main

import (
	"flag"
	c "github.com/BlackCarDriver/Driver-tools/function/common"
	"github.com/BlackCarDriver/Driver-tools/function/coster"
	"github.com/BlackCarDriver/Driver-tools/function/easyShot"
	"github.com/BlackCarDriver/Driver-tools/function/eventer"
	"github.com/BlackCarDriver/Driver-tools/function/example"
	"github.com/BlackCarDriver/Driver-tools/function/flowCounter"
	"github.com/BlackCarDriver/Driver-tools/function/stock"
	"github.com/astaxie/beego/logs"
)

var (
	fName string
)

func main() {
	loader := c.NewFuncLoader()
	loader.AddFunc(&example.DriverToolExample{})
	loader.AddFunc(&eventer.EventLog{})
	loader.AddFunc(&coster.CostLog{})
	loader.AddFunc(&flowCounter.FlowCounter{})
	loader.AddFunc(&stock.StockTool{})
	loader.AddFunc(&easyShot.EasyShoot{})

	flag.StringVar(&fName, "f", "event", loader.HelpDesc) // 功能名称
	//flag.StringVar(&fName, "f", "stock", loader.HelpDesc) // 功能名称
	flag.Parse()

	err := loader.Run(fName)
	if err != nil {
		logs.Error("exit with error: err=%v", err)
	}
	return
}
