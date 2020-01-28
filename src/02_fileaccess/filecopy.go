package main

import (
	"fmt"
	"os"
	"syscall"
)

var fileState syscall.Stat_t

func main() {
	// 入力ファイルを開き、バッファ付きリーダーを作成します
	fi, err := syscall.Open(os.Args[1], syscall.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	// main関数が終了する際に実行すべき処理を記述します
	defer func() {
		if err := syscall.Close(fi); err != nil {
			panic(err)
		}
	}()

	// 入力したファイルのサイズを取得します
	if err := syscall.Fstat(fi, &fileState); err != nil {
		panic(err)
	} else {
		fmt.Println("入力したファイルのサイズ: ", fileState.Size)
	}

	// 出力ファイルを開き、バッファ付きライターを作成します
	fo, err := syscall.Open(os.Args[2],
		syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	// main関数が終了する際に実行すべき処理を記述します
	defer func() {
		if err := syscall.Close(fo); err != nil {
			panic(err)
		}
	}()

	// ファイルをメモリにマップする
	data, err := syscall.Mmap(fi, 0, int(fileState.Size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}

	// main関数が終了する際に実行すべき処理を記述します
	defer func() {
		if err := syscall.Munmap(data); err != nil {
			panic(err)
		}
	}()

	// 半角英数, 半角スペース, 改行以外の文字を読み飛ばしてコピーする
	s := string(data)
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || (c == ' ') || (c == '\n') {
			if _, err := syscall.Write(fo, []byte(string(c))); err != nil {
				panic(err)
			}
		}
	}

	// 出力したファイルのサイズを取得します
	if err := syscall.Fstat(fo, &fileState); err != nil {
		panic(err)
	} else {
		fmt.Println("出力したファイルのサイズ: ", fileState.Size)
	}
}
