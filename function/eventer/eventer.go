package eventer

// 事件记录器

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"github.com/astaxie/beego/logs"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// EventLog 事件记录器
type EventLog struct {
	target      *os.File       // 日志文件读写对象
	config      eventLogConfig // 当前的配置
	isFirstTime bool           // 是否第一次使用
}

func (e *EventLog) GetInfo() (name string, desc string) {
	return "event", "事件记录工具"
}

func (e *EventLog) Run() (retCmd string, err error) {
	err = e.initEventLog()
	if err != nil {
		color.Red("初始化eventLog失败\n: err=%v", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		color.Yellow("Input event or command > ")
		var input string
		input, err = reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("Bufio ReadString fail!: %s \r\n", err)
		}
		input = strings.TrimSpace(input)
		switch input {
		case "":
			continue
		case "end", "exit", "turn": // 通用指令交到外层处理
			return input, nil
		case "clear":
			if err = util.ClearConsole(); err != nil {
				logs.Error(err)
			}
		case "show": // 打印当前月份日志
			if err = printEventLogFile(e.config.WritePath); err != nil {
				logs.Error("Printf it month event fail: %v \r\n", err)
			}
		case "ls": // 打印日志列表
			if err = printfEventLogsList(e.config.WritePath); err != nil {
				logs.Error("Printf list fail: %v \r\n", err)
			}
		case "his": // 打印指定月份的日志
			color.White("Input which file you want to see > ")
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)
			reg, _ := regexp.Compile(`^\d+-\d+$`)
			if !reg.MatchString(input) {
				color.White("File name not right, should like '2019-8'!\n")
				continue
			}
			if err = printHistoryEventLog(input, e.config.WritePath); err != nil {
				fmt.Println(err)
				logs.Error(err)
			}
		default: // 默认是记录日志
			now := time.Now()
			event := fmt.Sprintf("\r\n( %02d:%02d ) - - - - - - - - - - - - - - - %s\r\n", now.Hour(), now.Minute(), input)
			_, err = e.target.WriteString(event)
			if err != nil {
				logs.Error("Write string to target file fail! : %v \r\n", err)
				color.White("Record event fail: %v", err)
			} else {
				e.config.TodayTimes++
				color.White("Save scuess!\n")
			}
		}
	}
}

func (e *EventLog) Exit() {
	e.saveState()
	return
}

// ==============================

func (e *EventLog) initEventLog() (err error) {
	// 初始化logs
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	logs.SetLogger(logs.AdapterFile, `{"filename":"./logs/eventer.log"}`)

	// 读取配置
	file, err := os.Open(configPath)
	if err != nil { // 读取配置文件失败
		absConPath, _ := filepath.Abs(configPath)
		logs.Warn("eventLog read config file fail: path=%s err=%v", absConPath, err)
		fmt.Println("Init data not fond, Create a new noe? (yes/no) ")
		var input string
		fmt.Scanf("%s\n", &input)
		if input = strings.ToLower(input); input[0] == 'y' { // 初始化配置文件
			dfDataByte, _ := json.Marshal(defaultConfig)
			err = ioutil.WriteFile(configPath, dfDataByte, 0644)
			if err != nil {
				return fmt.Errorf("write init data to eventer config file fall: %v", err)
			}
			fmt.Println("Init data success!")
			file.Close()
			file, err = os.Open(configPath)
			if err != nil {
				return fmt.Errorf("open config file Fail after init a new config")
			}
		} else { // 输入"no"
			return fmt.Errorf("inti data not found")
		}
	}
	// 已成功打开配置文件
	defer file.Close()
	buf := bufio.NewReader(file)
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return fmt.Errorf("eventer read config file fall: %v", err)
	}
	// 成功读取文件
	err = json.Unmarshal(bytes, &e.config)
	if err != nil {
		return fmt.Errorf("unmarshal eventer config fail: %v", err)
	}

	// 检查文件日志文件是否存在
	_, err = os.Stat(e.config.WritePath)
	if err != nil {
		logs.Error("Open write file fail : %v", err)
		_, err := os.Create(e.config.WritePath)
		if err != nil {
			return fmt.Errorf("create write file fail: %v", err)
		} else {
			fmt.Println("Already create a new eventLog file!")
			e.isFirstTime = true
		}
	}

	// 进入新的一月,切换文件
	if e.config.LastTime.Month() != time.Now().Month() {
		absWPath, err := filepath.Abs(e.config.WritePath)
		if err != nil {
			logs.Error("Get absolute path of write-path fail:%v", err)
		}
		dir := filepath.Dir(absWPath)
		fileName := fmt.Sprintf("%s/%04d-%d.txt", dir, e.config.LastTime.Year(), e.config.LastTime.Month())
		err = os.Rename(e.config.WritePath, fileName)
		if err != nil {
			logs.Error("Rename %s to %s fail: %v", e.config.WritePath, fileName, err)
		} else {
			color.Yellow("\n============== Happy Good Month! ==============\n\n")
			_, err = os.Create(e.config.WritePath)
			if err != nil {
				logs.Error("Create new file fail after rename old file: %v", err)
			}
		}
	}

	// 追加模式打开文件
	e.target, err = os.OpenFile(e.config.WritePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open target file fail after guarantee it is exist!: %v", err)
	}

	// 当天第一次打开显示相关信息
	if time.Now().Day() != e.config.LastTime.Day() || e.isFirstTime {
		color.Magenta("Have A Good Day!\n")
		event := fmt.Sprintf("\n\n\n===============================[ %s ]===============================\n\n\n", time.Now().Format("01-02 Mon"))
		_, err = e.target.WriteString(event)
		e.config.TodayTimes = 0
	}

	// 初始化完成, 打印使用帮助
	e.printWelcome()
	return nil
}

// 打印帮助信息
func (e *EventLog) printWelcome() {
	duration := time.Since(e.config.LastTime)
	util.ClearConsole()
	color.White("======================\n==   事件记录器     ==\n======================\n")
	color.White("命令列表:\nshow - 展示本月日志\nclear - 清空控制台\nend - 出程序\nturn - 切换功能\nhis - 查看过往日志\nls - 展示日志列表\n")
	color.White("上次记录时间距今: %d hour %d minute \n", int(duration.Hours())%24, int(duration.Minutes())%60)
	color.White("今日日志数量:    %d \n", e.config.TodayTimes)
}

// 更新配置文件
func (e *EventLog) saveState() {
	e.config.LastTime = time.Now()
	after, _ := json.Marshal(e.config)
	err := ioutil.WriteFile(configPath, after, os.ModeSetuid)
	if err != nil {
		logs.Error("save lastest data to eventer config file fall: %v", err)
	} else {
		fmt.Println("eventer save state success!")
	}
}

// 打印时间日志文件
func printEventLogFile(filePath string) error {
	absPath, _ := filepath.Abs(filePath)
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("open file fail when printf the eventer logs: %v", err)
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString(byte('\n'))
		if err == io.EOF { //end of file
			break
		}
		if err != nil {
			return fmt.Errorf("error happen when printf the eventer data: %v", err)
		}
		logsReg, _ := regexp.Compile(`^\([\d\: ]+\)( -){10,} .+\s$`)
		if logsReg.MatchString(line) {
			color.Yellow(line[:9])
			color.White(line[9:39])
			color.Yellow(line[39:])
		} else {
			fmt.Print(line)
		}
	}
	return nil
}

// 展示日志文件列表
func printfEventLogsList(path string) error {
	absWPath, _ := filepath.Abs(path)
	historyDir := filepath.Dir(absWPath)
	file, err := os.Open(historyDir)
	if err != nil {
		return fmt.Errorf("open history directory fail: %v", err)
	}
	defer file.Close()
	fi, err := file.Readdir(0)
	if err != nil {
		return fmt.Errorf("read history directory files list fail: %v", err)
	}
	color.White("=========== History ===========\n")
	for _, info := range fi {
		color.White(info.Name())
		fmt.Println()
	}
	fmt.Println()
	return nil
}

// 打印特定月份的日志
func printHistoryEventLog(name string, writePath string) error {
	absWPath, _ := filepath.Abs(writePath)
	historyDir := filepath.Dir(absWPath)
	path := fmt.Sprintf("%s/%s.txt", historyDir, name)
	err := printEventLogFile(path)
	return err
}
