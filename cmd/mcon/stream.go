package main

import (
	"bufio"
	"os/exec"
)

func streamCommand(arg string, args ...string) (chan string, error) {
	cmd := exec.Command(arg, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	channel := make(chan string)

	go func() {
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				panic(err)
			}

			channel <- scanner.Text()
		}

		cmd.Process.Kill()
	}()

	return channel, nil
}
