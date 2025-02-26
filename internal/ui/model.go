package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fernandoabolafio/lazybucket/internal/gcs"
)

// Styling
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#4A86CF")).
			Width(30).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#727272")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EEEEEE")).
				Render

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	viewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A86CF")).
			PaddingLeft(1).
			PaddingRight(1)

	detailsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A86CF")).
			Padding(1, 2).
			Width(40)

	detailsHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#4A86CF")).
				Width(36).
				Align(lipgloss.Center).
				Padding(0, 1)

	detailsLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4A86CF")).
				Bold(true)

	detailsValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EEEEEE"))

	copyMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Render
)

// KeyMap defines the keybindings for the application
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	View     key.Binding
	Help     key.Binding
	Refresh  key.Binding
	Download key.Binding
	CopyURL  key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace", "b"),
			key.WithHelp("backspace/b", "go back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		View: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view file"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?/h", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Download: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "download file"),
		),
		CopyURL: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy gsutil URL"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.View, k.Download, k.CopyURL, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Back, k.View, k.Refresh},
		{k.Download, k.CopyURL},
		{k.Help, k.Quit},
	}
}

// ListItem represents an item in the list
type ListItem struct {
	item gcs.Item
}

// FilterValue implements list.Item interface
func (i ListItem) FilterValue() string {
	return i.item.Name
}

// Title returns the item name
func (i ListItem) Title() string {
	if i.item.IsDir {
		if i.item.Name == ".." {
			return "📁 .."
		}
		if i.item.IsBucket {
			return "🪣 " + i.item.Name
		}
		return "📁 " + i.item.Name
	}
	return "📄 " + i.item.Name
}

// Description returns the item details
func (i ListItem) Description() string {
	if i.item.IsDir {
		return ""
	}
	return fmt.Sprintf("Size: %d bytes, Updated: %s", i.item.Size, i.item.Updated.Format("2006-01-02 15:04:05"))
}

// Model represents the application state
type Model struct {
	gcsClient        *gcs.Client
	list             list.Model
	help             help.Model
	viewport         viewport.Model
	keyMap           KeyMap
	currentPath      string
	pathHistory      []string
	statusMsg        string
	showHelp         bool
	viewingFile      bool
	fileContent      string
	loadingItems     bool
	ready            bool
	width            int
	height           int
	showCopyMessage  bool
	copyMessageTimer int
}

// New creates a new UI model
func New(gcsClient *gcs.Client) Model {
	// Create list
	delegate := list.NewDefaultDelegate()
	listModel := list.New([]list.Item{}, delegate, 0, 0)
	listModel.Title = "LazyBucket"
	listModel.SetShowHelp(false)
	listModel.SetShowFilter(false)
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(false)
	listModel.DisableQuitKeybindings()

	// Create help
	helpModel := help.New()
	helpModel.ShowAll = false

	// Create viewport for file viewing
	viewportModel := viewport.New(0, 0)

	// Create model
	m := Model{
		gcsClient:        gcsClient,
		list:             listModel,
		help:             helpModel,
		viewport:         viewportModel,
		keyMap:           DefaultKeyMap(),
		currentPath:      "",
		pathHistory:      []string{},
		statusMsg:        "Loading...",
		showHelp:         false,
		viewingFile:      false,
		fileContent:      "",
		loadingItems:     true,
		ready:            false,
		showCopyMessage:  false,
		copyMessageTimer: 0,
	}

	return m
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.loadItems(),
	)
}

// loadItems loads items from the current path
func (m Model) loadItems() tea.Cmd {
	return func() tea.Msg {
		if m.currentPath == "" {
			// Load buckets
			items, err := m.gcsClient.ListBuckets()
			if err != nil {
				return errMsg{err}
			}
			return itemsLoadedMsg{items}
		}

		// Load objects from bucket/prefix
		bucketName, prefix := gcs.ParsePath(m.currentPath)
		items, err := m.gcsClient.ListObjects(bucketName, prefix)
		if err != nil {
			return errMsg{err}
		}
		return itemsLoadedMsg{items}
	}
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If we're viewing a file, handle viewport keybindings
		if m.viewingFile {
			switch {
			case key.Matches(msg, m.keyMap.Quit):
				return m, tea.Quit
			case key.Matches(msg, m.keyMap.Back):
				m.viewingFile = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}

		// Handle global keybindings
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, m.keyMap.Refresh):
			m.statusMsg = "Refreshing..."
			return m, m.loadItems()
		case key.Matches(msg, m.keyMap.Back):
			if len(m.pathHistory) > 0 {
				m.currentPath = m.pathHistory[len(m.pathHistory)-1]
				m.pathHistory = m.pathHistory[:len(m.pathHistory)-1]
				m.statusMsg = "Loading items..."
				return m, m.loadItems()
			}
			m.statusMsg = "Already at root level"
			return m, nil
		case key.Matches(msg, m.keyMap.Enter):
			if len(m.list.Items()) == 0 {
				return m, nil
			}

			// Check if the selected item is nil
			selectedItem := m.list.SelectedItem()
			if selectedItem == nil {
				return m, nil
			}

			// Safely type assert with check
			selected, ok := selectedItem.(ListItem)
			if !ok {
				return m, nil
			}

			if selected.item.IsDir {
				m.statusMsg = fmt.Sprintf("Navigating to %s", selected.item.Name)
				if selected.item.Name == ".." {
					// Go up one level
					if len(m.pathHistory) > 0 {
						m.currentPath = m.pathHistory[len(m.pathHistory)-1]
						m.pathHistory = m.pathHistory[:len(m.pathHistory)-1]
					} else {
						m.currentPath = ""
					}
				} else {
					// Navigate into directory
					if m.currentPath != "" {
						m.pathHistory = append(m.pathHistory, m.currentPath)
					}
					m.currentPath = selected.item.FullPath
				}
				return m, m.loadItems()
			}
			return m, nil
		case key.Matches(msg, m.keyMap.View):
			if len(m.list.Items()) == 0 {
				return m, nil
			}

			// Check if the selected item is nil
			selectedItem := m.list.SelectedItem()
			if selectedItem == nil {
				return m, nil
			}

			// Safely type assert with check
			selected, ok := selectedItem.(ListItem)
			if !ok {
				return m, nil
			}

			if !selected.item.IsDir {
				m.statusMsg = fmt.Sprintf("Viewing %s", selected.item.Name)
				return m, m.loadFile(selected.item)
			}
			return m, nil
		case key.Matches(msg, m.keyMap.Download):
			if len(m.list.Items()) == 0 {
				return m, nil
			}

			// Check if the selected item is nil
			selectedItem := m.list.SelectedItem()
			if selectedItem == nil {
				return m, nil
			}

			// Safely type assert with check
			selected, ok := selectedItem.(ListItem)
			if !ok {
				return m, nil
			}

			if !selected.item.IsDir {
				bucketName, objectName := gcs.ParsePath(selected.item.FullPath)
				m.statusMsg = fmt.Sprintf("Downloading %s to current directory...", selected.item.Name)
				return m, m.downloadFile(bucketName, objectName, selected.item.Name)
			}
			return m, nil
		case key.Matches(msg, m.keyMap.CopyURL):
			if len(m.list.Items()) == 0 {
				return m, nil
			}

			// Check if the selected item is nil
			selectedItem := m.list.SelectedItem()
			if selectedItem == nil {
				return m, nil
			}

			// Safely type assert with check
			selected, ok := selectedItem.(ListItem)
			if !ok {
				return m, nil
			}

			if !selected.item.IsDir {
				m.statusMsg = fmt.Sprintf("Copied gsutil URL for %s", selected.item.Name)
				m.showCopyMessage = true
				m.copyMessageTimer = 10 // Show message for 10 updates
				return m, m.copyGsutilURL(selected.item.FullPath)
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// Set up list and viewport when we first get a window size
			listWidth := msg.Width / 2
			if listWidth < 40 {
				listWidth = msg.Width
			}
			m.list.SetSize(listWidth, msg.Height-4) // Reserve space for title and help
			m.viewport = viewport.New(msg.Width-2, msg.Height-4)
			m.viewport.Style = viewportStyle
			m.ready = true
		} else {
			listWidth := msg.Width / 2
			if listWidth < 40 {
				listWidth = msg.Width
			}
			m.list.SetSize(listWidth, msg.Height-4)
			m.viewport.Width = msg.Width - 2
			m.viewport.Height = msg.Height - 4
		}

		return m, nil

	case itemsLoadedMsg:
		m.loadingItems = false
		items := []list.Item{}
		for _, item := range msg.items {
			items = append(items, ListItem{item: item})
		}
		m.list.SetItems(items)
		m.statusMsg = fmt.Sprintf("Loaded %d items", len(items))

		return m, nil

	case fileLoadedMsg:
		m.fileContent = msg.content
		m.viewingFile = true
		m.viewport.SetContent(m.fileContent)
		m.viewport.GotoTop()

		return m, nil

	case errMsg:
		m.loadingItems = false
		m.statusMsg = fmt.Sprintf("Error: %v", msg.err)

		return m, nil

	case copyDoneMsg:
		// Handle copy done message
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg{}
		})

	case tickMsg:
		// Handle timer tick for copy message
		if m.copyMessageTimer > 0 {
			m.copyMessageTimer--
			if m.copyMessageTimer == 0 {
				m.showCopyMessage = false
			}
			return m, nil
		}
		return m, nil

	case downloadDoneMsg:
		m.statusMsg = fmt.Sprintf("Downloaded file to %s", msg.path)
		return m, nil
	}

	// Handle list navigation
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var s strings.Builder

	// Title and path
	pathInfo := "/"
	if m.currentPath != "" {
		pathInfo = m.currentPath
	}
	title := titleStyle.Render("LazyBucket")
	path := infoStyle.Copy().Width(m.width - lipgloss.Width(title) - 1).Render(pathInfo)
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, title, path))
	s.WriteString("\n\n")

	// Content
	if m.viewingFile {
		s.WriteString(m.viewport.View())
	} else {
		// Split view with list on left and details on right if width allows
		if m.width >= 80 {
			listView := m.list.View()
			detailsView := m.renderFileDetails()
			s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listView, detailsView))
		} else {
			s.WriteString(m.list.View())
		}
	}

	s.WriteString("\n")

	// Status message
	statusMsg := m.statusMsg
	if m.showCopyMessage {
		statusMsg = copyMessageStyle("URL copied to clipboard!")
	}
	s.WriteString(statusMessageStyle(statusMsg))

	// Help
	if m.showHelp {
		s.WriteString("\n")
		s.WriteString(helpStyle.Render(m.help.View(m.keyMap)))
	} else {
		s.WriteString("\n")
		s.WriteString(helpStyle.Render(m.help.ShortHelpView(m.keyMap.ShortHelp())))
	}

	return s.String()
}

// renderFileDetails renders the file details panel
func (m Model) renderFileDetails() string {
	if len(m.list.Items()) == 0 {
		return ""
	}

	// Check if the selected item is nil
	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return ""
	}

	// Safely type assert with check
	selected, ok := selectedItem.(ListItem)
	if !ok {
		return ""
	}

	if selected.item.IsDir {
		return ""
	}

	var s strings.Builder
	s.WriteString(detailsHeaderStyle.Render("File Details"))
	s.WriteString("\n\n")

	// File name
	s.WriteString(detailsLabelStyle.Render("Name: "))
	s.WriteString(detailsValueStyle.Render(selected.item.Name))
	s.WriteString("\n\n")

	// File size
	s.WriteString(detailsLabelStyle.Render("Size: "))
	s.WriteString(detailsValueStyle.Render(formatSize(selected.item.Size)))
	s.WriteString("\n\n")

	// Last modified
	s.WriteString(detailsLabelStyle.Render("Last Modified: "))
	s.WriteString(detailsValueStyle.Render(selected.item.Updated.Format("Jan 02, 2006 15:04:05")))
	s.WriteString("\n\n")

	// Full path
	s.WriteString(detailsLabelStyle.Render("Full Path: "))
	s.WriteString(detailsValueStyle.Render(selected.item.FullPath))
	s.WriteString("\n\n")

	// GsUtil URI
	bucketName, objectName := gcs.ParsePath(selected.item.FullPath)
	gsutilURI := fmt.Sprintf("gs://%s/%s", bucketName, objectName)
	s.WriteString(detailsLabelStyle.Render("GsUtil URI: "))
	s.WriteString(detailsValueStyle.Render(gsutilURI))
	s.WriteString("\n\n")

	// Actions
	s.WriteString(detailsLabelStyle.Render("Actions:"))
	s.WriteString("\n")
	s.WriteString(detailsValueStyle.Render("Press 'd' to download"))
	s.WriteString("\n")
	s.WriteString(detailsValueStyle.Render("Press 'c' to copy gsutil URL"))

	return detailsStyle.Render(s.String())
}

// formatSize formats the file size in a human-readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

// loadFile loads the content of a file
func (m Model) loadFile(item gcs.Item) tea.Cmd {
	return func() tea.Msg {
		bucketName, objectName := gcs.ParsePath(item.FullPath)
		content, err := m.gcsClient.GetObjectContent(bucketName, objectName)
		if err != nil {
			return errMsg{err}
		}
		return fileLoadedMsg{content}
	}
}

// downloadFile downloads a file from GCS to the local filesystem
func (m Model) downloadFile(bucketName, objectName, fileName string) tea.Cmd {
	return func() tea.Msg {
		// Create a temporary file
		f, err := os.Create(fileName)
		if err != nil {
			return errMsg{err}
		}
		defer f.Close()

		// Get the object content
		content, err := m.gcsClient.GetObjectContent(bucketName, objectName)
		if err != nil {
			return errMsg{err}
		}

		// Write the content to the file
		_, err = f.WriteString(content)
		if err != nil {
			return errMsg{err}
		}

		return downloadDoneMsg{path: fileName}
	}
}

// copyGsutilURL copies the gsutil URL to the clipboard
func (m Model) copyGsutilURL(fullPath string) tea.Cmd {
	return func() tea.Msg {
		bucketName, objectName := gcs.ParsePath(fullPath)
		gsutilURI := fmt.Sprintf("gs://%s/%s", bucketName, objectName)

		// Use the 'pbcopy' command on macOS to copy to clipboard
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(gsutilURI)
		err := cmd.Run()
		if err != nil {
			return errMsg{err}
		}

		return copyDoneMsg{}
	}
}

// Message types
type itemsLoadedMsg struct {
	items []gcs.Item
}

type fileLoadedMsg struct {
	content string
}

type errMsg struct {
	err error
}

type copyDoneMsg struct{}

type tickMsg struct{}

type downloadDoneMsg struct {
	path string
}
