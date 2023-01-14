package main

import (
	"fmt"
	"monkey/repl"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	filePath := os.Args

	if len(filePath) != 2 {
		fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)
		fmt.Printf("Feel free to type in commands\n")
		repl.StartVM(os.Stdin, os.Stdout)
	}
	repl.StartFile(filePath[1])

	// repl.StartEval(os.Stdin, os.Stdout)
}
