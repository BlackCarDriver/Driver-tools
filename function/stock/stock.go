package stock

import (
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"strings"
	"time"
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
		case "overall":
			breakSig := make(chan string)
			go startMonitorOverAll(breakSig)
			any := util.ScanStdLine()
			breakSig <- any
		default:
			color.Red("未知命令,请重新输入...")
		}
	}
	return "exit", nil
}

// ====================================

// 打印使用帮助
func (s *StockTool) printWelcome() {
	color.HiRed("\n============================\n==     行情数据小助手     ==\n============================\n")
	color.Magenta("可用命令:\noverall - 全盘概述 \n")
	color.Magenta("end - 退出\nturn - 切换功能\nclear - 清空控制台\n")
}

// 涨跌状况监控
func startMonitorOverAll(breakSig <-chan string) {
	tr := time.NewTicker(time.Minute)
	var input string
	util.ClearConsole()
	color.Blue("监控开始,输入任何东西结束~")
	printOverAll()
	for len(input) == 0 {
		select {
		case <-tr.C:
			printOverAll()
		case input = <-breakSig:
			break
		}
	}
	color.Blue("监控结束~")
}

func printOverAll() {
	ret, err := getOverAllData()
	if err != nil {
		return
	}
	up, hold, down, flow := ret.F104, ret.F106, ret.F106, ret.F3
	total := float64(up + hold + down)
	maxLen := 60.0
	upLen := int(float64(up) / total * maxLen)
	downLen := int(float64(down) / total * maxLen)
	holdLen := int(maxLen) - upLen - downLen
	upStr := color.RedString("%s", strings.Repeat("#", upLen))
	holdStr := color.HiBlackString("%s", strings.Repeat("#", holdLen))
	downStr := color.GreenString("%s", strings.Repeat("#", downLen))
	flowStr := color.HiRedString("%.2f", flow)
	if flow < 0 {
		flowStr = color.HiBlueString("%.2f", flow)
	}
	fmt.Printf("%s %s%s%s %s\n", time.Now().Format("15:04"), upStr, holdStr, downStr, flowStr)
}
