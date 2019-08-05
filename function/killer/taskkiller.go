package killer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	c "../common"
	"github.com/astaxie/beego/logs"
)

var (
	confPath = "./config/killer.json"
	logsPath = "./logs/taskkiller.log"
	confData = config{}
)

type config struct {
	BlackList []string
}

func init() {
	//set logger
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	err := logs.SetLogger(logs.AdapterFile, `{"filename":"./logs/taskkiller.log"}`)
	if err != nil {
		fmt.Println(err)
	}
}

func taskkillerinit() error {
	//read blackList
	file, err := os.Open(confPath)
	if err != nil {
		logs.Error("Open config file error: %v", err)
		if !os.IsNotExist(err) {
			return err
		}
		//create a new config file and init data in it
		file, err = os.Create(confPath)
		if err != nil {
			logs.Error("Create config file error: %v", err)
			return err
		}
		dfDataByte, _ := json.Marshal(config{BlackList: []string{"wps", "8081"}})
		err := ioutil.WriteFile(confPath, dfDataByte, 0644)
		if err != nil {
			logs.Error("Can't not init config: %v", err)
		}
	}
	//read config from file
	defer file.Close()
	buf := bufio.NewReader(file)
	bytes, err := ioutil.ReadAll(buf)
	err = json.Unmarshal(bytes, &confData)
	if err != nil {
		logs.Error("Can't parse config to struct", err)
		return err
	}
	return nil
}

func Run(taskBus chan<- func()) (status int, err error) {
	printWelcome()
	for {
		c.ColorPrint(c.Light_cyan, "Input event or command > ")
		input := ""
		fmt.Scanf("%s\n", &input)
		input = strings.TrimSpace(input)
		switch input {
		case "":
			continue
		case "clear":
			if err = c.ClearConsole(); err != nil {
				logs.Error(err)
			}
		case "turn":
			tcase := c.GetTurnCode()
			if tcase != c.NotFound {
				return tcase, nil
			}
		case "end":
			return c.NormalReturn, nil
		case "kill":
			c.ColorPrint(c.Light_cyan, "Please input a tag to search >")
			tag := scanfWord()
			TaskKiller(tag)
		}
	}
}

func printWelcome() {
	c.ColorPrint(c.Light_cyan, "\n=====================\n==     TASKKILLER     ==\n=====================\n")
	c.ColorPrint(c.Light_cyan, "command: clear, end, turn, kill \n")
}

func TaskKiller(taskName string) {
	var cmd *exec.Cmd
	cmd = exec.Command("cmd", "/c", "tasklist") //Windows example, its tested
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Run()
	regStr := fmt.Sprintf(`(?i)[^\n ]*(%s)[^ ]*.exe\s+\d+.*`, taskName)
	reg, _ := regexp.Compile(regStr)
	ress := reg.FindAllString(buf.String(), -1)
	if len(ress) == 0 {
		fmt.Println("Find nothing!")
		return
	}

	c.ColorPrint(c.Light_blue, "Image Name                     PID Session Name        Session#    Mem Usage\n ")
	c.ColorPrint(c.Light_blue, "==============================================================================\n")
	for _, v := range ress {
		c.ColorPrint(c.Light_blue, "%v\n", v)
	}
	c.ColorPrint(c.Light_red, "Find above task, Do you going to kill all of them?    ")
	input := scanfWord()
	if len(input) == 0 || input[0] != 'y' {
		c.ColorPrint(c.Light_green, "Cancel....\r\n")
		return
	}
	// kill all found task
	reg2, _ := regexp.Compile(`(?i)[^\n ]*.exe\s+`)
	reg3, _ := regexp.Compile(`\d+`)
	for _, v := range ress {
		index := reg2.FindStringIndex(v)
		if index == nil {
			fmt.Printf("Can't find pid from : %s \n", v)
			continue
		}
		PID := reg3.FindString(v[index[1]:])
		killRes := killPid(PID)
		fmt.Printf("%s - - - - %s \n", PID, killRes)
	}
}

func killPid(pid string) string {
	var cmd *exec.Cmd
	cmd = exec.Command("cmd", "/c", "taskkill /pid ", pid) //Windows example, its tested
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Run()
	return strings.TrimSpace(buf.String())
}

func scanfWord() string {
	var temp string
	fmt.Scanf("%s\n", &temp)
	return strings.TrimSpace(temp)
}

func KillBlackList() error {
	err := taskkillerinit()
	if err != nil {
		return err
	}
	for _, v := range confData.BlackList {
		TaskKiller(v)
	}
	return nil
}
