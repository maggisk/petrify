package main

import "fmt"
import "os"

// permission bitmasks for new files and directories
var FILE_BITMASK os.FileMode = 0644
var DIR_BITMASK os.FileMode = 0755

func info(format string, args ...interface{}) {
	fmt.Print("[INFO] ")
	fmt.Printf(format+"\n", args...)
}

func warn(format string, args ...interface{}) {
	fmt.Print("[WARNING] ")
	fmt.Printf(format+"\n", args...)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
