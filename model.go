package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	styleBold      = lipgloss.NewStyle().Bold(true)
	styleDim       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	styleSelected  = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("255"))
	styleHeader    = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	styleFooter    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	styleError     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	styleSectionHd = lipgloss.NewStyle().Bold(true).Underline(true)
	styleSource    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// ── state ─────────────────────────────────────────────────────────────────────

type viewState int

const (
	stateLoading viewState = iota
	stateList
	stateDetail
	stateError
)

// ── messages ──────────────────────────────────────────────────────────────────

type fetchDoneMsg struct{ feed Feed }
type fetchErrMsg struct{ err error }
type clearErrMsg struct{}

// ── model ─────────────────────────────────────────────────────────────────────

type model struct {
	state        viewState
	feed         Feed
	cursor       int
	detailCursor int
	err          error
	transientErr string
	loading      bool
	width        int
	height       int
	listOffset   int
}

func (m model) Init() tea.Cmd {
	return doFetch
}

func doFetch() tea.Msg {
	feed, err := fetchFeed()
	if err != nil {
		return fetchErrMsg{err}
	}
	return fetchDoneMsg{feed}
}

// ── update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case fetchDoneMsg:
		m.feed = msg.feed
		m.state = stateList
		m.loading = false
		m.cursor = 0
		m.listOffset = 0
		m.transientErr = ""
		return m, nil

	case fetchErrMsg:
		if m.state == stateLoading {
			m.err = msg.err
			m.state = stateError
		} else {
			m.loading = false
			m.transientErr = fmt.Sprintf("Refresh failed: %v", msg.err)
			return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg { return clearErrMsg{} })
		}
		return m, nil

	case clearErrMsg:
		m.transientErr = ""
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case m.state == stateError && key.Matches(msg, keys.Enter):
			m.state = stateLoading
			m.err = nil
			return m, doFetch

		case m.state == stateList:
			return m.updateList(msg)

		case m.state == stateDetail:
			return m.updateDetail(msg)
		}
	}

	return m, nil
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(m.feed.Headlines)
	switch {
	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
			m.adjustListOffset()
		}
	case key.Matches(msg, keys.Down):
		if m.cursor < n-1 {
			m.cursor++
			m.adjustListOffset()
		}
	case key.Matches(msg, keys.Enter):
		if n > 0 {
			m.state = stateDetail
			m.detailCursor = 0
		}
	case key.Matches(msg, keys.Open):
		if n > 0 {
			_ = openBrowser(m.feed.Headlines[m.cursor].URL)
		}
	case key.Matches(msg, keys.Refresh):
		if !m.loading {
			m.loading = true
			return m, doFetch
		}
	}
	return m, nil
}

func (m model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	h := m.feed.Headlines[m.cursor]
	items := detailURLs(h)
	n := len(items)
	switch {
	case key.Matches(msg, keys.Back):
		m.state = stateList
	case key.Matches(msg, keys.Up):
		if m.detailCursor > 0 {
			m.detailCursor--
		}
	case key.Matches(msg, keys.Down):
		if m.detailCursor < n-1 {
			m.detailCursor++
		}
	case key.Matches(msg, keys.Open), key.Matches(msg, keys.Enter):
		if n > 0 && items[m.detailCursor] != "" {
			_ = openBrowser(items[m.detailCursor])
		}
	case key.Matches(msg, keys.OpenHead):
		_ = openBrowser(h.URL)
	}
	return m, nil
}

// detailURLs returns the ordered URL list for the detail view (discussions then commentary).
func detailURLs(h Headline) []string {
	var urls []string
	for _, d := range h.Discussion {
		urls = append(urls, d.URL)
	}
	for _, c := range h.Commentary {
		urls = append(urls, c.URL)
	}
	return urls
}

// adjustListOffset keeps the selected item visible.
func (m *model) adjustListOffset() {
	visible := m.visibleListRows()
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	} else if m.cursor >= m.listOffset+visible {
		m.listOffset = m.cursor - visible + 1
	}
}

func (m model) visibleListRows() int {
	// 3 lines header + 2 lines footer + 1 transient error line
	reserved := 5
	if m.transientErr != "" {
		reserved++
	}
	if m.height > reserved {
		return m.height - reserved
	}
	return 1
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	switch m.state {
	case stateLoading:
		return "\n\n  Loading techmeme.com…"
	case stateError:
		return fmt.Sprintf("\n\n  %s\n\n  %s",
			styleError.Render("Error: "+m.err.Error()),
			"Press enter to retry.")
	case stateList:
		return m.viewList()
	case stateDetail:
		return m.viewDetail()
	}
	return ""
}

func (m model) viewList() string {
	var b strings.Builder

	// Header
	loading := ""
	if m.loading {
		loading = "  refreshing…"
	}
	header := styleHeader.Render("techmeme-cli" + loading)
	hints := styleDim.Render("r: refresh  q: quit")
	gap := m.width - lipgloss.Width(header) - lipgloss.Width(hints)
	if gap < 1 {
		gap = 1
	}
	b.WriteString(header + strings.Repeat(" ", gap) + hints + "\n\n")

	// List items
	visible := m.visibleListRows()
	end := m.listOffset + visible
	if end > len(m.feed.Headlines) {
		end = len(m.feed.Headlines)
	}

	for i := m.listOffset; i < end; i++ {
		h := m.feed.Headlines[i]
		disc := len(h.Discussion)
		comm := len(h.Commentary)

		title := truncate(h.Title, m.width-4)
		meta := fmt.Sprintf("  %s · %d discussions · %d commentary", h.Source, disc, comm)
		if h.Time != "" {
			meta += "  " + h.Time
		}
		meta = truncate(meta, m.width-2)

		if i == m.cursor {
			b.WriteString(styleSelected.Render(" ▸ "+title) + "\n")
			b.WriteString(styleSelected.Render(meta) + "\n")
		} else {
			b.WriteString("   " + title + "\n")
			b.WriteString(styleDim.Render(meta) + "\n")
		}
		b.WriteString("\n")
	}

	// Transient error
	if m.transientErr != "" {
		b.WriteString(styleError.Render("  "+m.transientErr) + "\n")
	}

	// Footer
	footer := styleFooter.Render("  ↑↓ navigate · enter expand · o open · r refresh · q quit")
	b.WriteString(footer)

	return b.String()
}

func (m model) viewDetail() string {
	if m.cursor >= len(m.feed.Headlines) {
		return ""
	}
	h := m.feed.Headlines[m.cursor]
	var b strings.Builder

	// Header
	header := styleHeader.Render("techmeme-cli")
	hints := styleDim.Render("esc: back  O: open headline")
	gap := m.width - lipgloss.Width(header) - lipgloss.Width(hints)
	if gap < 1 {
		gap = 1
	}
	b.WriteString(header + strings.Repeat(" ", gap) + hints + "\n\n")

	// Headline info
	b.WriteString(" " + styleBold.Render(truncate(h.Title, m.width-2)) + "\n")
	meta := h.Source
	if h.Time != "" {
		meta += " · " + h.Time
	}
	b.WriteString(" " + styleSource.Render(meta) + "\n\n")

	// Build flat item index for cursor tracking
	idx := 0

	// Discussions
	if len(h.Discussion) > 0 {
		b.WriteString(" " + styleSectionHd.Render("Discussion:") + "\n")
		for _, d := range h.Discussion {
			title := truncate(d.Title, m.width-4)
			if idx == m.detailCursor {
				b.WriteString(styleSelected.Render(" ▸ "+title) + "\n")
				if d.Source != "" {
					b.WriteString(styleSelected.Render("   "+d.Source) + "\n")
				}
			} else {
				b.WriteString("   " + title + "\n")
				if d.Source != "" {
					b.WriteString(styleDim.Render("   "+d.Source) + "\n")
				}
			}
			b.WriteString("\n")
			idx++
		}
	}

	// Commentary
	if len(h.Commentary) > 0 {
		b.WriteString(" " + styleSectionHd.Render("Commentary:") + "\n")
		for _, c := range h.Commentary {
			text := c.Text
			if c.Author != "" && text != "" {
				text = c.Author + `: "` + truncate(text, m.width-6) + `"`
			} else if c.Author != "" {
				text = c.Author
			} else {
				text = truncate(text, m.width-4)
			}
			src := c.Source
			if idx == m.detailCursor {
				b.WriteString(styleSelected.Render(" ▸ "+text) + "\n")
				if src != "" {
					b.WriteString(styleSelected.Render("   "+src) + "\n")
				}
			} else {
				b.WriteString("   " + text + "\n")
				if src != "" {
					b.WriteString(styleDim.Render("   "+src) + "\n")
				}
			}
			b.WriteString("\n")
			idx++
		}
	}

	// Footer
	footer := styleFooter.Render("  ↑↓ navigate · enter/o open link · O open headline · esc back")
	b.WriteString(footer)

	return b.String()
}

func truncate(s string, max int) string {
	if max < 4 {
		max = 4
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
