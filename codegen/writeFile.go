package codegen

import (
	"fmt"
	"os"
)

func Exist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return false
	}
	if os.IsExist(err) {
		return true
	}
	return true
}

func WriteFile(filename string, content string) {
	var file *os.File
	var err error
	if Exist(filename) {
		_, err := os.OpenFile(filename, os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Open file failed: ", err)
			return
		}
	} else {
		file, err = os.Create(filename)
		if err != nil {
			fmt.Println("file create failed")
			return
		}
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("write file failed: ", err)
		return
	}
	fmt.Println("write file success.")
}
