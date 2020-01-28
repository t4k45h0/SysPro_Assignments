package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
)

var fileState syscall.Stat_t

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s port\n", os.Args[0])
		os.Exit(1)
	}
	portno := os.Args[1]
	http_server_loop(portno)
}

func http_server_loop(portno string) {
	listener, err := tcp_listen_port(portno)
	if err != nil {
		panic(err)
	}

	fi, err := syscall.Open("./quote2019_enc.wav", syscall.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := syscall.Close(fi); err != nil {
			panic(err)
		}
	}()

	_, err = syscall.Seek(fi, 78, 0)
	if err != nil {
		panic(err)
	}

	if err := syscall.Fstat(fi, &fileState); err != nil {
		panic(err)
	}

	data, err := syscall.Mmap(fi, 0, int(fileState.Size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := syscall.Munmap(data); err != nil {
			panic(err)
		}
	}()

	for {
		fmt.Println("Accepting incoming connections...")
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		go func() {
			in, out := fdopen_sock(conn)
			line, err := http_receive_request(in)
			if err == nil && line == "SYSPRO" {
				http_send_reply(out)
			} else {
				http_send_reply_bad_request(out)
				conn.Close()
				return
			}
			line, err = http_receive_request(in)
			if err == nil {
				http_send_reply2(out, data, line)
			} else {
				http_send_reply_bad_request(out)
				conn.Close()
				return
			}
			conn.Close()
		}()
	}
}

func tcp_listen_port(portno string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+portno)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}

	return listener, err
}

func fdopen_sock(conn *net.TCPConn) (*bufio.Reader, *bufio.Writer) {
	in := bufio.NewReader(conn)
	out := bufio.NewWriter(conn)

	return in, out
}

func http_receive_request(in *bufio.Reader) (string, error) {
	line, _, err := in.ReadLine()
	if err != nil {
		return "", err
	}
	fmt.Printf("requestline is [%s]\n", line)

	return string(line), nil
}

func http_send_reply(out *bufio.Writer) {
	out.WriteString("___XXXXXXX\r\n")
	out.Flush()
}

func http_send_reply2(out *bufio.Writer, data []byte, line string) {
	num, _ := strconv.ParseUint(line, 0, 8)
	for _, b := range data {
		out.WriteByte(b ^ uint8(num))
	}
	out.Flush()
}

func http_send_reply_bad_request(out *bufio.Writer) {
	out.WriteString("HTTP/1.0 400 Bad Request\r\nContent-Type: text/html\r\n\r\n")
	out.WriteString("<html><head></head><body>400 Bad Request</body></html>\n")
	out.Flush()
}
