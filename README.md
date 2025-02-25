# GBuckets

A terminal UI for navigating and exploring Google Cloud Storage buckets, inspired by lazygit.

## Features

- Browse your GCS buckets in a terminal interface
- Navigate through subfolders in buckets
- View and interact with files
- User-friendly interface with keyboard shortcuts

## Installation

```bash
go install github.com/fernandoabolafio/gbuckets/cmd/gbuckets@latest
```

## Usage

```bash
# Make sure you have authenticated with Google Cloud
gcloud auth application-default login

# Set your Google Cloud Project ID
export GOOGLE_CLOUD_PROJECT=your-project-id

# Run the application
gbuckets

# Alternatively, specify the project ID via command line flag
gbuckets --project=your-project-id
```

## Development

```bash
# Clone the repository
git clone https://github.com/fernandoabolafio/gbuckets.git
cd gbuckets

# Build and run
go build -o gbuckets ./cmd/gbuckets
export GOOGLE_CLOUD_PROJECT=your-project-id
./gbuckets
```

## Dependencies

- [Google Cloud Storage Go SDK](https://pkg.go.dev/cloud.google.com/go/storage)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - A powerful TUI framework
- [Bubble](https://github.com/charmbracelet/bubbles) - Common TUI components for Bubble Tea
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal applications

## Authentication

GBuckets uses Google Cloud application default credentials. Make sure you're authenticated:

```bash
gcloud auth application-default login
```

You also need to specify your Google Cloud Project ID:

```bash
export GOOGLE_CLOUD_PROJECT=your-project-id
```

## License

MIT
