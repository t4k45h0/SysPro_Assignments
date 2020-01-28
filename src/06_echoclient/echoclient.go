package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s host port 'message'\n", os.Args[0])
		os.Exit(1)
	}
	server := os.Args[1]
	portno := os.Args[2]
	message := os.Args[3]
	echo_client_one(server, portno, message)
}

func echo_client_one(server, portno, message string) {
	conn, err := tcp_connect(server, portno)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()
	in, out := fdopen_sock(conn)
	n, err := echo_send_request(out, message+"\n")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d bytes sent [%s]\n", n, message)
	line, err := echo_receive_reply(in)
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Printf("%d bytes received. [%s]\n", len(line), strings.TrimRight(line, "\n"))
}

func echo_send_request(out *bufio.Writer, message string) (int, error) {
	n, err := out.WriteString(message)
	out.Flush()
	return n, err
}

func echo_receive_reply(in *bufio.Reader) (string, error) {
	line, err := in.ReadString('\n')
	return line, err
}

func tcp_connect(server, portno string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", strings.Join([]string{server, portno}, ":"))
	if err != nil {
		panic(err)
	}
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(err)
	}
	return tcpConn, err
}

func fdopen_sock(conn *net.TCPConn) (*bufio.Reader, *bufio.Writer) {
	in := bufio.NewReader(conn)
	out := bufio.NewWriter(conn)
	return in, out
}
