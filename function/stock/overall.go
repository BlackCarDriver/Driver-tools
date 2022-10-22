package stock

import (
	"bufio"
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"os"
	"strings"
	"time"
)

// 大盘涨跌比例监控
func startMonitorOverAll(breakSig <-chan string) {
	tr := time.NewTicker(time.Minute * 5)
	var input string
	util.ClearConsole()
	color.Blue("监控开始,输入任何东西结束~")
	printOverAll()
	for len(input) == 0 {
		select {
		case <-tr.C:
			if !isStocking() {
				break
			}
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
	up, hold, down, flow, prize := ret.F104, ret.F106, ret.F105, ret.F3, ret.F2
	total := float64(up + hold + down)
	maxLen := 60.0
	upLen := int(float64(up) / total * maxLen)

	downLen := int(float64(down) / total * maxLen)
	holdLen := int(maxLen) - upLen - downLen
	upStr := color.RedString("%s", strings.Repeat("#", upLen))
	holdStr := color.HiBlackString("%s", strings.Repeat("#", holdLen))
	downStr := color.GreenString("%s", strings.Repeat("#", downLen))
	flowStr := color.HiRedString("%.2f %.2f", prize, flow)
	if flow < 0 {
		flowStr = color.HiGreenString("%.2f %.2f%%", prize, flow)
	}
	fmt.Printf("%s %s%s%s %s\n", time.Now().Format("15:04"), upStr, holdStr, downStr, flowStr)
}

// 是否开盘时间
func isStocking() bool {
	now := time.Now().Format("15:04")
	if now > "09:25" && now < "11:35" {
		return true
	}
	if now > "11:55" && now < "15:05" {
		return true
	}
	return false
}

func scanStdLine() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}
