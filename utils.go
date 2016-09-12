package main

import (
	"log"
	"os/exec"

	"github.com/reconquest/executil-go"
)

func execute(cmd *exec.Cmd) (string, error) {
	if verbose {
		if cmd.Dir != "" {
			log.Printf("exec %q in %s", cmd.Args, cmd.Dir)
		} else {
			log.Printf("exec %q", cmd.Args)
		}
	}

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
	for key := range items {
		keys = append(keys, key)
	}

	return keys
}
