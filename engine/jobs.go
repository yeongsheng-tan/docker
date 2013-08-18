package engine

import (
	"github.com/dotcloud/beam"
	"fmt"
	"os/exec"
)

func JobNotImplemented(name string, args []string, env map[string]string, streams beam.Streamer, db beam.DB) error {
	return fmt.Errorf("Not yet implemented: %s", name)
}

func JobExec(name string, args []string, env map[string]string, streams beam.Streamer, db beam.DB) error {
	var (
		cmdName string
		cmdArgs []string
	)
	if len(args) >= 1 {
		cmdName = args[0]
	} else {
		return fmt.Errorf("Not enough arguments")
	}
	if len(args) > 1 {
		cmdArgs = args[1:]
	}
	p := exec.Command(cmdName, cmdArgs...)
	p.Stdin = streams.OpenRead("stdin")
	p.Stdout = streams.OpenWrite("stdout")
	p.Stderr = streams.OpenWrite("stderr")
	return p.Run()
}

