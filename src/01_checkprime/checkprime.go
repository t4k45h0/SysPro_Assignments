package main

import (
	"fmt"
	"math"
)

func isPrimeNumber(num int) []bool {
	num++
	list := make([]bool, num)
	list[0] = true
	list[1] = true
	for i := 2; i < int(math.Pow(float64(num), 0.5))+1; i += 1 {
		if !list[i] {
			for j := int(math.Pow(float64(i), 2)); j < num; j += i {
				list[j] = true
			}
		}
	}

	return list
}

func main() {
	imput := 1257787
	list := isPrimeNumber(imput)
	if !list[len(list)-1] {
		count := 0
		for _, i := range list {
			if !i {
				count += 1
			}
		}
		fmt.Printf("%d は %d 番目の素数です。\n", imput, count)
	} else {
		fmt.Printf("%d は素数ではありません。\n", imput)
	}
}
