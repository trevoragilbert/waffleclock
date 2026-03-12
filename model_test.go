package main

import (
	"strings"
	"testing"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func makeHeadline(title string) Headline {
	return Headline{Title: title, URL: "https://example.com", Source: "Example"}
}

func makeHeadlineWithDiscussion(title string, nDisc int) Headline {
	h := makeHeadline(title)
	for i := range nDisc {
		h.Discussion = append(h.Discussion, Discussion{
			Title:  strings.TrimSpace(strings.Repeat("word ", 20)),
			URL:    "https://example.com/disc",
			Source: "Example",
		})
		_ = i
	}
	return h
}

func makeHeadlineWithCommentary(nComm int) Headline {
	h := makeHeadline("Short title")
	for i := range nComm {
		h.Commentary = append(h.Commentary, Commentary{
			Author: "Author",
			Text:   strings.TrimSpace(strings.Repeat("word ", 20)),
			URL:    "https://example.com/comm",
			Source: "LinkedIn",
		})
		_ = i
	}
	return h
}

func makeModel(headlines []Headline, width, height int) model {
	return model{
		state:  stateList,
		feed:   Feed{Headlines: headlines},
		width:  width,
		height: height,
	}
}

// ── wordWrap ──────────────────────────────────────────────────────────────────

func TestWordWrap_ShortTitle(t *testing.T) {
	lines := wordWrap("Short title", 16)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0] != "Short title" {
		t.Fatalf("unexpected content: %q", lines[0])
	}
}

func TestWordWrap_ExactlyOneLineOfWords(t *testing.T) {
	// 16 words exactly
	words := strings.Repeat("word ", 16)
	lines := wordWrap(strings.TrimSpace(words), 16)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for 16 words, got %d", len(lines))
	}
}

func TestWordWrap_WrapsAt16Words(t *testing.T) {
	// 17 words → 2 lines
	words := strings.Repeat("word ", 17)
	lines := wordWrap(strings.TrimSpace(words), 16)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines for 17 words, got %d", len(lines))
	}
	if len(strings.Fields(lines[0])) != 16 {
		t.Fatalf("first line should have 16 words, got %d", len(strings.Fields(lines[0])))
	}
	if len(strings.Fields(lines[1])) != 1 {
		t.Fatalf("second line should have 1 word, got %d", len(strings.Fields(lines[1])))
	}
}

func TestWordWrap_32Words(t *testing.T) {
	words := strings.Repeat("word ", 32)
	lines := wordWrap(strings.TrimSpace(words), 16)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines for 32 words, got %d", len(lines))
	}
}

func TestWordWrap_Empty(t *testing.T) {
	lines := wordWrap("", 16)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for empty string, got %d", len(lines))
	}
}

// ── itemLines ─────────────────────────────────────────────────────────────────

func TestItemLines_ShortTitle(t *testing.T) {
	m := makeModel([]Headline{makeHeadline("Short title")}, 80, 40)
	// 1 title line + 1 meta + 1 blank = 3
	if got := m.itemLines(0); got != 3 {
		t.Fatalf("expected 3 lines for short title, got %d", got)
	}
}

func TestItemLines_17WordTitle(t *testing.T) {
	title := strings.TrimSpace(strings.Repeat("word ", 17))
	m := makeModel([]Headline{makeHeadline(title)}, 80, 40)
	// 2 title lines + 1 meta + 1 blank = 4
	if got := m.itemLines(0); got != 4 {
		t.Fatalf("expected 4 lines for 17-word title, got %d", got)
	}
}

func TestItemLines_32WordTitle(t *testing.T) {
	title := strings.TrimSpace(strings.Repeat("word ", 32))
	m := makeModel([]Headline{makeHeadline(title)}, 80, 40)
	// 2 title lines + 1 meta + 1 blank = 4
	if got := m.itemLines(0); got != 4 {
		t.Fatalf("expected 4 lines for 32-word title, got %d", got)
	}
}

// ── viewList line count ───────────────────────────────────────────────────────

// countLines counts rendered terminal lines, stripping ANSI codes.
func countLines(s string) int {
	return strings.Count(s, "\n")
}

func TestViewList_DoesNotExceedTerminalHeight(t *testing.T) {
	titles := []string{
		"Short title one",
		strings.TrimSpace(strings.Repeat("word ", 20)), // wraps to 2 title lines
		"Short title three",
		strings.TrimSpace(strings.Repeat("word ", 20)),
		"Short title five",
		strings.TrimSpace(strings.Repeat("word ", 20)),
		"Short title seven",
		"Short title eight",
		"Short title nine",
		"Short title ten",
	}
	var headlines []Headline
	for _, t := range titles {
		headlines = append(headlines, makeHeadline(t))
	}

	for _, height := range []int{10, 20, 24, 40, 80} {
		m := makeModel(headlines, 120, height)
		m.state = stateList
		output := m.viewList()
		got := countLines(output)
		if got > height {
			t.Errorf("height=%d: rendered %d lines, exceeds terminal height", height, got)
		}
	}
}

func TestViewList_CursorAlwaysVisible(t *testing.T) {
	var headlines []Headline
	for i := range 20 {
		title := strings.TrimSpace(strings.Repeat("word ", 10+i%10))
		headlines = append(headlines, makeHeadline(title))
		_ = i
	}

	m := makeModel(headlines, 120, 24)
	m.state = stateList

	for cursor := range len(headlines) {
		m.cursor = cursor
		m.listOffset = 0
		m.adjustListOffset()
		output := m.viewList()
		// The selected item renders with "> " prefix — verify it appears.
		if !strings.Contains(output, "> ") {
			t.Errorf("cursor=%d: selected indicator not found in output", cursor)
		}
	}
}

// ── adjustListOffset ──────────────────────────────────────────────────────────

func TestAdjustListOffset_CursorAboveOffset(t *testing.T) {
	headlines := make([]Headline, 10)
	for i := range headlines {
		headlines[i] = makeHeadline("Short title")
	}
	m := makeModel(headlines, 80, 24)
	m.listOffset = 5
	m.cursor = 2
	m.adjustListOffset()
	if m.listOffset != 2 {
		t.Fatalf("expected listOffset=2, got %d", m.listOffset)
	}
}

func TestAdjustListOffset_CursorBelowViewport(t *testing.T) {
	var headlines []Headline
	for range 20 {
		headlines = append(headlines, makeHeadline("Short title"))
	}
	m := makeModel(headlines, 80, 24)
	m.cursor = 19
	m.adjustListOffset()

	// After adjustment, the cursor item must be renderable within available rows.
	available := m.availableRows()
	used := 0
	found := false
	for i := m.listOffset; i < len(headlines); i++ {
		lines := m.itemLines(i)
		if used+lines > available {
			break
		}
		used += lines
		if i == m.cursor {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("cursor %d not visible after adjustListOffset (offset=%d, available=%d)",
			m.cursor, m.listOffset, available)
	}
}

// ── detail view scroll ────────────────────────────────────────────────────────

func TestViewDetail_DoesNotExceedTerminalHeight(t *testing.T) {
	h := makeHeadlineWithCommentary(15)
	h.Discussion = []Discussion{
		{Title: strings.TrimSpace(strings.Repeat("word ", 20)), URL: "https://example.com", Source: "Src"},
		{Title: strings.TrimSpace(strings.Repeat("word ", 20)), URL: "https://example.com", Source: "Src"},
		{Title: strings.TrimSpace(strings.Repeat("word ", 20)), URL: "https://example.com", Source: "Src"},
	}
	m := makeModel([]Headline{h}, 120, 24)
	m.state = stateDetail

	output := m.viewDetail()
	got := countLines(output)
	if got > 24 {
		t.Errorf("rendered %d lines, exceeds terminal height of 24", got)
	}
}

func TestAdjustDetailOffset_CursorAlwaysVisible(t *testing.T) {
	h := makeHeadlineWithCommentary(10)
	h.Discussion = []Discussion{
		{Title: "Disc one", URL: "https://example.com", Source: "Src"},
		{Title: "Disc two", URL: "https://example.com", Source: "Src"},
		{Title: "Disc three", URL: "https://example.com", Source: "Src"},
	}
	m := makeModel([]Headline{h}, 120, 24)
	m.state = stateDetail

	total := 1 + len(h.Discussion) + len(h.Commentary)
	for cursor := range total {
		m.detailCursor = cursor
		m.detailOffset = 0
		m.adjustDetailOffset(h)
		output := m.viewDetail()
		if !strings.Contains(output, "> ") {
			t.Errorf("cursor=%d: selected indicator not visible in detail view", cursor)
		}
		if got := countLines(output); got > 24 {
			t.Errorf("cursor=%d: rendered %d lines, exceeds height 24", cursor, got)
		}
	}
}

func TestAdjustListOffset_CursorAlwaysFitsForAllPositions(t *testing.T) {
	var headlines []Headline
	for i := range 30 {
		// Mix short and long titles
		words := 8 + (i%3)*10
		headlines = append(headlines, makeHeadline(strings.TrimSpace(strings.Repeat("word ", words))))
	}

	for _, height := range []int{15, 24, 40} {
		for cursor := range len(headlines) {
			m := makeModel(headlines, 120, height)
			m.cursor = cursor
			m.adjustListOffset()

			available := m.availableRows()
			used := 0
			found := false
			for i := m.listOffset; i < len(headlines); i++ {
				lines := m.itemLines(i)
				if used+lines > available {
					break
				}
				used += lines
				if i == m.cursor {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("height=%d cursor=%d: not visible after adjust (offset=%d)",
					height, cursor, m.listOffset)
			}
		}
	}
}
