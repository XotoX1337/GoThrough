//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// redirectStderr points the Windows STDERR handle at f so that CGo and Go
// runtime crash messages are captured in the log file when there is no console
// (a GUI app launched from Windows Explorer receives NUL as its stderr handle).
func redirectStderr(f *os.File) {
	_ = windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(f.Fd()))
	os.Stderr = f
}
