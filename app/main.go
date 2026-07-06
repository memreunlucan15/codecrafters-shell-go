package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {

	// TODO: Uncomment the code below to pass the first stage
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")

		command, err := reader.ReadString('\n')
		command = strings.TrimSpace(command)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)

		}
		//fmt.Println(command[:len(command)-1] + ": command not found")

		if strings.HasPrefix(command, "type") {

			if strings.HasSuffix(command, "exit") {
				fmt.Println(strings.TrimPrefix(command, "type ") + " is a shell builtin")
			} else if strings.HasSuffix(command, "echo") {
				fmt.Println(strings.TrimPrefix(command, "type ") + " is a shell builtin")
			} else if strings.HasSuffix(command, "type") {
				fmt.Println(strings.TrimPrefix(command, "type ") + " is a shell builtin")
			} else {
				fmt.Println(strings.TrimPrefix(command + ": command not found"))
			}

		} else if command == "exit" {
			break
		} else if strings.HasPrefix(command, "echo") {
			fmt.Println(strings.TrimPrefix(command, "echo "))
		} else if !strings.HasPrefix(command, "echo") {
			fmt.Println(command + ": command not found")
		}

	}
}
