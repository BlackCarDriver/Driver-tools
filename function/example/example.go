package example

import (
	"fmt"
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
	fmt.Println("example 运行中,请输入结束命令:")
	cmd := util.ScanStdLine()
	return cmd, nil
}
