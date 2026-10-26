//go:build windows

package ui

import (
	"syscall"
	"unsafe"
)

// enableVT turns on ANSI colours and UTF-8 output in Windows consoles.
func enableVT() {
	k32 := syscall.NewLazyDLL("kernel32.dll")
	k32.NewProc("SetConsoleOutputCP").Call(65001)
	h, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}
	var mode uint32
	if r, _, _ := k32.NewProc("GetConsoleMode").Call(uintptr(h), uintptr(unsafe.Pointer(&mode))); r == 0 {
		return
	}
	k32.NewProc("SetConsoleMode").Call(uintptr(h), uintptr(mode|0x0004)) // ENABLE_VIRTUAL_TERMINAL_PROCESSING
}

