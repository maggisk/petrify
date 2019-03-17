package main

import "fmt"
import "os"
import "net"
import "net/http"
import "bufio"
import "strings"

var VERBOSE bool = false

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

func verbose(format string, args ...interface{}) {
	if VERBOSE {
		fmt.Print("[VERBOSE] ")
		fmt.Printf(format+"\n", args...)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func ServeStatic(dir string) int {
	listener, err := net.Listen("tcp", ":0")
	checkError(err)
	port := listener.Addr().(*net.TCPAddr).Port
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)
	go http.Serve(listener, nil)
	return port
}

func ReadLine() string {
	reader := bufio.NewReader(os.Stdin)
	s, _ := reader.ReadString('\n')
	return s
}

func ReadYesNo(prompt string) bool {
	fmt.Println(prompt)
	for {
		a := strings.ToLower(strings.TrimSpace(ReadLine()))
		if a == "y" || a == "yes" {
			return true
		} else if a == "n" || a == "no" {
			return false
		}
		fmt.Println("Please answer yes or no")
	}
}
