package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

var sc = bufio.NewScanner(os.Stdin)

func findDelimiter(target []string, delimiter string) (result bool) {
	result = false
	for _, i := range target {
		if i == delimiter {
			result = true
			break
		}
	}
	return
}

func forkExecve(command string, args []string) (result bool) {
	cpath, err := exec.LookPath(command)
	if err != nil {
		fmt.Printf("%s: No such file or directory\n", command)
		return
	}
	attr := syscall.ProcAttr{Files: []uintptr{0, 1, 2}}
	pid, err := syscall.ForkExec(cpath, args, &attr)
	if err != nil {
		fmt.Println(err)
		return
	}
	proc, err := os.FindProcess(pid)
	status, err := proc.Wait()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !status.Success() {
		status := strings.Split(status.String(), " ")
		fmt.Printf("Process %d existed with status(%s).\n", pid, status[2])
		return
	}
	result = status.Success()
	return
}

func main() {
	// Ctrl + Cでの終了を防止
	// Ctrl + Zでの終了の防止
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTSTP)

	var input string
	cnt := 0

	for {
		fmt.Printf("./myshell[%02d]> ", cnt)
		if sc.Scan() {
			input = sc.Text()
		} else {
			// Ctrl + Dで終了
			fmt.Println()
			break
		}
		array := strings.Split(input, " ")
		// byeで終了
		if array[0] == "bye" {
			break
		}
		if len(array[0]) != 0 {
			if findDelimiter(array, "?") {
				if findDelimiter(array, ":") {
					area0 := strings.Split(input, "?")
					area1 := strings.Split(area0[1], ":")

					array := strings.Split(strings.Trim(area0[0], " "), " ")
					args := array[0:]
					result := forkExecve(array[0], args)

					if result {
						array := strings.Split(strings.Trim(area1[0], " "), " ")
						args := array[0:]
						forkExecve(array[0], args)
					} else {
						array := strings.Split(strings.Trim(area1[1], " "), " ")
						args := array[0:]
						forkExecve(array[0], args)
					}
				} else {
					fmt.Println("Syntax error.")
				}
			} else {
				args := array[0:]
				forkExecve(array[0], args)
			}

		}
		cnt++
	}
}
