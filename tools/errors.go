package tools

import (
	"fmt"
	"runtime"
)

func Recover() any {
	var p any
	if p = recover(); p != nil {
		fmt.Println(p)
		var buf [4096]byte
		fmt.Printf(string(buf[:runtime.Stack(buf[:], false)]))
	}
	return p
}
