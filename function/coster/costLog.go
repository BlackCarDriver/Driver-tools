package coster

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"github.com/astaxie/beego/logs"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CostLog struct {
	config *costLogConfig
	target *os.File // 消费数据存储文件的读写入口
}

func (c *CostLog) GetInfo() (name string, desc string) {
	return "cost", "记账工具"
}

func (c *CostLog) Run() (retCmd string, err error) {
	err = c.initCostLog()
	if err != nil {
		color.Red("init fail: err=%v\n", err)
		return
	}

	for {
		color.Cyan("Input materials or command > ")
		input := util.ScanInput()
		input = strings.TrimSpace(input)
		color.Red(input)

		switch input {
		case "":
			continue
		case "end", "exit", "turn":
			return input, nil
		case "clear":
			if err = util.ClearConsole(); err != nil {
				log.Error("%v", err)
			}
		case "show":
			if err = printfCost(c.config.WritePath); err != nil {
				log.Error("Printf it month cost record fail: %v \r\n", err)
			}
		case "ls":
			if err = printfList(c.config.WritePath); err != nil {
				log.Error("Printf list fail: %v \r\n", err)
			}
		case "his":
			color.Cyan("Input which file you want to see > ")
			input = util.ScanInput()
			reg, _ := regexp.Compile(`^\d+-\d+$`)
			if !reg.MatchString(input) {
				color.Red("File name not right, should like '2019-8'!\n")
				continue
			}
			absWPath, _ := filepath.Abs(c.config.WritePath)
			historyDir := filepath.Dir(absWPath)
			path := fmt.Sprintf("%s/%s.txt", historyDir, input)
			err = printfCost(path)
			if err != nil {
				fmt.Println("show history fail: ", err)
			}

		default:
			price := 0.0
			now := time.Now()
			color.Cyan("Input how many it cost: > ")
			for {
				cost := util.ScanInput()
				color.Red(cost)
				if strings.TrimSpace(cost) == "" {
					continue
				} else if price, err = strconv.ParseFloat(cost, 64); err != nil {
					fmt.Println("input money is unexpect, please try again")
					continue
				}
				break
			}
			if err != nil {
				log.Warn("%v", err)
				fmt.Println(err)
				continue
			}
			record := fmt.Sprintf("\r\n时间:[ %02d-%02d ] - - - - - - - - - - - - - 物品: %-17s  金额:  %.1f \r\n",
				now.Month(), now.Day(), input, price)
			_, err = c.target.WriteString(record)
			if err != nil {
				log.Error("Write string to target file fail! : %v \r\n", err)
				color.Red("Record event fail: %v", err)
			} else {
				c.config.MonthCost += price
				c.config.TotalCost += price
				color.Green("save success!\n")
			}
		}
	}
}

func (c *CostLog) Exit() {
	c.saveState()
	return
}

// ===================

// 初始化
func (c *CostLog) initCostLog() (err error) {
	// log初始化
	log = logs.NewLogger()
	log.EnableFuncCallDepth(true)
	log.SetLogFuncCallDepth(3)
	err = log.SetLogger(logs.AdapterFile, `{"filename":"./logs/coster.log"}`)
	if err != nil {
		fmt.Printf("Create log file fail: %v", err)
	}

	// 读取配置文件
	file, err := os.Open(configPath)
	if err != nil {
		absConPath, _ := filepath.Abs(configPath)
		log.Warn("costLog read config file fall: path=%s  err=%v", absConPath, err)
		fmt.Println("Init data not fond, Create a new noe? (yes/no) ")
		var input string
		fmt.Scanf("%s\n", &input)
		if input = strings.ToLower(input); input[0] == 'y' { // 创建配置文件
			dfDataByte, _ := json.Marshal(dfData)
			err = ioutil.WriteFile(configPath, dfDataByte, 0644)
			if err != nil {
				err = fmt.Errorf("write init data to coster config file fall: err=%v", err)
				log.Error("%v", err)
				return err
			}
			fmt.Println("Init data success!")
			file.Close()
			file, err = os.Open(configPath)
			if err != nil {
				err = fmt.Errorf("open config file Fail after init a new config")
				log.Error("%v", err)
				return err
			}
		} else { // 输入"no"
			err = fmt.Errorf("init data not found")
			log.Error("%v", err)
			return err
		}
	}
	// 成功打开配置文件
	defer file.Close()
	buf := bufio.NewReader(file)
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		err = fmt.Errorf("coster read config file fall: %v", err)
		log.Error("%v", err)
		return err
	}
	// 解析配置
	err = json.Unmarshal(bytes, &c.config)
	if err != nil {
		err = fmt.Errorf("unmarshal coster config fail: %v", err)
		log.Error("%v", err)
		return err
	}

	// 打开日志文件，没的话初始化一个
	_, err = os.Stat(c.config.WritePath)
	if err != nil {
		log.Error("open write file fail : %v", err)
		_, err = os.Create(c.config.WritePath)
		if err != nil {
			err = fmt.Errorf("create write file fail: %v", err)
			log.Error("%v", err)
			return err
		} else {
			fmt.Println("Already create a new costLog file!")
		}
	}

	// 打开文件
	c.target, err = os.OpenFile(c.config.WritePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		err = fmt.Errorf("open target file fail after guarantee it is exist!: %v", err)
		log.Error("%v", err)
		return err
	}

	// 新的月份开始时，重命名文件
	if c.config.LastTime.Month() != time.Now().Month() {
		tailStamp := fmt.Sprintf("\n statics :  [ Month: %s ]   [ MonthlyCost: %.1f ]   [ TotallyCost: %.1f ] \n",
			c.config.LastTime.Month(), c.config.MonthCost, c.config.TotalCost)
		_, err = c.target.WriteString(tailStamp)
		color.Green(tailStamp)
		if err != nil {
			err = fmt.Errorf("write tailStamp to coster target file fail: %v", err)
			log.Error("%v", err)
			return err
		}
		var absWPath string
		absWPath, err = filepath.Abs(c.config.WritePath)
		if err != nil {
			log.Error("Get absolute path of write-path fail: %v", err)
		}
		dir := filepath.Dir(absWPath)
		c.target.Close()
		fileName := fmt.Sprintf("%s/%04d-%d.txt", dir, c.config.LastTime.Year(), c.config.LastTime.Month())
		err = os.Rename(c.config.WritePath, fileName)
		if err != nil {
			log.Error("Rename %s to %s fail: %v", c.config.WritePath, fileName, err)
		} else {
			color.Yellow("============== Happy New Month! ==============\n")
			_, err = os.Create(c.config.WritePath)
			if err != nil {
				log.Error("Create new file fail after rename old file: %v", err)
			}
		}
		c.target, err = os.OpenFile(c.config.WritePath, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Error("Open target file fail after rename: %v", err)
			return err
		}
		c.config.MonthCost = 0.0
		c.saveState()
	}
	c.printWelcome()
	return
}

// 打印使用帮助
func (c *CostLog) printWelcome() {
	duration := time.Since(c.config.LastTime)
	color.Yellow("\n=====================\n==     记账工具     ==\n=====================\n")
	color.Yellow("命令列表:\nshow - 展示本月日志\nclear - 清空控制台\nend - 退出\nturn - 切换功能\nhis - 查看指定月份的日志\nls - 查看历史日志文件列表\n")
	color.Cyan("距离上次记账已过:  %d hour %d minute \n", int(duration.Hours())%24, int(duration.Minutes())%60)
	color.Cyan("本月已消费:    %.1f \n", c.config.MonthCost)
	color.Cyan("至今已消费:    %.1f  \n", c.config.TotalCost)
}

// 保持状态到配置文件
func (c *CostLog) saveState() {
	c.config.LastTime = time.Now()
	latestConfig, _ := json.Marshal(c.config)
	err := ioutil.WriteFile(configPath, latestConfig, os.ModeSetuid)
	if err != nil {
		log.Error("save latest data to config file fall: err=%v", err)
	} else {
		fmt.Println("costLog save state success!")
	}
}

// 打印消费日志
func printfCost(filePath string) (err error) {
	absPath, _ := filepath.Abs(filePath)
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("open file fail when print the costLog: err=%v", err)
		return
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	for {
		var line string
		line, err = buf.ReadString(byte('\n'))
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error happen when printf the coster log: err=%v", err)
		}
		if strings.HasPrefix(line, "时间") {
			format := "时间:[ %s ] - - - - - - - - - - - - - 物品: %s             金额:  %f  "
			var date string
			var object string
			var money float64
			fmt.Sscanf(line, format, &date, &object, &money)
			color.Cyan("时间:[")
			color.Green(" %s ", date)
			color.Cyan("] - - - - - - - - - - - - - 物品:")
			color.Green(" %s\t\t\t", object)
			color.Cyan("金额：")
			color.Green("%.1f\n", money)
		} else {
			fmt.Print(line)
		}
	}
	return nil
}

// 展示历史消费日志文件
func printfList(path string) error {
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
		color.Yellow(info.Name())
		fmt.Println()
	}
	fmt.Println()
	return nil
}
