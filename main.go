package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/rabbitt/maxmind/command"
	"github.com/rabbitt/maxmind/mm"
)

func main() {

	ui := &command.BaseUi{
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	c := cli.NewCLI(mm.NewPathname(os.Args[0]).Basename().Path(), "0.0.1")
	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"server": func() (cli.Command, error) {
			return &command.ServerCommand{
				Ui:     &command.PrefixedUi{Ui: ui},
				Config: mm.NewConfiguration(),
			}, nil
		},
		"lookup": func() (cli.Command, error) {
			return &command.LookupCommand{Ui: ui}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	os.Exit(exitStatus)
}
