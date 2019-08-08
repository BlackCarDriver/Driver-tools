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
			c.ColorPrint(c.Light_yellow, "Please input a tag to search >")
			tag := c.ScanfWord()
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
	input := c.ScanfWord()
	if len(input) == 0 || input[0] != 'y' {
		c.ColorPrint(c.Light_green, "Cancel....\r\n")
		return
	}
	//collect all process name in nonredundant
	pnames := make(map[string]string, 0)
	for _, v := range ress {
		if pname, err := getName(v); err != nil {
			c.ColorPrint(c.Light_red, "%v", err)
		} else {
			pnames[pname] = pname
		}
	}
	// kill all process record in pnames
	for _, n := range pnames {
		killRes := KillPname(n)
		c.ColorPrint(c.Light_green, "- - - - - %s - - - - \n", n)
		c.ColorPrint(c.Light_blue, "%s\n", killRes)
	}

}

// find process name
func getName(line string) (string, error) {
	reg, _ := regexp.Compile(`[^\n][\S]+.exe`)
	index := reg.FindStringIndex(line)
	if index == nil {
		return "", fmt.Errorf("Not find process name in %s", line)
	}
	return line[index[0]:index[1]], nil
}

// kill a process by name
func KillPname(name string) string {
	var cmd *exec.Cmd
	cmd = exec.Command("cmd", "/c", "taskkill /F /IM ", name) //Windows example, its tested
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Run()
	return strings.TrimSpace(buf.String())
}

// autoly search and kill all process in blacklist
func KillBlackList() error {
	err := taskkillerinit()
	if err != nil {
		return err
	}
	for _, v := range confData.BlackList {
		c.ColorPrint(c.Light_yellow, "%s \n", v)
		TaskKiller(v)
	}
	fmt.Print("Input anything > ")
	c.WaitInput()
	return nil
}
