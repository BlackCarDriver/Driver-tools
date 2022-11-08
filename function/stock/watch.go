package stock

import (
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"strings"
	"time"
)

type watchInfo struct {
	TAG     string
	XJ      float64 // 现价
	ZDF     float64 // 涨跌幅
	HSL     float64 // 换手率
	ZF      float64 // 振幅
	Above   float64 // 卖出涨幅
	Lower   float64 // 买入跌幅
	Offset  float64 // 对比上次交易的涨跌幅
	Note    string  // 备忘录
	Decimal int
}

func (w watchInfo) GenDesc() (desc string, header string) {
	header = " 名称      现价     涨跌幅    换手率    振幅   触发卖   触发买   偏移值   备忘录"
	stdFormat := fmt.Sprintf("%%.%df", w.Decimal)
	items := []string{
		color.HiBlackString(w.TAG),
		color.HiBlackString(stdFormat, w.XJ),
		colorFloat(w.ZDF, 2, "%"),
		color.HiBlackString("%.2f%%", w.HSL),
		color.HiBlackString("%.2f%%", w.ZF),
		colorFloat(w.Above, 2, "%"),
		colorFloat(w.Lower, 2, "%"),
		colorFloat(w.Offset, 2, "%"),
		color.HiBlackString(w.Note),
	}
	return strings.Join(items, " "), header
}

// 个股盯盘
func startMonitorStock(breakSig <-chan string) {
	tr := time.NewTicker(time.Minute * 5)
	var input string
	printWatch()
	for len(input) == 0 {
		select {
		case <-tr.C:
			if !isStocking() {
				break
			}
			printWatch()
		case input = <-breakSig:
			break
		}
	}
	color.Blue("监控结束~")
}

func printWatch() {
	targets := stockConfig.MyStock
	if len(targets) == 0 {
		color.Red("no stock found in config: %+v", stockConfig)
		return
	}
	util.ClearConsole()
	color.HiBlack("更新时间: %s \n", time.Now().Format("01-02 15:04"))
	printer := util.NewColumnPrinter(3)
	for i, item := range targets {
		info, err := getWatchInfo(item)
		if err != nil {
			color.HiBlack("get data fail: code=%s err=%v", item.Code, err)
			continue
		}
		desc, header := info.GenDesc()
		if i == 0 {
			color.HiBlack(header)
		}
		printer.Write(desc)
	}
	printer.Print()
}

// 涨跌相关数据上色
func colorFloat(before float64, decimal int, suffix string) (after string) {
	format := fmt.Sprintf("%%.%df", decimal)
	str := fmt.Sprintf(format, before) + suffix
	if before == 0 {
		after = color.HiBlackString(str)
	} else if before > 0 {
		after = color.RedString(str)
	} else {
		after = color.GreenString(str)
	}
	return after
}

// 整数转浮点数
func parsePrize(before int64, decimal int) (after float64) {
	if decimal <= 0 || decimal > 3 {
		color.Red("unexpect decimal: %d", decimal)
		decimal = 1
	}
	base := 1.0
	for i := 0; i < decimal; i++ {
		base = base * 10.0
	}
	return float64(before) / base
}

// 生成个股最新监控信息
func getWatchInfo(target monitorCfg) (info watchInfo, err error) {
	data, err := getStockData(target.Mode, target.Code)
	if err != nil {
		return
	}
	info = parseWatchInfo(data, target)
	return
}

// 数据转换
func parseWatchInfo(before *GetStockRespData, target monitorCfg) (after watchInfo) {
	if before == nil {
		return
	}
	currentPrize := parsePrize(before.F43, before.F59)
	after = watchInfo{
		TAG:     target.Tag,
		Note:    target.Note,
		XJ:      currentPrize,
		Decimal: before.F59,
		HSL:     parsePrize(before.F168, 2),
		ZF:      parsePrize(before.F171, 2),
		ZDF:     parsePrize(before.F170, 2),
	}
	if len(after.Note) == 0 {
		after.Note = "-"
	}
	if target.SellPrize > 0 {
		after.Above = countRiseRange(currentPrize, target.SellPrize)
	}
	if target.BuyPrize > 0 {
		after.Lower = countRiseRange(currentPrize, target.BuyPrize)
	}
	if target.LastDeal > 0 {
		after.Offset = countRiseRange(target.LastDeal, currentPrize)
	}
	return after
}
