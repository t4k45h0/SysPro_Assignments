package main

import (
	"bufio"
	"io"
	"os"
)

func main() {
	// open input file and make a buffered reader
	// 入力ファイルを開き、バッファ付きリーダーを作成します
	fi, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	br := bufio.NewReader(fi)
	// open output file and make a buffered writer
	// 出力ファイルを開き、バッファ付きライターを作成します
	fo, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	bw := bufio.NewWriter(fo)
	// make a buffer to read data
	buf := make([]byte, 1024)
	// copy the whole content of
	// the input file to the output file
	for {
		n, err := br.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}
		if _, err := bw.Write(buf[:n]); err != nil {
			panic(err)
		}
	}
	if err = bw.Flush(); err != nil {
		panic(err)
	}
}
