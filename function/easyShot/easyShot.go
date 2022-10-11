package easyShot

// 截屏自动保存工具

import (
	"crypto/md5"
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"github.com/fatih/color"
	"golang.design/x/clipboard"
	"strings"
	"time"
)

type EasyShoot struct {
	lastMd5Val  string // 上一个保存图片的md5值
	lastSize    int    // 上次保存图片的大小
	saveCounter int    // 总计保存多少个图片
	savePath    string // 保存文件的路径
}

func (d *EasyShoot) GetInfo() (name string, desc string) {
	return "shot", "截屏自动保存工具"
}

func (d *EasyShoot) Exit() {
	fmt.Println("bey bey")
	return
}

func (d *EasyShoot) Run() (retCmd string, err error) {
	for {
		util.ClearConsole()
		d.printWelcome()
		color.Yellow("请输入保存截屏的路径, 或其他可用命令:")
		var isDir bool
		input := util.ScanStdLine()
		if input == "end" || input == "exit" || input == "turn" { // 交到外层处理
			return input, nil
		}
		if input == "clear" {
			util.ClearConsole()
			continue
		}
		input = strings.TrimRight(input, "/")
		isDir, err = util.CheckDirIsExist(input)
		if !isDir {
			color.Red("打开路径失败,请重新输入: path=%q  err=%v\n", input, err)
			continue
		}

		d.savePath = input
		stopSignal := make(chan string, 1) // 向stopSignal写入任何字符, 监听任务都会结束
		go d.startClipBoardMonitor(stopSignal)
		color.White("监听任务运行中,输入任何东西可结束任务 \n")
		input = util.ScanStdLine()
		stopSignal <- input

		time.Sleep(time.Second * 3)
	}

	return "", nil
}

// ====================================

// 打印使用帮助
func (d *EasyShoot) printWelcome() {
	color.Magenta("\n===========================\n==     截屏自动保存工具     ==\n===========================\n")
	color.Magenta("使用方法: 输入保存路径后,任务自动启动。 按'PrintScreen'键截图, 图片将自动保存到指定路径\n")
	color.Magenta("命令列表: \nend - 退出\nturn - 切换功能\nclear - 清空控制台\n")
}

// startClipBoardMonitor 定时每秒从剪切板查看图片数据, 如果发现出现新截图,则保存到指定路径
func (d *EasyShoot) startClipBoardMonitor(stopSignal <-chan string) {
	trigger := time.NewTicker(time.Second)
	stop := false
	for !stop {
		select {
		case <-stopSignal:
			color.Yellow("监听终止, 已保存截图数量: %d \n", d.saveCounter)
			stop = true
		case <-trigger.C:
			d.tryWatchClipBoard()
		}
	}
	color.Yellow("任务结束")
}

// 获取剪切板图片并保存的逻辑
func (d *EasyShoot) tryWatchClipBoard() (err error) {
	content := clipboard.Read(clipboard.FmtImage)
	if len(content) == 0 { // 剪切板的内容非图片
		return
	}
	if len(content) == d.lastSize { // 大小没变,可判断没变更
		return
	}
	currentMD5 := getMD5Value(content) // 剪切板没变更
	if currentMD5 == d.lastMd5Val {
		return
	}

	// 发现新图片，保存到指定路径
	filePath := fmt.Sprintf("%s/%d.png", d.savePath, time.Now().Unix())
	err = util.SaveDataToFile(content, filePath)
	if err != nil {
		color.Red("保存截图失败: path=%s err=%v \n", filePath, err)
		return
	}

	d.lastSize = len(content)
	d.lastMd5Val = currentMD5
	d.saveCounter++
	color.Green("保存截图成功: path=%s size=%d \n", filePath, len(content))
	return
}

// 求md5值
func getMD5Value(byteAry []byte) string {
	md5Encoder := md5.New()
	md5Encoder.Write(byteAry)
	md5Value := fmt.Sprintf("%x", md5Encoder.Sum(nil))
	return md5Value
}
