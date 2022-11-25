package bossExport

import (
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
)

type MongoExport struct{}

func (m *MongoExport) GetInfo() (name string, desc string) {
	return "bossExport", "报表数据导出"
}

func (m *MongoExport) Exit() {
	fmt.Println("bey bey")
	return
}

func (m *MongoExport) Run() (retCmd string, err error) {
	initGlobalConfig()
	for {
		m.printWelcome()
		color.HiBlue("请输入命令:")
		rawInput := util.ScanStdLine()
		switch rawInput {
		case "turn", "end", "exit": // 交到外部处理
			retCmd = rawInput
			return
		case "clear":
			util.ClearConsole()
			continue
		case "config":
			initGlobalConfig()
			continue
		case "bossHistory":
			exportBossHistory()
		case "bossUserStat":
			exportBossUrlStat()
		case "bossUpdateLog":
			exportBossOpLog()
		case "bossMenuAuth":
			exportBossAuthItem()
		case "bossApprovalFlow":
			exportApprovalFlow()
		default:
			color.Yellow("未知命令 %q ,请重新输入：", rawInput)
		}
	}
	return "exit", nil
}

// ====================================

// 打印使用帮助
func (m *MongoExport) printWelcome() {
	color.HiRed("\n===========================\n==     报表导出工具     ==\n===========================\n")
	color.HiBlue("bossHistory - 导出BOSS页面请求记录 (数据量大一个月约15万行,注意查询范围)")
	color.HiBlue("bossUserStat - 统计接口请求频率 和 用户访问次数")
	color.HiBlue("bossUpdateLog - 导出更新记录")
	color.HiBlue("bossMenuAuth - 导出当前的页面权限列表")
	color.HiBlue("bossApprovalFlow - 审批流程数据导出")
	color.HiMagenta("其他命令: \nend - 退出\nturn - 切换功能\nclear - 清空控制台\nconfig - 重置全局配置")
}
