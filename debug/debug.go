package debug

import (
	"log"
)

var isOutputDebug bool = false

func Debug(output bool) {
	isOutputDebug = output
}

func Println(a ...interface{}) {
	if isOutputDebug {
		log.Println(a...)
	}
}
