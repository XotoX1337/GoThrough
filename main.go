package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"

	"github.com/XotoX1337/GoThrough/cmd"
)

func main() {
	openLogFile()

	log.Printf("GoThrough %s starting (args: %v)", Version, os.Args[1:])
	if err := cmd.Execute(); err != nil {
		log.Printf("error: %v", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	log.Println("exited cleanly")
}

// openLogFile sets up file-based logging so startup errors and crashes are
// visible even when the app is launched without a console (e.g. double-click).
func openLogFile() {
	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	logDir := filepath.Join(dir, "GoThrough")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return
	}
	path := filepath.Join(logDir, "gothrough.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("=== session %s ===", time.Now().Format(time.RFC3339))

	// Redirect the Windows STDERR handle so CGo / Go runtime crash messages
	// are captured when there is no console (GUI app from Explorer has NUL stderr).
	_ = windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(f.Fd()))
	os.Stderr = f
}
