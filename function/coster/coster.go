package coster

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	c "../common"

	"github.com/astaxie/beego/logs"
)

const configPath = "./config/coster.json"

var (
	l      *log.Logger
	data   *costerData //latest data
	target *os.File    //the file write and read cost logs from
)

type costerData struct {
	LastTime  time.Time //lasttime of using it tool
	MonthCost float64   //how many money have cost it month
	TotalCost float64   //how many money totaly cost
	LogsPath  string
	WritePath string
}

var dfData = costerData{ //default data
	LastTime:  time.Now(),
	MonthCost: 0.0,
	TotalCost: 0.0,
	WritePath: "./data/coster/coster.txt",
}

func costerInit() error {
	//setting up logger
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	logs.SetLogger(logs.AdapterFile, `{"name":"./logs/coster.log"}`)
	//read config file
	file, err := os.Open(configPath)
	if err != nil { //read config file fail
		absConPath, _ := filepath.Abs(configPath)
		logs.Warn("Coster read config file '%s' fall: %v", absConPath, err)
		fmt.Println("Init data not fond, Create a new noe? (yes/no) ")
		var input string
		fmt.Scanf("%s\n", &input)
		if input = strings.ToLower(input); input[0] == 'y' { //create a init data on config file
			dfDataByte, _ := json.Marshal(dfData)
			err := ioutil.WriteFile(configPath, dfDataByte, 0644)
			if err != nil { //can not create a config file
				err = fmt.Errorf("Write init data to coster config file fall: %v", err)
				logs.Error(err)
				return err
			}
			fmt.Println("Init data scuess!")
			file.Close() //reopen config file
			file, err = os.Open(configPath)
			if err != nil { //can not read config file after create new one
				err = fmt.Errorf("Open config file Fail after init a new config!")
				logs.Error(err)
				return err
			}
		} else { //user input no
			err = fmt.Errorf("Init data not found")
			logs.Error(err)
			return err
		}
	}
	//open config scuess and going to read config
	defer file.Close()
	buf := bufio.NewReader(file)
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		err = fmt.Errorf("coster read config file fall: %v", err)
		logs.Error(err)
		return err
	}
	//read config scuess and going to load config
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		err = fmt.Errorf("Unmarshal coster config fail: %v", err)
		logs.Error(err)
		return err
	}
	//config load scuess
	//check target file exist, create a new one if not exist
	_, err = os.Stat(data.WritePath)
	if err != nil {
		logs.Error("Open write file fail : %v", err)
		_, err := os.Create(data.WritePath)
		if err != nil {
			err = fmt.Errorf("Create write file fail: %v", err)
			logs.Error(err)
			return err
		} else {
			fmt.Println("Already create a new enverter file!")
		}
	}
	//open target file with write to end model
	target, err = os.OpenFile(data.WritePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		err = fmt.Errorf("Open target file fail after guarantee it is exist!: %v", err)
		logs.Error(err)
		return err
	}
	//write to another file if start a new month
	if data.LastTime.Month() != time.Now().Month() {
		//record the monthly statistics at the end of target file
		tailStamp := fmt.Sprintf("\n============[Month:%d]=================[MonthlyCost:%.1f]=================[TotallyCost:%.1f]==============\n",
			data.LastTime.Month(), data.MonthCost, data.TotalCost)
		_, err = target.WriteString(tailStamp)
		if err != nil {
			err = fmt.Errorf("Write tailStamp to coster target file fail: %v", err)
			logs.Error(err)
			return err
		}
		data.MonthCost = 0.0
		saveState()
		//save lastmonth history to another file
		absWPath, err := filepath.Abs(data.WritePath)
		if err != nil {
			logs.Error("Get absolute path of write-path fail: %v", err)
		}
		dir := filepath.Dir(absWPath)
		fileName := fmt.Sprintf("%s/%04d-%d.txt", dir, data.LastTime.Year(), data.LastTime.Month())
		err = os.Rename(data.WritePath, fileName)
		if err != nil {
			logs.Error("Rename %s to %s fail: %v", data.WritePath, fileName, err)
		} else {
			c.ColorPrint(c.Light_yellow, "============== Happy New Month! ==============\n")
			_, err = os.Create(data.WritePath)
			if err != nil {
				logs.Error("Create new file fail after rename old file: %v", err)
			}
		}
	}
	printWelcome()
	return nil
}

func Run(taskBus chan<- func()) (status int, err error) {
	err = costerInit()
	if err != nil {
		return c.ErrorExit, err
	}
	taskBus <- saveState
	defer target.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		c.ColorPrint(c.Light_blue, "Input materials or command > ")
		input := ""
		fmt.Scanf("%s\n", &input)
		input = strings.TrimSpace(input)
		switch input {
		case "":
			continue
		case "end", "exit":
			return c.NormalReturn, nil
		case "clear":
			if err = c.ClearConsole(); err != nil {
				logs.Error(err)
			}
		case "turn":
			tcase := c.GetTurnCode()
			if tcase != c.NotFound {
				return tcase, nil
			}
		case "show":
			if err = printfCost(data.WritePath); err != nil {
				logs.Error("Printf it month cost record fail: %v \r\n", err)
			}
		case "ls":
			if err = printfList(); err != nil {
				logs.Error("Printf list fail: %v \r\n", err)
			}
		case "key":
			tcase := c.GetKeyCode()
			if tcase != c.NotFound {
				return tcase, nil
			}
		case "history":
			c.ColorPrint(c.Light_cyan, "Input which file you want to see > ")
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)
			reg, err := regexp.Compile(`^\d+-\d+$`)
			if !reg.MatchString(input) {
				c.ColorPrint(c.Light_red, "File name not right, should like '2019-8'!\n")
				continue
			}
			if err = printfCost(input); err != nil {
				fmt.Println(err)
				logs.Error(err)
			}
		default:
			now := time.Now()
			c.ColorPrint(c.Light_blue, "Input how many it cost: > ")
			price := 0.0
			_, err = fmt.Scanf("%f\n", &price)
			if err != nil {
				logs.Warn(err)
				fmt.Println(err)
				continue
			}
			fmt.Println(price)
			record := fmt.Sprintf("\r\n( %02d-%02d ) - - - - - - - - - - - - - - - %-15s %.1f  \r\n",
				now.Month(), now.Day(), input, price)
			_, err = target.WriteString(record)
			if err != nil {
				logs.Error("Write string to target file fail! : %v \r\n", err)
				c.ColorPrint(c.Light_red, "Record event fail: %v", err)
			} else {
				data.MonthCost += price
				data.TotalCost += price
				c.ColorPrint(c.Light_green, "Save scuess!\n")
			}
		}
	}
}

//printf a welcome statemt every times open it tools
func printWelcome() {
	c.ColorPrint(c.Light_blue, "\n=====================\n==     COSTER     ==\n=====================\n")
	c.ColorPrint(c.Light_blue, "command: show, clear, end, turn, history, ls\n")
	c.ColorPrint(c.Light_purple, "Welcome Back to Coster !!! \n")
	c.ColorPrint(c.Light_purple, "Last time of using it tool is: ")
	duration := time.Since(data.LastTime)
	c.ColorPrint(c.Light_green, " %d hour %d minute \n", int(duration.Hours())%24, int(duration.Minutes())%60)
	c.ColorPrint(c.Light_purple, "It month temply cost: ")
	c.ColorPrint(c.Light_green, " %.1f \n", data.MonthCost)
	c.ColorPrint(c.Light_purple, "Totaly cost: ")
	c.ColorPrint(c.Light_green, " %.1f \n", data.TotalCost)
}

//use to update the program state before leave
func saveState() {
	data.LastTime = time.Now()
	lastestData, _ := json.Marshal(data)
	err := ioutil.WriteFile(configPath, lastestData, os.ModeSetuid)
	if err != nil {
		logs.Error("Save lastest data to eventer config file fall: %v", err)
	} else {
		fmt.Println("Coster save state scuess!")
	}
}

//printf cost logs of it month to console
func printfCost(filePath string) error {
	absPath, _ := filepath.Abs(filePath)
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("Open file fail when printf the coster logs: %v", err)
	}
	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString(byte('\n'))
		if err == io.EOF { //end of file
			break
		}
		if err != nil {
			return fmt.Errorf("Error happen when printf the coster log: %v", err)
		}
		logsReg, _ := regexp.Compile(`^\([\d\- ]+\)( -){10,} .+\s$`)
		if logsReg.MatchString(line) {
			c.ColorPrint(c.Light_blue, line[:9])
			c.ColorPrint(c.Light_cyan, line[9:39])
			c.ColorPrint(c.Light_green, line[39:])
		} else {
			fmt.Print(line)
		}
	}
	return nil
}

//display the history coster files list
func printfList() error {
	absWPath, _ := filepath.Abs(data.WritePath)
	historyDir := filepath.Dir(absWPath)
	file, err := os.Open(historyDir)
	if err != nil {
		return fmt.Errorf("Open history directory fail: %v", err)
	}
	defer file.Close()
	fi, err := file.Readdir(0)
	if err != nil {
		return fmt.Errorf("Read history directory files list fail: %v", err)
	}
	c.ColorPrint(11, "=========== History ===========\n")
	for _, info := range fi {
		c.ColorPrint(11, info.Name())
		fmt.Println()
	}
	fmt.Println()
	return nil
}
