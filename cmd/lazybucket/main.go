package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fernandoabolafio/lazybucket/internal/gcs"
	"github.com/fernandoabolafio/lazybucket/internal/ui"
)

func main() {
	// Parse command line flags
	var projectID string
	flag.StringVar(&projectID, "project", "", "Google Cloud Project ID")
	flag.Parse()

	// Check if project ID is provided via environment variable if not specified via flag
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	// If still empty, prompt the user
	if projectID == "" {
		fmt.Println("Error: Google Cloud Project ID is required.")
		fmt.Println("Please provide it using one of the following methods:")
		fmt.Println("1. Command line flag: ./lazybucket --project=your-project-id")
		fmt.Println("2. Environment variable: export GOOGLE_CLOUD_PROJECT=your-project-id")
		os.Exit(1)
	}

	// Create a context that will be canceled on ctrl+c
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		os.Exit(0)
	}()

	// Create GCS client with project ID
	gcsClient, err := gcs.NewClient(ctx, projectID)
	if err != nil {
		fmt.Printf("Error creating GCS client: %v\n", err)
		fmt.Println("Make sure you are authenticated with Google Cloud:")
		fmt.Println("  gcloud auth application-default login")
		os.Exit(1)
	}
	defer gcsClient.Close()

	// Create and start UI
	model := ui.New(gcsClient)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
