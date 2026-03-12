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
	detailOffset int // first visible item index in detail view
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
			m.detailOffset = 0
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
			m.adjustDetailOffset(h)
		}
	case key.Matches(msg, keys.Down):
		if m.detailCursor < n-1 {
			m.detailCursor++
			m.adjustDetailOffset(h)
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

// detailURLs returns the ordered URL list for the detail view:
// main article first, then discussions, then commentary.
func detailURLs(h Headline) []string {
	urls := []string{h.URL}
	for _, d := range h.Discussion {
		urls = append(urls, d.URL)
	}
	for _, c := range h.Commentary {
		urls = append(urls, c.URL)
	}
	return urls
}

// itemLines returns the number of terminal lines an item occupies:
// wrapped title lines + 1 meta line + 1 blank line.
func (m model) itemLines(i int) int {
	return len(wordWrap(m.feed.Headlines[i].Title, 16)) + 2
}

// availableRows returns the rows usable by list items.
func (m model) availableRows() int {
	reserved := 3 // header (2 lines) + footer (1 line)
	if m.transientErr != "" {
		reserved++
	}
	if m.height > reserved {
		return m.height - reserved
	}
	return 1
}

// detailItemLines returns line counts for each navigable item in the detail view.
// The first item of each section absorbs the section header line (+1) so that
// adjustDetailOffset and viewDetail share the same budget.
// Order: original article, discussions, commentary.
func detailItemLines(h Headline) []int {
	lines := []int{1 + 1} // original article line + blank
	for i, d := range h.Discussion {
		srcLines := 0
		if d.Source != "" {
			srcLines = 1
		}
		cost := len(wordWrap(d.Title, 16)) + srcLines + 1 // +1 blank
		if i == 0 {
			cost++ // "Discussion:" header
		}
		lines = append(lines, cost)
	}
	for i, c := range h.Commentary {
		raw := c.Text
		if c.Author != "" {
			raw = c.Author + `: "` + c.Text + `"`
		}
		srcLines := 0
		if c.Source != "" {
			srcLines = 1
		}
		cost := len(wordWrap(raw, 16)) + srcLines + 1 // +1 blank
		if i == 0 {
			cost++ // "Commentary:" header
		}
		lines = append(lines, cost)
	}
	return lines
}

// adjustDetailOffset keeps the detail cursor visible.
func (m *model) adjustDetailOffset(h Headline) {
	if m.detailCursor < m.detailOffset {
		m.detailOffset = m.detailCursor
		return
	}
	// header (5 lines) + footer (1)
	available := m.height - 6
	if available < 1 {
		available = 1
	}
	itemLines := detailItemLines(h)
	for {
		used := 0
		visible := false
		for i := m.detailOffset; i < len(itemLines); i++ {
			if used+itemLines[i] > available {
				break
			}
			used += itemLines[i]
			if i == m.detailCursor {
				visible = true
				break
			}
		}
		if visible {
			break
		}
		m.detailOffset++
	}
}

// adjustListOffset keeps the selected item visible by advancing or
// retreating listOffset based on actual per-item line counts.
func (m *model) adjustListOffset() {
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
		return
	}
	available := m.availableRows()
	for {
		used := 0
		visible := false
		for i := m.listOffset; i < len(m.feed.Headlines); i++ {
			lines := m.itemLines(i)
			if used+lines > available {
				break
			}
			used += lines
			if i == m.cursor {
				visible = true
				break
			}
		}
		if visible {
			break
		}
		m.listOffset++
	}
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

	// List items — stop when adding the next item would overflow.
	available := m.availableRows()
	used := 0
	for i := m.listOffset; i < len(m.feed.Headlines); i++ {
		if used+m.itemLines(i) > available {
			break
		}
		used += m.itemLines(i)
		h := m.feed.Headlines[i]
		disc := len(h.Discussion)
		comm := len(h.Commentary)

		titleLines := wordWrap(h.Title, 16)
		meta := fmt.Sprintf("  %s · %d discussions · %d commentary", h.Source, disc, comm)
		if h.Time != "" {
			meta += "  " + h.Time
		}
		meta = truncate(meta, m.width-2)

		if i == m.cursor {
			b.WriteString("> " + styleBold.Render(titleLines[0]) + "\n")
			for _, l := range titleLines[1:] {
				b.WriteString("  " + styleBold.Render(l) + "\n")
			}
			b.WriteString("  " + styleDim.Render(meta) + "\n")
		} else {
			b.WriteString("  " + titleLines[0] + "\n")
			for _, l := range titleLines[1:] {
				b.WriteString("  " + l + "\n")
			}
			b.WriteString("  " + styleDim.Render(meta) + "\n")
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

	// Static header = 5 lines (header row + blank + title + source + blank).
	// Footer = 1 line. available = height - 6.
	available := m.height - 6
	if available < 1 {
		available = 1
	}
	usedRows := 0

	// renderItem writes one navigable item if it fits, returns false when full.
	// sectionHeader is written (and charged) only for the first item (idx==sectionStart).
	renderItem := func(idx int, textLines []string, source, sectionHeader string) bool {
		if idx < m.detailOffset {
			return true // not yet in viewport, skip
		}
		srcLines := 0
		if source != "" {
			srcLines = 1
		}
		hdrLines := 0
		if sectionHeader != "" {
			hdrLines = 1
		}
		cost := len(textLines) + srcLines + 1 + hdrLines // +1 blank
		if usedRows+cost > available {
			return false
		}
		usedRows += cost
		if sectionHeader != "" {
			b.WriteString(" " + styleSectionHd.Render(sectionHeader) + "\n")
		}
		prefix, style := "  ", styleBold
		if idx == m.detailCursor {
			prefix = "> "
		} else {
			style = lipgloss.NewStyle()
		}
		b.WriteString(prefix + style.Render(textLines[0]) + "\n")
		for _, l := range textLines[1:] {
			b.WriteString("  " + style.Render(l) + "\n")
		}
		if source != "" {
			b.WriteString("  " + styleDim.Render(source) + "\n")
		}
		b.WriteString("\n")
		return true
	}

	idx := 0
	// Item 0: original article
	label := "Original article — " + h.Source
	if !renderItem(idx, []string{label}, "", "") {
		goto footer
	}
	idx++

	// Discussions
	for i, d := range h.Discussion {
		hdr := ""
		if i == 0 {
			hdr = "Discussion:"
		}
		if !renderItem(idx, wordWrap(d.Title, 16), d.Source, hdr) {
			goto footer
		}
		idx++
	}

	// Commentary
	for i, c := range h.Commentary {
		hdr := ""
		if i == 0 {
			hdr = "Commentary:"
		}
		raw := c.Text
		if c.Author != "" {
			raw = c.Author + `: "` + c.Text + `"`
		}
		if !renderItem(idx, wordWrap(raw, 16), c.Source, hdr) {
			goto footer
		}
		idx++
	}

footer:
	footer := styleFooter.Render("  ↑↓ navigate · enter/o open link · esc back")
	b.WriteString(footer)

	return b.String()
}

// wordWrap splits s into lines of at most wordsPerLine words.
func wordWrap(s string, wordsPerLine int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	for i := 0; i < len(words); i += wordsPerLine {
		end := i + wordsPerLine
		if end > len(words) {
			end = len(words)
		}
		lines = append(lines, strings.Join(words[i:end], " "))
	}
	return lines
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
