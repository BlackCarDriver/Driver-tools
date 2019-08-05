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
	WritePath  string    //where to write record
}

var (
	configPath          = "./config/eventer.json"
	target     *os.File = nil           //the file write and read event logs from
	data                = everterData{} //latest data
	dfData              = everterData{  //default data
		LastTime:   time.Now(),
		TodayTimes: 0,
		WritePath:  "./data/eventer/itmonth.txt",
	}
)

func eventerInit() error {
	//setting up logger
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	logs.SetLogger(logs.AdapterFile, `{"filename":"./logs/eventer.log"}`)
	//read config file
	file, err := os.Open(configPath)
	if err != nil { //read config file fail
		absConPath, _ := filepath.Abs(configPath)
		logs.Warn("Eventer read config file %s fall: %v", absConPath, err)
		fmt.Println("Init data not fond, Create a new noe? (yes/no) ")
		var input string
		fmt.Scanf("%s\n", &input)
		if input = strings.ToLower(input); input[0] == 'y' { //create a init data on config file
			dfDataByte, _ := json.Marshal(dfData)
			err := ioutil.WriteFile(configPath, dfDataByte, 0644)
			if err != nil { //can not create a config file
				return fmt.Errorf("Write init data to eventer config file fall: %v", err)
			}
			fmt.Println("Init data scuess!")
			file.Close() //reopen config file
			file, err = os.Open(configPath)
			if err != nil { //can not read config file after create new one
				return fmt.Errorf("Open config file Fail after init a new config!")
			}
		} else { //user input no
			return fmt.Errorf("Inti data not found")
		}
	}
	//open config scuess and going to read config
	defer file.Close()
	buf := bufio.NewReader(file)
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return fmt.Errorf("Eventer read config file fall: %v", err)
	}
	//read config scuess and going to load config
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return fmt.Errorf("Unmarshal eventer config fail: %v", err)
	}
	//config load scuess
	//check target file exist, create a new one if not exist
	_, err = os.Stat(data.WritePath)
	if err != nil {
		logs.Error("Open write file fail : %v", err)
		_, err := os.Create(data.WritePath)
		if err != nil {
			return fmt.Errorf("Create write file fail: %v", err)
		} else {
			fmt.Println("Already create a new enverter file!")
		}
	}
	//write to another file if start a new month
	if data.LastTime.Month() != time.Now().Month() {
		absWPath, err := filepath.Abs(data.WritePath)
		if err != nil {
			logs.Error("Get absolute path of write-path fail:%v", err)
		}
		dir := filepath.Dir(absWPath)
		fileName := fmt.Sprintf("%s/%04d-%d.txt", dir, data.LastTime.Year(), data.LastTime.Month())
		err = os.Rename(data.WritePath, fileName)
		if err != nil {
			logs.Error("Rename %s to %s fail: %v", data.WritePath, fileName, err)
		} else {
			c.ColorPrint(c.Light_yellow, "============== Happy Good Month! ==============\n")
			_, err = os.Create(data.WritePath)
			if err != nil {
				logs.Error("Create new file fail after rename old file: %v", err)
			}
		}
	}
	//open target file with write to end model
	target, err = os.OpenFile(data.WritePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("Open target file fail after guarantee it is exist!: %v", err)
	}
	//printf a timestamp if start a new day
	if time.Now().Day() != data.LastTime.Day() {
		c.ColorPrint(c.Light_purple, "Have A Good Day!\n")
		event := fmt.Sprintf("\n===============================[ %s ]===============================\n", time.Now().Format("01-02 Mon"))
		_, err = target.WriteString(event)
		data.TodayTimes = 0
	}
	printWelcome()
	return nil
}

func Run(taskBus chan<- func()) (status int, err error) {
	err = eventerInit()
	taskBus <- saveState
	defer target.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		c.ColorPrint(9, "Input event or command > ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return c.ErrorExit, fmt.Errorf("Bufio ReadString fail!: %s \r\n", err)
		}
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
		case "key":
			tcase := c.GetKeyCode()
			if tcase != c.NotFound {
				return tcase, nil
			}
		case "show":
			if err = printftEvent(data.WritePath); err != nil {
				logs.Error("Printf it month event fail: %v \r\n", err)
			}
		case "ls":
			if err = printfList(); err != nil {
				logs.Error("Printf list fail: %v \r\n", err)
			}
		case "history":
			c.ColorPrint(9, "Input which file you want to see > ")
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)
			reg, err := regexp.Compile(`^\d+-\d+$`)
			if !reg.MatchString(input) {
				c.ColorPrint(12, "File name not right, should like '2019-8'!\n")
				continue
			}
			if err = printfHistory(input); err != nil {
				fmt.Println(err)
				logs.Error(err)
			}
		default:
			now := time.Now()
			event := fmt.Sprintf("\r\n( %02d:%02d ) - - - - - - - - - - - - - - - %s\r\n", now.Hour(), now.Minute(), input)
			_, err = target.WriteString(event)
			if err != nil {
				logs.Error("Write string to target file fail! : %v \r\n", err)
				c.ColorPrint(12, "Record event fail: %v", err)
			} else {
				data.TodayTimes++
				c.ColorPrint(3, "Save scuess!\n")
			}
		}
	}
}

//printf welcome message
func printWelcome() {
	c.PrintfColorExample()
	c.ColorPrint(13, "\n=====================\n==     EVENTER     ==\n=====================\n")
	c.ColorPrint(13, "command: show, clear, end, turn, history, ls\n")
	c.ColorPrint(11, "Welcome Back to Eventer !!! \n")
	c.ColorPrint(11, "Last time of using it tool is: ")
	duration := time.Since(data.LastTime)
	c.ColorPrint(10, " %d hour %d minute \n", int(duration.Hours())%24, int(duration.Minutes())%60)
	c.ColorPrint(11, "The numbers you save event is: ")
	c.ColorPrint(10, " %d \n", data.TodayTimes)
}

//printf event logs of it month to console
func printftEvent(filePath string) error {
	absPath, _ := filepath.Abs(filePath)
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("Open file fail when printf the eventer logs: %v", err)
	}
	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString(byte('\n'))
		if err == io.EOF { //end of file
			break
		}
		if err != nil {
			return fmt.Errorf("Error happen when printf the eventer log: %v", err)
		}
		dateReg, _ := regexp.Compile(`^=+\[[^\]]+\]=+\s*$`)
		if dateReg.MatchString(line) {
			c.ColorPrint(6, line)
			continue
		}
		logsReg, _ := regexp.Compile(`^\([\d\: ]+\)( -){10,} .+\s$`)
		if logsReg.MatchString(line) {
			c.ColorPrint(12, line[:9])
			c.ColorPrint(5, line[9:39])
			c.ColorPrint(11, line[39:])
		} else {
			fmt.Print(line)
		}
	}
	return nil
}

//use to update the program state before leave
func saveState() {
	data.LastTime = time.Now()
	lastestData, _ := json.Marshal(data)
	err := ioutil.WriteFile(configPath, lastestData, os.ModeSetuid)
	if err != nil {
		logs.Error("Save lastest data to eventer config file fall: %v", err)
	} else {
		fmt.Println("Eventer save state scuess!")
	}
}

//display the history eventer files list
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

//printf specified month event logs
func printfHistory(name string) error {
	absWPath, _ := filepath.Abs(data.WritePath)
	historyDir := filepath.Dir(absWPath)
	path := fmt.Sprintf("%s/%s.txt", historyDir, name)
	err := printftEvent(path)
	return err
}
