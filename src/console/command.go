package console

import (
	"RocketTool/src/business"
	"bytes"
	"flag"
	"fmt"
	"os"
)

const version = "0.0.1"

type commandHandler interface {
	process(params []string)

	showUsage()
}

type baseCommand struct {
	name string

	description string
}

type helpCommand struct {
	baseCommand
}

type versionCommand struct {
	baseCommand
}

type exitCommand struct {
	baseCommand
}

type createAccountCommand struct {
	baseCommand
	nodeType int
}

type createGenesisGroupCommand struct {
	baseCommand

	genesisGroupMemberNum uint64
	paramParser           *flag.FlagSet
}

func (command *versionCommand) process(params []string) {
	fmt.Println("Version:" + version)
}

func (command *versionCommand) showUsage() {
	command.baseCommand.showUsage()
}

func (command *exitCommand) process(params []string) {
	fmt.Printf("Thank you, bye\n")
	os.Exit(0)
}

func (command *exitCommand) showUsage() {
	command.baseCommand.showUsage()
}

func (command *helpCommand) process(params []string) {
	for _, command := range consoleInstance.commandHandlerMap {
		command.showUsage()
	}
}

func (command *helpCommand) showUsage() {
	command.baseCommand.showUsage()
}

func (command *createAccountCommand) process(params []string) {
	business.CreateNewAccount(command.nodeType)
}

func (command *createAccountCommand) showUsage() {
	command.baseCommand.showUsage()
}

func (command *createGenesisGroupCommand) process(params []string) {
	if err := command.paramParser.Parse(params); err != nil {
		fmt.Println(err.Error())
		return
	}

	if command.genesisGroupMemberNum < 3 {
		fmt.Println("Illegal param: member_count.The minimum value should be 3.")
	}
	business.CreateGenesisGroup(command.genesisGroupMemberNum)
}

func (command *createGenesisGroupCommand) showUsage() {
	buffer := bytes.Buffer{}
	buffer.WriteString(" ")
	buffer.WriteString(command.name)
	buffer.WriteString(":\t")
	buffer.WriteString(command.description)
	fmt.Println(buffer.String())

	command.paramParser.PrintDefaults()
	fmt.Print("\n")
}

func (command *baseCommand) showUsage() {
	buffer := bytes.Buffer{}
	buffer.WriteString(" ")
	buffer.WriteString(command.name)
	buffer.WriteString(":\t")
	buffer.WriteString(command.description)
	buffer.WriteString("\n")

	fmt.Println(buffer.String())
}

func loadCommandHandler() map[string]commandHandler {
	commandMap := make(map[string]commandHandler)

	helpCommand := helpCommand{newBaseCommand("help", "show command info")}
	commandMap[helpCommand.name] = &helpCommand

	versionCommand := versionCommand{newBaseCommand("version", "show program version")}
	commandMap[versionCommand.name] = &versionCommand

	exitCommand := exitCommand{newBaseCommand("exit", "exit the program")}
	commandMap[exitCommand.name] = &exitCommand

	createNomalAccountCommand := createAccountCommand{newBaseCommand("create_account", "create new proposer account"), -1}
	commandMap[createNomalAccountCommand.name] = &createNomalAccountCommand

	createProposerAccountCommand := createAccountCommand{newBaseCommand("create_proposer_account", "create new proposer account"), 1}
	commandMap[createProposerAccountCommand.name] = &createProposerAccountCommand

	createValidatorAccountCommand := createAccountCommand{newBaseCommand("create_validator_account", "create new validator account"), 0}
	commandMap[createValidatorAccountCommand.name] = &createValidatorAccountCommand

	createGenesisGroupCommand := createGenesisGroupCommand{baseCommand: newBaseCommand("create_genesis_group", "create new genesis group")}
	createGenesisGroupCommand.paramParser = flag.NewFlagSet(createGenesisGroupCommand.name, flag.ContinueOnError)
	createGenesisGroupCommand.paramParser.Uint64Var(&createGenesisGroupCommand.genesisGroupMemberNum, "member_count", 3, "genesis group member count")
	commandMap[createGenesisGroupCommand.name] = &createGenesisGroupCommand

	return commandMap
}

func newBaseCommand(name string, description string) baseCommand {
	command := baseCommand{name: name, description: description}
	return command
}
