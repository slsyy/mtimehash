package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmdtest"
	"github.com/stretchr/testify/require"
)

func TestCLI(t *testing.T) {
	silentLogs = true
	ts, err := cmdtest.Read("testdata")
	require.NoError(t, err)
	ts.Commands["mtimehash"] = cmdtest.InProcessProgram("mtimehash", run)
	ts.Commands["fileNotEmpty"] = cmdtest.InProcessProgram("fileNotEmpty", func() int {
		args := os.Args
		s, err := os.Stat(args[1])
		if err != nil {
			fmt.Printf("os.Stat: %s\n", err.Error())
			return 1
		}
		if s.Size() == 0 {
			fmt.Print("file is empty\n")
			return 1
		}
		return 0
	})
	ts.Run(t, false)
}
