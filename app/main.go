package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
		tokens := strings.Split(command, " ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)

		}

		if _, err := exec.LookPath(tokens[0]); err == nil {

			var prog = exec.Command(tokens[0], tokens[1:]...)
			prog.Stdout = os.Stdout
			prog.Stderr = os.Stderr
			prog.Run()
		} else {

			fmt.Println(tokens[0] + ": command not found")
		}

	}
}
