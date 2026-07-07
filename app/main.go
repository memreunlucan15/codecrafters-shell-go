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
		tokens := strings.Split(command, " ") // Tokenlere ayırma
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)

		}

		if strings.ContainsAny(command, "'") {
			tokens_q := strings.SplitAfter(command, "'")
			fmt.Println(tokens_q[0:])
		}

		if _, err := exec.LookPath(tokens[0]); err == nil { // Path kontrolü

			var prog = exec.Command(tokens[0], tokens[1:]...)
			prog.Stdout = os.Stdout
			prog.Stderr = os.Stderr
			prog.Run()
		} else if tokens[0] == "type" {

			switch tokens[1] { // type sonrası builtin komut kontrolü
			case "exit", "echo", "type", "pwd", "cd":
				fmt.Println(tokens[1] + " is a shell builtin")
			default:

				path, err1 := exec.LookPath(tokens[1])
				if err1 == nil {
					fmt.Println(tokens[1] + " is " + path)
				} else {
					fmt.Println(tokens[1] + ": not found")
				}
			}

		} else if tokens[0] == "pwd" { // pwd ile ablosute path alma

			abs_path, _ := os.Getwd()
			fmt.Println(abs_path)

		} else if tokens[0] == "cd" { // cd ile directory değişimi

			if tokens[1] != "~" {
				err = os.Chdir(tokens[1])
				if err != nil {
					fmt.Println("cd: " + tokens[1] + ": No such file or directory")
				}
			} else {
				home_dir, _ := os.UserHomeDir()
				_ = os.Chdir(home_dir)
			}

		} else if command == "exit" {
			break
		} else if tokens[0] == "echo" {
			fmt.Println(strings.TrimPrefix(command, "echo "))
		} else {
			fmt.Println(command + ": command not found")
		}

	}
}
