package eventer

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

var (
	l *log.Logger
)

type everterData struct {
	LastTime   time.Time //lasttime of using it tool
	TodayTimes int       //how many times using it tool itday
	LogsPath   string    //where to save the logs file
	WritePath  string    //where to write record
}

//default config
var (
	configPath          = "./config/eventer.conf"
	target     *os.File = nil           //the file write and read event logs from
	data                = everterData{} //latest data
	dfData              = everterData{  //default data
		LastTime:   time.Now(),
		TodayTimes: 0,
		WritePath:  "./data/eventer/itmonth.txt",
		LogsPath:   "./logs/eventer.log",
	}
)

func init() {
	//setting up logger
	//logs.SetLogger(logs.AdapterFile, `{"filename":"eventer.log"}`)
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	//read config file
	file, err := os.Open(configPath)
	defer file.Close()
	absConPath, _ := filepath.Abs(configPath)
	if err != nil { //read config file fail
		logs.Warn("Eventer read config file %s fall: %v", absConPath, err)
		fmt.Println("Init data not fond, Create a new noe? (yes/no) ")
		var input string
		fmt.Scanf("%s\n", &input)
		if input = strings.ToLower(input); input[0] == 'y' { //create a init data on config file
			dfDataByte, _ := json.Marshal(dfData)
			err := ioutil.WriteFile(configPath, dfDataByte, 0644)
			if err != nil {
				logs.Error("Write init data to eventer config file fall: %v", err)
				close()
			}
			fmt.Println("Init data scuess!")
			file.Close()
			file, err = os.Open(configPath) //reopen config file
			if err != nil {
				logs.Error("Open config file Fail after init a new config!")
				return
			}
		} else {
			fmt.Println("exit")
			close()
		}
	}
	//load config
	buf := bufio.NewReader(file)
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		logs.Error("Eventer read config file fall: %v", err)
	}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		logs.Error("Unmarshal eventer config fail: %v", err)
		close()
	}
	//check target file exist, create a new one if not exist
	_, err = os.Stat(data.WritePath)
	if err != nil {
		logs.Error("Open write file fail : %v", err)
		_, err := os.Create(data.WritePath)
		if err != nil {
			logs.Error("Create write file fail: %v", err)
			close()
		} else {
			fmt.Println("Already create a new enverter file!")
		}
	}
	//write to another file if start a new month
	if data.LastTime.Month() == time.Now().Month() {
		absWPath, err := filepath.Abs(data.WritePath)
		if err != nil {
			logs.Error("Get absolute path of write-path fail:%v", err)
		}
		dir := filepath.Dir(absWPath)
		fileName := fmt.Sprintf("%s/%04d-%d.txt", dir, data.LastTime.Year(), data.LastTime.Month())
		fmt.Println(fileName)
		err = os.Rename(data.WritePath, fileName)
		if err != nil {
			logs.Error("Rename %s to %s fail: %v", data.WritePath, fileName, err)
		} else {
			c.ColorPrint(14, "Happy Good Month!\n")
			_, err = os.Create(data.WritePath)
			if err != nil {
				logs.Error("Create new file fail after rename old file: %v", err)
			}
		}
	}
	//open target file with write to end model
	target, err = os.OpenFile(data.WritePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logs.Error("Open target file fail after guarantee it is exist!: %v", err)
		close()
	}
	//printf a timestamp if start a new day
	now := time.Now()
	if now.Day() != data.LastTime.Day() {
		c.ColorPrint(5, "Have A Good Day!\n")
		event := fmt.Sprintf("\n===============================[ %s ]===============================\n", now.Format("01-02 Mon"))
		_, err = target.WriteString(event)
		data.TodayTimes = 0
	}
	printWelcome()
}

func Run() (int, error) {
	defer saveState()
	defer target.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		c.ColorPrint(9, "Input event or command > ")
		input, err := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if err != nil {
			logs.Error("Bufio ReadString fail!: ", err)
			close()
		}
		if input == "" {
			continue
		}
		if input == "end" {
			return c.Normal, nil
		}
		if input == "clear" {
			if err = c.ClearConsole(); err != nil {
				logs.Error(err)
			}
			continue
		}
		if input == "turn" { //switch to another application
			tcase := c.GetTurnCode()
			if tcase == c.NotFound {
				continue
			}
			return tcase, nil
		}
		if input == "show" {
			printfEvent()
			continue
		}
		now := time.Now()
		event := fmt.Sprintf("\n( %02d:%02d ) - - - - - - - - - - - - - - - %s\n", now.Hour(), now.Minute(), input)
		_, err = target.WriteString(event)
		if err != nil {
			logs.Error("Write string to target file fail! : %v", err)
		} else {
			data.TodayTimes++
			c.ColorPrint(3, "Save scuess!\n")
		}
	}
}

//printf welcome message
func printWelcome() {
	for i := 0; i <= 15; i++ {
		c.ColorPrint(i, "=%d=", i)
	}
	c.ColorPrint(5, "\n=====================\n==     EVENTER     ==\n=====================\n")
	c.ColorPrint(5, "cmd: clear, end, turn\n")
	c.ColorPrint(11, "Welcome Back to Eventer !!! \n")
	c.ColorPrint(11, "Last time of using it tool is: ")
	duration := time.Since(data.LastTime)
	c.ColorPrint(10, " %d hour %d minute \n", int(duration.Hours())%24, int(duration.Minutes())%60)
	c.ColorPrint(11, "The numbers you save event is: ")
	c.ColorPrint(10, " %d \n", data.TodayTimes)
}

//printf event log to console
func printfEvent() {
	absPath, _ := filepath.Abs(data.WritePath)
	file, err := os.Open(absPath)
	if err != nil {
		logs.Error("Open file faill when printf the eventer logs: %v", err)
		return
	}
	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString(byte('\n'))
		if err == io.EOF { //end of file
			break
		}
		if err != nil {
			logs.Error("Error happen when printf the eventer log: %v", err)
			break
		}
		dateReg, _ := regexp.Compile(`^=+\[[^\]]+\]=+\s*$`)
		if dateReg.MatchString(line) {
			c.ColorPrint(6, line)
			continue
		}
		logsReg, _ := regexp.Compile(`^\([\d\: ]+\)( -){10,} .+\s$`)
		if logsReg.MatchString(line) {
			c.ColorPrint(12, line[:10])
			c.ColorPrint(5, line[10:40])
			c.ColorPrint(11, line[40:])
		} else {
			fmt.Print(line)
		}
	}
}

//use to update the program state before leave
func saveState() {
	fmt.Println("Eventer save state scuess!")
	data.LastTime = time.Now()
	lastestData, _ := json.Marshal(data)
	err := ioutil.WriteFile(configPath, lastestData, os.ModeSetuid)
	if err != nil {
		logs.Error("Save lastest data to eventer config file fall: %v", err)
		close()
	}
}

//close the program
func close() {
	time.Sleep(time.Second * 1)
	os.Exit(1)
}
