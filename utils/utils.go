package utils

import (
	"runtime"
)

const stackSize = 8192 //bytes

func GetCurrentStack() []byte {
	buf := make([]byte, stackSize)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
