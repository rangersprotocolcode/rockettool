package console

import (
	"fmt"
	"github.com/peterh/liner"
	"os"
	"regexp"
	"strings"
)

var consoleInstance *console

type console struct {
	commandHandlerMap map[string]commandHandler
}

func InitConsole() {
	ctrlC := getExitSignalsChan()
	go handleExit(ctrlC)

	if consoleInstance == nil {
		consoleInstance = &console{}
		consoleInstance.commandHandlerMap = loadCommandHandler()
	}

	fmt.Println("Welcome to use rocket tool!")
	consoleInstance.readCommandLoop()
}

func (console *console) readCommandLoop() {
	reader := liner.NewLiner()
	defer reader.Close()
	reader.SetCtrlCAborts(true)

	for {
		commandLine := parseInput(reader)
		if len(commandLine) == 0 {
			continue
		}

		commandName := commandLine[0]
		command, exist := console.commandHandlerMap[commandName]
		if !exist {
			fmt.Printf("not supported command %v\n", commandName)
			continue
		}

		params := commandLine[1:]
		command.process(params)
	}
}

func parseInput(reader *liner.State) []string {
	input, err := reader.Prompt(fmt.Sprintf("rocketTool:> "))
	if err != nil {
		if err == liner.ErrPromptAborted {
			reader.Close()
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
	}

	input = strings.TrimSpace(input)
	if input == "" || len(input) == 0 {
		return nil
	}
	regex, _ := regexp.Compile("\\s{2，}")
	input = regex.ReplaceAllString(input, " ")
	commandLine := strings.Split(input, " ")

	reader.AppendHistory(input)
	return commandLine
}
