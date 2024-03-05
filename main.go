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
		// codegen
		// fmt.Println("input filename!")
	}
	repl.StartFile(filePath[1])
	//codegen
	// repl.StartWriteFile(filePath[1])
	// eval
	// repl.StartEval(os.Stdin, os.Stdout)
}
