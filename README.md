# LazyBucket

A terminal UI for navigating and exploring Google Cloud Storage buckets, inspired by lazygit.

## Features

- Browse your GCS buckets in a terminal interface
- Navigate through subfolders in buckets
- View and interact with files
- User-friendly interface with keyboard shortcuts

## Installation

```bash
go install github.com/fernandoabolafio/lazybucket/cmd/lazybucket@latest
```

## Usage

```bash
# Make sure you have authenticated with Google Cloud
gcloud auth application-default login

# Set your Google Cloud Project ID
export GOOGLE_CLOUD_PROJECT=your-project-id

# Run the application
lazybucket

# Alternatively, specify the project ID via command line flag
lazybucket --project=your-project-id
```

### Keyboard Shortcuts

- `↑/k`: Move up
- `↓/j`: Move down
- `Enter`: Open directory
- `Backspace/b`: Go back
- `v`: View file content
- `r`: Refresh
- `?/h`: Toggle help
- `q`: Quit

## Development

```bash
# Clone the repository
git clone https://github.com/fernandoabolafio/lazybucket.git
cd lazybucket

# Build and run
go build -o lazybucket ./cmd/lazybucket
export GOOGLE_CLOUD_PROJECT=your-project-id
./lazybucket
```

## Dependencies

- [Google Cloud Storage Go SDK](https://pkg.go.dev/cloud.google.com/go/storage)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - A powerful TUI framework
- [Bubble](https://github.com/charmbracelet/bubbles) - Common TUI components for Bubble Tea
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal applications

## Authentication

LazyBucket uses Google Cloud application default credentials. Make sure you're authenticated:

```bash
gcloud auth application-default login
```

You also need to specify your Google Cloud Project ID:

```bash
export GOOGLE_CLOUD_PROJECT=your-project-id
```

## License

MIT
