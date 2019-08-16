package learner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	c "../common"

	"github.com/astaxie/beego/logs"
	// "github.com/axgle/mahonia"
)

const configPath = "./config/learner.json"
const writePath = "./data/learner/learner.txt"

var (
	log    *logs.BeeLogger
	data   *learnerData //latest data
	target *os.File     //the file write and read cost log from
)

type learnerData struct {
	LastTime   time.Time //lasttime of using it tool
	TotalTag   int       //total times of record tag
	ItMonthTag int       //how many tag have been record in it month
}

var dfData = learnerData{ //default data
	LastTime:   time.Now(),
	TotalTag:   0,
	ItMonthTag: 0,
}

func learnerInit() error {
	//setting up logger
	log = logs.NewLogger()
	log.EnableFuncCallDepth(true)
	log.SetLogFuncCallDepth(3)
	err := log.SetLogger(logs.AdapterFile, `{"filename":"./logs/learner.log"}`)
	if err != nil {
		fmt.Printf("Create log file fail: %v", err)
	}
	//read config file
	file, err := os.Open(configPath)
	if err != nil { //read config file fail
		absConPath, _ := filepath.Abs(configPath)
		log.Warn("Learner read config file '%s' fall: %v", absConPath, err)
		fmt.Println("Init data not fond, Create a new noe? (yes/no) ")
		var input string
		fmt.Scanf("%s\n", &input)
		if input = strings.ToLower(input); input[0] == 'y' { //create a init data on config file
			dfDataByte, _ := json.Marshal(dfData)
			err := ioutil.WriteFile(configPath, dfDataByte, 0644)
			if err != nil { //can not create a config file
				err = fmt.Errorf("Write init data to learner config file fall: %v", err)
				log.Error("%v", err)
				return err
			}
			fmt.Println("Init data scuess!")
			file.Close() //reopen config file
			file, err = os.Open(configPath)
			if err != nil { //can not read config file after create new one
				err = fmt.Errorf("Open config file Fail after init a new config!")
				log.Error("%v", err)
				return err
			}
		} else { //user input 'no'
			err = fmt.Errorf("Init data not found")
			log.Error("%v", err)
			return err
		}
	}
	//open config scuess and going to read config
	defer file.Close()
	buf := bufio.NewReader(file)
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		err = fmt.Errorf("Learner read config file fall: %v", err)
		log.Error("%v", err)
		return err
	}
	//read config scuess and going to load config
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		err = fmt.Errorf("Unmarshal learner config fail: %v", err)
		log.Error("%v", err)
		return err
	}
	//learner load scuess
	//check target file exist, create a new one if not exist
	_, err = os.Stat(writePath)
	if err != nil {
		log.Error("Open write file fail : %v", err)
		_, err := os.Create(writePath)
		if err != nil {
			err = fmt.Errorf("Create write file fail: %v", err)
			log.Error("%v", err)
			return err
		} else {
			fmt.Println("Already create a new enverter file!")
		}
	}
	//open target file with write to end model
	target, err = os.OpenFile(writePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		err = fmt.Errorf("Open target file fail after guarantee it is exist!: %v", err)
		log.Error("%v", err)
		return err
	}
	//write to another file if start a new month
	if data.LastTime.Month() != time.Now().Month() {
		//record the monthly statistics at the end of target file
		tailStamp := fmt.Sprintf("\n statics :  [ Month: %s ]   [ ItMonty: %d ]   [ Total: %d ] \n", data.LastTime.Month(), data.ItMonthTag, data.TotalTag)
		_, err = target.WriteString(tailStamp)
		c.ColorPrint(c.Light_green, tailStamp)
		if err != nil {
			err = fmt.Errorf("Write tailStamp to learner target file fail: %v", err)
			log.Error("%v", err)
			return err
		}
		//save lastmonth history to another file
		absWPath, err := filepath.Abs(writePath)
		if err != nil {
			log.Error("Get absolute path of writePath fail: %v", err)
		}
		dir := filepath.Dir(absWPath)
		target.Close()
		fileName := fmt.Sprintf("%s/%04d-%d.txt", dir, data.LastTime.Year(), data.LastTime.Month())
		err = os.Rename(writePath, fileName)
		if err != nil {
			log.Error("Rename %s to %s fail: %v", writePath, fileName, err)
		} else {
			c.ColorPrint(c.Light_yellow, "============== Happy New Month! ==============\n")
			_, err = os.Create(writePath)
			if err != nil {
				log.Error("Create new file fail after rename old file: %v", err)
			}
		}
		target, err = os.OpenFile(writePath, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Error("Open target file fail after rename: %v", err)
			return err
		}
		data.ItMonthTag = 0
		saveState()
	}
	printWelcome()
	return nil
}

func Run(taskBus chan<- func()) (status int, err error) {
	err = learnerInit()
	if err != nil {
		return c.ErrorExit, err
	}
	taskBus <- saveState
	defer target.Close()
	for {
		c.ColorPrint(c.Light_cyan, "Input new tag or command > ")
		input := c.ScanfWord()
		input = strings.TrimSpace(input)
		switch input {
		case "":
			continue
		case "end", "exit":
			return c.NormalReturn, nil
		case "clear":
			if err = c.ClearConsole(); err != nil {
				log.Error("%v", err)
			}
		case "turn":
			tcase := c.GetTurnCode()
			if tcase != c.NotFound {
				return tcase, nil
			}
		case "show":
			if err = printfStudy(writePath); err != nil {
				log.Error("Printf it month cost record fail: %v \r\n", err)
			}
		case "ls":
			if err = printfList(); err != nil {
				log.Error("Printf list fail: %v \r\n", err)
			}
		case "key":
			tcase := c.GetKeyCode()
			if tcase != c.NotFound {
				return tcase, nil
			}
		case "his":
			c.ColorPrint(c.Light_cyan, "Input which file you want to see > ")
			input := c.ScanfWord()
			reg, err := regexp.Compile(`^\d+-\d+$`)
			if !reg.MatchString(input) {
				c.ColorPrint(c.Light_red, "File name not right, should like '2019-8'!\n")
				continue
			}
			absWPath, _ := filepath.Abs(writePath)
			historyDir := filepath.Dir(absWPath)
			path := fmt.Sprintf("%s/%s.txt", historyDir, input)
			err = printfStudy(path)
			if err != nil {
				fmt.Println("show history fail: ", err)
			}

		default:
			now := time.Now()
			c.ColorPrint(c.Light_cyan, "Input what do you learn just now: > ")
			record := fmt.Sprintf("时间:[ %02d-%02d ] - - - - - - - - - - - - - 技术：%s \r\n", now.Month(), now.Day(), input)
			_, err = target.WriteString(record)
			if err != nil {
				log.Error("Write string to target file fail! : %v \r\n", err)
				c.ColorPrint(c.Light_red, "Record event fail: %v", err)
			} else {
				data.TotalTag += 1
				data.ItMonthTag += 1
				c.ColorPrint(c.Light_green, "Save scuess!\n")
			}
		}
	}
}

//printf a welcome statemt every times open it tools
func printWelcome() {
	c.ColorPrint(c.Light_cyan, "\n=====================\n==     Learner     ==\n=====================\n")
	c.ColorPrint(c.Light_cyan, "command: show, clear, end, turn, his, ls\n")
	c.ColorPrint(c.Light_purple, "Welcome Back to learner !!! \n")
	duration := time.Since(data.LastTime)
	c.ColorPrint(c.Light_purple, "Last time of using it tool is: ")
	c.ColorPrint(c.Light_green, " %.0f days %d hour %d minute \n", duration.Hours()/24, int(duration.Hours())%24, int(duration.Minutes())%60)
	c.ColorPrint(c.Light_purple, "Numbers of new technologys it month have learn: ")
	c.ColorPrint(c.Light_green, " %d \n", data.ItMonthTag)
	c.ColorPrint(c.Light_purple, "Numbers of new technologys in total: ")
	c.ColorPrint(c.Light_green, " %d \n", data.TotalTag)
}

//use to update the program state before leave
func saveState() {
	data.LastTime = time.Now()
	lastestData, _ := json.Marshal(data)
	err := ioutil.WriteFile(configPath, lastestData, os.ModeSetuid)
	if err != nil {
		log.Error("Save lastest data to eventer config file fall: %v", err)
	} else {
		fmt.Println("learner save state scuess!")
	}
}

//printf cost log of it month to console
func printfStudy(filePath string) error {
	absPath, _ := filepath.Abs(filePath)
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("Open file fail when printf the learner log: %v", err)
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString(byte('\n'))
		if err == io.EOF { //end of file
			break
		}
		if err != nil {
			return fmt.Errorf("Error happen when printf the learner log: %v", err)
		}
		if strings.HasPrefix(line, "时间") {
			format := "时间:[ %s ] - - - - - - - - - - - - - 技术：%s"
			var date string
			var technology string
			fmt.Sscanf(line, format, &date, &technology)
			c.ColorPrint(c.Light_cyan, "时间:[")
			c.ColorPrint(c.Light_green, " %s ", date)
			c.ColorPrint(c.Light_cyan, "] - - - - - - - - - - - - - 技术：")
			c.ColorPrint(c.Light_green, " %s \r\n", technology)
		} else {
			fmt.Print(line)
		}
	}
	return nil
}

//display the history learner files list
func printfList() error {
	absWPath, _ := filepath.Abs(writePath)
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
