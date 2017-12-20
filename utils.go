package main

import (
	"os/exec"

	"github.com/reconquest/lexec-go"
)

func execute(cmd *exec.Cmd) (string, error) {
	if verbose {
		logger.Debugf("%s", cmd.Args)
	}

	execution := lexec.NewExec(lexec.Loggerf(logger.Tracef), cmd)
	stdout, stderr, err := execution.Output()
	return string(stdout) + string(stderr), err
}

func getMaxLength(elements []string) int {
	maxlength := 0
	for _, element := range elements {
		length := len(element)
		if length > maxlength {
			maxlength = length
		}
	}

	return maxlength
}

func getKeys(items map[string]string) []string {
	keys := []string{}
	for key := range items {
		keys = append(keys, key)
	}

	return keys
}
