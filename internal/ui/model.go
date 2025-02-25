package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fernandoabolafio/gbuckets/internal/gcs"
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
)

// KeyMap defines the keybindings for the application
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Quit    key.Binding
	View    key.Binding
	Help    key.Binding
	Refresh key.Binding
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
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.View, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Back, k.View, k.Refresh},
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
	gcsClient    *gcs.Client
	list         list.Model
	help         help.Model
	viewport     viewport.Model
	keyMap       KeyMap
	currentPath  string
	pathHistory  []string
	statusMsg    string
	showHelp     bool
	viewingFile  bool
	fileContent  string
	loadingItems bool
	ready        bool
	width        int
	height       int
}

// New creates a new model
func New(gcsClient *gcs.Client) Model {
	keyMap := DefaultKeyMap()
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "GBuckets"
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	helpModel := help.New()
	helpModel.ShowAll = false

	m := Model{
		gcsClient:   gcsClient,
		list:        l,
		help:        helpModel,
		keyMap:      keyMap,
		currentPath: "",
		pathHistory: []string{},
		showHelp:    false,
		viewingFile: false,
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

			selected := m.list.SelectedItem().(ListItem)
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

			selected := m.list.SelectedItem().(ListItem)
			if !selected.item.IsDir {
				m.statusMsg = fmt.Sprintf("Viewing %s", selected.item.Name)
				return m, m.loadFile(selected.item)
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// Set up list and viewport when we first get a window size
			m.list.SetSize(msg.Width, msg.Height-4) // Reserve space for title and help
			m.viewport = viewport.New(msg.Width-2, msg.Height-4)
			m.viewport.Style = viewportStyle
			m.ready = true
		} else {
			m.list.SetSize(msg.Width, msg.Height-4)
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

	title := titleStyle.Render("GBuckets")
	path := infoStyle.Copy().Width(m.width - lipgloss.Width(title) - 1).Render(pathInfo)
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, title, path))
	s.WriteString("\n\n")

	// Content
	if m.viewingFile {
		s.WriteString(m.viewport.View())
	} else {
		s.WriteString(m.list.View())
	}

	s.WriteString("\n")

	// Status message
	s.WriteString(statusMessageStyle(m.statusMsg))

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
