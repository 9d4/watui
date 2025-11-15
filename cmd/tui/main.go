package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/9d4/watui/chatstore"
	"github.com/9d4/watui/internal/tui"
	"github.com/9d4/watui/wa"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	devMode := flag.Bool("dev", false, "enable developer notifications")
	flag.Parse()

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	logger, err := wa.CreateFileLogger("watui.log")
	if err != nil {
		log.Fatalf("cannot open file for log: %v", err)
	}

	roomStore, err := chatstore.New("watui.db")
	if err != nil {
		log.Fatalf("cannot init room store: %v", err)
	}
	defer roomStore.Close()

	m := wa.NewManager(logger)
	t := tui.New(m, roomStore, *devMode)
	p := tea.NewProgram(t)
	if _, err := p.Run(); err != nil {
		log.Fatal("Failed to start watui", err)
		os.Exit(1)
	}
}
