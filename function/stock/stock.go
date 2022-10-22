package stock

import (
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
)

type StockTool struct{}

func (s *StockTool) GetInfo() (name string, desc string) {
	return "stock", "股市行情小助手"
}

func (s *StockTool) Exit() {
	fmt.Println("bey bey")
	return
}

func (s *StockTool) Run() (retCmd string, err error) {
	err = readConfig()
	if err != nil {
		util.ScanStdLine()
		retCmd = "end"
		return
	}
	s.printWelcome()
	for {
		rawInput := util.ScanStdLine()
		switch rawInput {
		case "turn", "end", "exit": // 交到外部处理
			retCmd = rawInput
			return
		case "clear":
			util.ClearConsole()
			continue
		case "watch":
			breakSig := make(chan string)
			go startMonitorStock(breakSig)
			any := scanStdLine()
			breakSig <- any
		case "overall":
			breakSig := make(chan string)
			go startMonitorOverAll(breakSig)
			any := scanStdLine()
			breakSig <- any
		default:
			color.Red("未知命令,请重新输入...")
		}
	}
	return "exit", nil
}

// 打印使用帮助
func (s *StockTool) printWelcome() {
	util.ClearConsole()
	color.HiRed("\n============================\n==     行情数据小助手     ==\n============================\n")
	color.Magenta("可用命令:\noverall - 全盘概述 \nwatch - 自选盯盘")
	color.Magenta("end - 退出\nturn - 切换功能\nclear - 清空控制台\n")
}
