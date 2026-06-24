package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/XotoX1337/GoThrough/cmd"
)

func main() {
	openLogFile()

	log.Printf("GoThrough %s starting (args: %v)", Version, os.Args[1:])
	cmd.SetVersion(Version)
	if err := cmd.Execute(); err != nil {
		log.Printf("error: %v", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	log.Println("exited cleanly")
}

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
	redirectStderr(f)
}
