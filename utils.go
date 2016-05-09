package main

import (
	"os/exec"

	"github.com/kovetskiy/executil"
)

func execute(cmd *exec.Cmd) (string, error) {
	stdout, _, err := executil.Run(cmd)
	return string(stdout), err
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
	for key, _ := range items {
		keys = append(keys, key)
	}

	return keys
}
