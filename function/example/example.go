package example

import (
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
)

// DriverToolExample 实现示例
type DriverToolExample struct{}

func (d *DriverToolExample) GetInfo() (name string, desc string) {
	return "example", "使用示例,没有作用"
}

func (d *DriverToolExample) Exit() {
	fmt.Println("bey bey")
	return
}

func (d *DriverToolExample) Run() (retCmd string, err error) {
	d.printWelcome()
	for {
		rawInput := util.ScanStdLine()
		switch rawInput {
		case "turn", "end", "exit": // 交到外部处理
			retCmd = rawInput
			return
		case "clear":
			util.ClearConsole()
			continue
		default:
			// write your logic here
		}
	}
	return "exit", nil
}

// ====================================

// 打印使用帮助
func (d *DriverToolExample) printWelcome() {
	color.HiRed("\n===========================\n==     行情数据小助手     ==\n===========================\n")
	color.Magenta("功能: overall - 全盘概述")
	color.Magenta("其他命令: end - 退出\nturn - 切换功能\nclear - 清空控制台\n")
}
