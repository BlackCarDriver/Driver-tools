package flowCounter

import (
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"regexp"
	"strconv"
	"strings"
)

type FlowCounter struct{}

func (f *FlowCounter) GetInfo() (name string, desc string) {
	return "flow", "价格浮动计算器"
}

func (f *FlowCounter) Run() (retCmd string, err error) {
	f.printHelp()
	for {
		rawInput := util.ScanStdLine()
		switch rawInput {
		case "turn", "end", "exit": // 交到外部处理
			return
		case "clear":
			util.ClearConsole()
			continue
		default:
			args := regexp.MustCompile("\\s+").Split(rawInput, -1)
			if args[0] == "rise" && len(args) == 3 {
				f1, _ := strconv.ParseFloat(args[1], 64)
				f2, _ := strconv.ParseFloat(args[2], 64)
				printRiseResult(f1, f2)
				break
			}
			if args[0] == "flow" && (len(args) == 3 || len(args) == 4) {
				currentPrize, _ := strconv.ParseFloat(args[1], 64)
				increase, _ := strconv.ParseFloat(strings.Trim(args[2], "%"), 64)
				step := int64(10)
				if len(args) > 3 {
					step, _ = strconv.ParseInt(args[3], 10, 64)
				}
				printFlowResult(currentPrize, increase, step)
				break
			}
			util.ColorPrintf(util.ColorRed, "错误的命令: %q\n", rawInput)
		}
	}
	return "exit", nil
}

func (f *FlowCounter) Exit() {
	fmt.Println("bey bey")
	return
}

// ==============================

// 打印使用帮助
func (f *FlowCounter) printHelp() {
	util.ColorPrintf(util.ColorLightCyan, "\n============================\n==     价格浮动计算器     ==\n============================\n")
	util.ColorPrintf(util.ColorLightCyan, "命令列表:  rise-涨跌幅计算 flow-梯度计算 turn-切换功能 end-退出程序 clear-清空控制台 \n")
	util.ColorPrintf(util.ColorLightCyan, "rise 用法： rise $价格1 $价格2")
	util.ColorPrintf(util.ColorGray, "     # 如 rise 100 101 , 计算100->101的涨幅\n")
	util.ColorPrintf(util.ColorLightCyan, "flow 用法： flow $当前价格 $每档浮动 $档数(默认10)")
	util.ColorPrintf(util.ColorGray, "     # 如 flow 100 1% 10, 生成涨跌10档价格\n")
}

// 计算浮动
func printRiseResult(before, after float64) {
	util.ColorPrintf(util.ColorGray, "原价: %.3f\n现价: %.3f\n", before, after)
	flow := countRiseRange(before, after)
	if after >= before {
		util.ColorPrintf(util.ColorLightRed, "涨幅:  +%.3f%% \n", flow)
	} else {
		util.ColorPrintf(util.ColorLightGreen, "涨幅: %.3f%% \n", flow)
	}
}

// 计算梯度
func printFlowResult(current float64, flow float64, step int64) {
	if flow < 0 {
		util.ColorPrintln(util.ColorRed, "$每档浮动 必须大于0")
		return
	}
	if step <= 0 || step > 1000000 {
		util.ColorPrintln(util.ColorRed, "$d档数 必须大于0")
		return
	}
	util.ColorPrintf(util.ColorGray, "当前价=%.3f   每档浮动=%.3f  档位=%d \n", current, flow, step)

	var upList, downList []float64

	// 涨10档
	tmp := current
	for i := 0; i < int(step); i++ {
		tmp = tmp * ((100.0 + flow) / 100.0)
		upList = append(upList, tmp)
	}
	// 跌10档
	tmp = current
	for i := 0; i < int(step); i++ {
		tmp = tmp * ((100.0 - flow) / 100.0)
		downList = append(downList, tmp)
	}

	// 打印
	util.ColorPrintln(util.ColorWhite, "档位 ---- 价格 ----------- 涨幅 --")
	for i := len(upList) - 1; i >= 0; i-- {
		r := countRiseRange(current, upList[i])
		util.ColorPrintf(util.ColorRed, "+%d  \t  %.3f \t +%.3f%%\n", i+1, upList[i], r)
	}
	util.ColorPrintf(util.ColorWhite, "0  \t  %.3f\n \t 0%%", current)
	for i := 0; i < len(downList); i++ {
		r := countRiseRange(current, downList[i])
		util.ColorPrintf(util.ColorGreen, "-%d  \t  %.3f \t %.3f%%\n", i+1, downList[i], r)
	}
}

// 百分比涨幅计算
func countRiseRange(before, after float64) (rise float64) {
	if before == 0 {
		return 0
	}
	rise = (after - before) / before
	return rise * 100.0
}
