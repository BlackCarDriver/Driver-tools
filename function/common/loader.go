package common

import (
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
)

type FuncLoader struct {
	HelpDesc string
	FuncMap  map[string]DriverToolFunction
}

func NewFuncLoader() (loader FuncLoader) {
	return FuncLoader{
		HelpDesc: "=========== 功能列表 ==========",
		FuncMap:  make(map[string]DriverToolFunction),
	}
}

// AddFunc 注册功能
func (ld *FuncLoader) AddFunc(item DriverToolFunction) {
	if ld.FuncMap == nil {
		ld.FuncMap = map[string]DriverToolFunction{}
	}
	name, desc := item.GetInfo()
	ld.FuncMap[name] = item
	ld.HelpDesc = fmt.Sprintf("%s\n%s -- %s", ld.HelpDesc, name, desc)
}

// Run 加载功能
func (ld *FuncLoader) Run(fName string) (err error) {
	for {
		selectFunc, exist := ld.FuncMap[fName]
		if !exist {
			color.Red("No such function: %s\n", fName)
			break
		}
		var exitCmd string // 上个程序的结束返回
		exitCmd, err = selectFunc.Run()
		selectFunc.Exit()
		if err != nil {
			fmt.Errorf("end with error: %v", err)
			break
		}

		// 主动切换到下个功能
		selectFunc, exist = ld.FuncMap[exitCmd]
		if exist {
			continue
		}

		// 打印帮助并切换功能
		if exitCmd == "turn" {
			util.ClearConsole()
			fName = ld.mustGetNextFunc()
			if fName != "end" && fName != "exit" {
				continue
			}
		}
		break
	}

	return
}

// 打印全局使用帮助
func (ld *FuncLoader) printHelp() {
	color.HiCyan(ld.HelpDesc)
}

// 堵塞, 从控制台获取功能名输入,直到得到存在的功能名或end
func (ld *FuncLoader) mustGetNextFunc() (validCmd string) {
	ld.printHelp()
	color.HiMagenta("请输入功能名, 或输入 end 退出")
	for {
		validCmd = util.ScanStdLine()
		if validCmd == "end" || validCmd == "exit" {
			break
		}
		_, exist := ld.FuncMap[validCmd]
		if exist {
			break
		}
		color.Red("请输入存在的功能名")
	}
	return validCmd
}
