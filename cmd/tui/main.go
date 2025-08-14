package main

import (
	"fmt"
	"log"
	"os"

	"github.com/9d4/watui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	t := tui.New()
	p := tea.NewProgram(t)
	if _, err := p.Run(); err != nil {
		log.Fatal("Failed to start watui", err)
		os.Exit(1)
	}
}
