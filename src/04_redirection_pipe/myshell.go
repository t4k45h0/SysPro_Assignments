package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

var sc = bufio.NewScanner(os.Stdin)

// 配列から指定の記号を探す
func findCharacter(target []string, character string) (result bool) {
	result = false
	for _, i := range target {
		if strings.Index(i, character) != -1 {
			result = true
			break
		}
	}
	return
}

// コマンド実行：ファイルディスクリプタがすべてターミナル
func forkExecve(command []string) (result bool) {
	cpath, err := exec.LookPath(command[0])
	if err != nil {
		fmt.Printf("%s: No such file or directory\n", command[0])
		return
	}
	args := command[0:]
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

// コマンド実行：ファイルディスクリプタの入力がファイル
func redirectionInput(command []string, filename string) (result bool) {
	cpath, err := exec.LookPath(command[0])
	if err != nil {
		fmt.Printf("%s: No such file or directory\n", command[0])
		return
	}
	args := command[0:]

	fr, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fr.Close(); err != nil {
			panic(err)
		}
	}()

	attr := syscall.ProcAttr{Files: []uintptr{fr.Fd(), 1, 2}}
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

// コマンド実行：ファイルディスクリプタの出力がファイル
func redirectionOutput(command []string, filename string) (result bool) {
	cpath, err := exec.LookPath(command[0])
	if err != nil {
		fmt.Printf("%s: No such file or directory\n", command[0])
		return
	}
	args := command[0:]

	fw, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fw.Close(); err != nil {
			panic(err)
		}
	}()

	attr := syscall.ProcAttr{Files: []uintptr{0, fw.Fd(), 2}}
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

// コマンド実行：ファイルディスクリプタのすべてファイル
func redirectionUnion(command []string, iFilename string, oFilename string) (result bool) {
	cpath, err := exec.LookPath(command[0])
	if err != nil {
		fmt.Printf("%s: No such file or directory\n", command[0])
		return
	}
	args := command[0:]

	fr, err := os.OpenFile(iFilename, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fr.Close(); err != nil {
			panic(err)
		}
	}()

	fw, err := os.OpenFile(oFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fw.Close(); err != nil {
			panic(err)
		}
	}()

	attr := syscall.ProcAttr{Files: []uintptr{fr.Fd(), fw.Fd(), 2}}
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

// コマンド実行：パイプライン
func redirectionPipeline(command [][]string, size int) {
	if size > 2 {
		cmds := make([]*exec.Cmd, len(command))
		var err error

		for i, c := range command {
			cmds[i] = exec.Command(c[0], c[1:]...)
			if i > 0 {
				if cmds[i].Stdin, err = cmds[i-1].StdoutPipe(); err != nil {
					panic(err)
					return
				}
			}
			cmds[i].Stderr = os.Stderr
		}
		var out bytes.Buffer
		cmds[len(cmds)-1].Stdout = &out
		for _, c := range cmds {
			if err = c.Start(); err != nil {
				panic(err)
				return
			}
		}
		for _, c := range cmds {
			if err = c.Wait(); err != nil {
				panic(err)
				return
			}
		}
		fmt.Println(string(out.Bytes()))
	} else {
		pin, pout, err := os.Pipe()
		if err != nil {
			panic(err)
			return
		}

		cpath_begin, err := exec.LookPath(command[0][0])
		if err != nil {
			fmt.Printf("%s: No such file or directory\n", command[0][0])
			return
		}

		args_begin := command[0][0:]
		attr_begin := syscall.ProcAttr{Files: []uintptr{0, pout.Fd(), 2}}
		_, err = syscall.ForkExec(cpath_begin, args_begin, &attr_begin)
		if err != nil {
			panic(err)
			return
		}
		pout.Close()

		cpath_end, err := exec.LookPath(command[1][0])
		if err != nil {
			fmt.Printf("%s: No such file or directory\n", command[1][0])
			return
		}
		args_end := command[1][0:]
		attr_end := syscall.ProcAttr{Files: []uintptr{pin.Fd(), 1, 2}}
		pid, err := syscall.ForkExec(cpath_end, args_end, &attr_end)
		if err != nil {
			panic(err)
			return
		}

		pin.Close()
		proc, err := os.FindProcess(pid)
		status, err := proc.Wait()

		if err != nil {
			panic(err)
			return
		}
		if !status.Success() {
			status := strings.Split(status.String(), " ")
			fmt.Printf("Process %d existed with status(%s).\n", pid, status[2])
			return
		}
	}
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
			input = strings.Trim(sc.Text(), " ")
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
			if findCharacter(array, "?") {
				if findCharacter(array, ":") {
					area0 := strings.Split(input, "?")
					area1 := strings.Split(area0[1], ":")

					var result bool
					if findCharacter(array, "<") {
						cash := strings.Split(strings.Trim(area0[0], " "), "<")
						array := strings.Split(strings.Trim(cash[0], " "), " ")
						result = redirectionInput(array, strings.Trim(cash[1], " "))
					} else if findCharacter(array, ">") {
						cash := strings.Split(strings.Trim(area0[0], " "), ">")
						array := strings.Split(strings.Trim(cash[0], " "), " ")
						result = redirectionOutput(array, strings.Trim(cash[1], " "))
					} else {
						array := strings.Split(strings.Trim(area0[0], " "), " ")
						result = forkExecve(array)
					}

					if result {
						array = strings.Split(strings.Trim(area1[0], " "), " ")
					} else {
						array = strings.Split(strings.Trim(area1[1], " "), " ")
					}
					forkExecve(array)
				} else {
					fmt.Println("Syntax error.")
				}
			} else {
				// sort < orig.txt > result.txt
				if findCharacter(array, "<") && findCharacter(array, ">") {
					cash0 := strings.Split(input, "<")
					cash1 := strings.Split(strings.Trim(cash0[1], " "), ">")
					array := strings.Split(strings.Trim(cash0[0], " "), " ")
					redirectionUnion(array, strings.Trim(cash1[0], " "), strings.Trim(cash1[1], " "))
				} else if findCharacter(array, "<") {
					cash := strings.Split(input, "<")
					array := strings.Split(strings.Trim(cash[0], " "), " ")
					redirectionInput(array, strings.Trim(cash[1], " "))
				} else if findCharacter(array, ">") {
					cash := strings.Split(input, ">")
					array := strings.Split(strings.Trim(cash[0], " "), " ")
					redirectionOutput(array, strings.Trim(cash[1], " "))
				} else if findCharacter(array, "|") {
					// seq 1 3| sort -rn | cat -n
					cash := strings.Split(input, "|")
					array := make([][]string, len(cash))
					for i, a := range cash {
						array[i] = strings.Split(strings.Trim(a, " "), " ")
					}
					redirectionPipeline(array, len(cash))
				} else {
					forkExecve(array)
				}
			}
		}
		cnt++
	}
}
