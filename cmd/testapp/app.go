package main

import (
	"fmt"
	"unsafe"
)

func main() {
	e := [...]uint8{11, 22, 33}
	p := &e[0]

	fmt.Println(*p)

	c := (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer((p))) + 1))
	fmt.Println(*c)
}
