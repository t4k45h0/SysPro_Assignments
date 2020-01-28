package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func main() {
	pid, r2, _ := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	// fixup for macOS
	if runtime.GOOS == "darwin" && r2 == 1 {
		pid = 0
	}

	if pid > 0 {
		// parent process
		var rusage syscall.Rusage
		status := syscall.WaitStatus(0)
		_, err := syscall.Wait4(int(pid), &status,
			syscall.WSTOPPED, &rusage)
		if err != nil {
			panic(err)
		}
		if status != 0 {
			fmt.Println("exit status", status)
		}
		os.Exit(0)
	}
	// child process
	cpath, err := exec.LookPath(os.Args[1])
	if err != nil {
		log.Fatalf("%s not found in $PATH.", os.Args[1])
	}
	args := os.Args[1:]
	err = syscall.Exec(cpath, args, os.Environ())
	panic(err)
}
