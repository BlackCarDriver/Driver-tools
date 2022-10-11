package stock

import (
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
)

type StockTool struct{}

func (d *StockTool) GetInfo() (name string, desc string) {
	return "stock", "股市行情小助手"
}

func (d *StockTool) Exit() {
	fmt.Println("bey bey")
	return
}

func (d *StockTool) Run() (retCmd string, err error) {
	fmt.Println("example 运行中,请输入结束命令:")
	cmd := util.ScanStdLine()
	return cmd, nil
}

// ====================================

// 打印使用帮助
func (d *StockTool) printWelcome() {
	color.Magenta("\n===========================\n==     截屏自动保存工具     ==\n===========================\n")
	color.Magenta("使用方法: 输入保存路径后,任务自动启动。 按'PrintScreen'键截图, 图片将自动保存到指定路径\n")
	color.Magenta("命令列表: \nend - 退出\nturn - 切换功能\nclear - 清空控制台\n")
}
