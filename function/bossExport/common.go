package bossExport

import (
	"git.yy.com/server/jiaoyou/go_projects/api/basedao/webdbdao"
	"git.yy.com/server/jiaoyou/go_projects/api/common/mgopoolclient"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"github.com/globalsign/mgo"
	"os"
	"time"
)

const timeFormat = "2006-01-02 15:04:05"
const dateFormat = "2006-01-02"

// 全局配置
var globalConfig struct {
	MongoURL string
	IsTest   bool
}

// 常用配置
var usualConfig struct {
	StartTime time.Time
	EndTime   time.Time
}

func initMongoClient() {
	sessionConfNew := mgopoolclient.MgoSessionConf{20, mgopoolclient.SetMode{mgo.Strong, true}}
	err := mgopoolclient.InitMgoSession(globalConfig.MongoURL, sessionConfNew)
	if err != nil {
		color.Red("failed to init mgo pool client: err=%v uir=%q", err, globalConfig.MongoURL)
		return
	}
	color.HiBlack("init mongo client success: uri=%q", globalConfig)
}

func initThriftClient() {
	webdbdao.InitS2SWebDB(globalConfig.IsTest)
	color.HiBlack("init thrift client success: isTest=%v", globalConfig.IsTest)
}

// 初始化全局配置
func initGlobalConfig() {
	for loop := true; loop; loop = false {
		color.HiBlue("boss数据导出工具,开始前请先按提示输入全局配置:")
		color.HiBlue("请先输入mongoURI:")
		globalConfig.MongoURL = util.ScanStdLine()
		color.HiBlue("请先输入当前环境: test 或 prod")
		input := util.ScanStdLine()
		globalConfig.IsTest = input == "test"

		color.HiBlack("获取配置完成, 全局配置=%+v", globalConfig)

		color.HiBlue("输入'ok' 测试连接, 输入其他重写配置: ")
		input = util.ScanStdLine()
		if input != "ok" {
			color.YellowString("重置配置~")
			continue
		}

		color.YellowString("测试连接~")

		initMongoClient()
		initThriftClient()

		color.HiBlue("输入 'ok' 确认配置, 输入其他重写配置: ")
		input = util.ScanStdLine()
		if input != "ok" {
			color.YellowString("重置配置~")
			continue
		}
	}
	color.YellowString("完成全局配置~")
}

func initUsualConfig() {
	for loop := true; loop; loop = false {
		color.HiBlue("请输入 startTime, 格式: 2006-01-02")
		input := util.ScanStdLine()
		usualConfig.StartTime, _ = time.ParseInLocation(dateFormat, input, time.Local)
		color.HiBlue("请输入 endTime， 格式: 2006-01-02")
		input = util.ScanStdLine()
		usualConfig.EndTime, _ = time.ParseInLocation(dateFormat, input, time.Local)

		color.HiBlack("常用配置=%+v", usualConfig)
		color.HiBlue("输入 'ok' 确认配置, 输入其他重写配置: ")
		input = util.ScanStdLine()
		if input != "ok" {
			color.YellowString("重置配置~")
			continue
		}
	}
	color.YellowString("完成常用配置~")
}

// 堵塞直到输入确定-不用改
func wait() {
	var input string
	times := 3
	for times > 0 {
		color.Yellow("输入 c 继续, 输入 e 中断程序:")
		input = util.ScanInput()
		times--
		if input == "c" || input == "C" {
			return
		}
		if input == "e" || input == "E" {
			break
		}
	}
	color.Red("bey bey~")
	os.Exit(1)
}
