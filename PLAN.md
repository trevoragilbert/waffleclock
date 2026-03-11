# PLAN.md

## Implementation Plan — techmeme-cli

Each step is self-contained and executable in order. Files go in the repo root.

---

### Step 1 — Initialize Go module

```bash
go mod init github.com/trevoragilbert/techmeme-cli
```

---

### Step 2 — Add dependencies

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get github.com/PuYam/goquery
```

---

### Step 3 — Create `types.go`

Define all shared data types used across the project.

```go
package main

type Commentary struct {
    Author string
    Text   string
    URL    string
    Source string
}

type Discussion struct {
    Title  string
    URL    string
    Source string
}

type Headline struct {
    Title       string
    URL         string
    Source      string
    Time        string
    Discussion  []Discussion
    Commentary  []Commentary
}

type Feed struct {
    Headlines []Headline
}
```

---

### Step 4 — Create `scraper.go`

Fetch and parse techmeme.com HTML into a `Feed`.

**Function:** `func fetchFeed() (Feed, error)`

Implementation notes:
- GET `https://www.techmeme.com/` with a `User-Agent` header (e.g. `"Mozilla/5.0 (compatible; techmeme-cli)"`) to avoid being blocked.
- Use `goquery` to parse the response body.
- Techmeme HTML structure (as of 2025):
  - Each story cluster is a `div.item` containing:
    - `a.ourh` — the main headline anchor; `.text()` = title, `[href]` = URL
    - `span.src` inside `a.ourh` parent — source name
    - `span.time` — relative time string (e.g. `"3h ago"`)
    - `div.ii` — discussion block; each `a.ii` inside is one discussion link
      - `.text()` on the `a.ii` = discussion title
      - `[href]` = URL
      - `span.src` inside = source
    - `div.cmtt` — commentary block; each `.cmtt` item contains:
      - `.cmttauthor` — author name
      - `.cmttxt` — text snippet
      - `a` — URL and source (source is the domain or publication name)
  - Skip any `div.item` that has no `a.ourh` (ads, separators).
- Return a populated `Feed` or an error if the HTTP request fails or returns non-200.

---

### Step 5 — Create `keys.go`

Define all key bindings using `bubbles/key`.

```go
package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
    Up       key.Binding
    Down     key.Binding
    Enter    key.Binding
    Open     key.Binding
    OpenHead key.Binding
    Refresh  key.Binding
    Back     key.Binding
    Quit     key.Binding
}

var keys = keyMap{
    Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
    Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
    Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "expand")),
    Open:     key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open")),
    OpenHead: key.NewBinding(key.WithKeys("O"), key.WithHelp("O", "open headline")),
    Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
    Back:     key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
    Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}
```

---

### Step 6 — Create `browser.go`

Open a URL in the system default browser.

```go
package main

import (
    "os/exec"
    "runtime"
)

func openBrowser(url string) error {
    var cmd string
    var args []string
    switch runtime.GOOS {
    case "darwin":
        cmd = "open"
        args = []string{url}
    case "linux":
        cmd = "xdg-open"
        args = []string{url}
    case "windows":
        cmd = "cmd"
        args = []string{"/c", "start", url}
    default:
        return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
    return exec.Command(cmd, args...).Start()
}
```

---

### Step 7 — Create `model.go`

The bubbletea `Model` and all UI logic.

#### 7a — State enum and model struct

```go
package main

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type viewState int

const (
    stateLoading viewState = iota
    stateList
    stateDetail
    stateError
)

type model struct {
    state        viewState
    feed         Feed
    cursor       int        // selected headline index (list view)
    detailCursor int        // selected item index in detail view
    err          error      // startup or current error
    transientErr string     // shown briefly on refresh failure, then cleared
    loading      bool       // true while re-fetching
    width        int
    height       int
}
```

#### 7b — Messages

```go
type fetchDoneMsg struct{ feed Feed }
type fetchErrMsg  struct{ err error }
type clearErrMsg  struct{}
```

#### 7c — `Init()`

Return a command that calls `fetchFeed()` in a goroutine and returns either `fetchDoneMsg` or `fetchErrMsg`.

```go
func (m model) Init() tea.Cmd {
    return func() tea.Msg {
        feed, err := fetchFeed()
        if err != nil {
            return fetchErrMsg{err}
        }
        return fetchDoneMsg{feed}
    }
}
```

#### 7d — `Update()`

Handle the following messages and key presses:

| Message / Key | Action |
|---|---|
| `tea.WindowSizeMsg` | Store `width` / `height` |
| `fetchDoneMsg` | Set `feed`, switch to `stateList`, reset `cursor` |
| `fetchErrMsg` (during init) | Set `err`, switch to `stateError` |
| `fetchErrMsg` (during refresh) | Set `transientErr`, stop loading, return `clearErrMsg` after 3s |
| `clearErrMsg` | Clear `transientErr` |
| `keys.Quit` | `tea.Quit` |
| `keys.Up` | Decrement cursor (clamp to 0) |
| `keys.Down` | Increment cursor (clamp to max) |
| `keys.Enter` (stateList) | Switch to `stateDetail`, reset `detailCursor` to 0 |
| `keys.Open` (stateList) | Open `feed.Headlines[cursor].URL` in browser |
| `keys.Open` (stateDetail) | Open selected discussion/commentary URL in browser |
| `keys.OpenHead` (stateDetail) | Open `feed.Headlines[cursor].URL` in browser |
| `keys.Back` (stateDetail) | Switch back to `stateList` |
| `keys.Refresh` | Set `loading = true`, fire fetch command |
| `keys.Enter` (stateError) | Reset, fire fetch command again (retry) |

In `stateDetail`, the cursor navigates a flat list: `Discussion[0..n]` then `Commentary[0..m]`. Build a helper `func detailItems(h Headline) []string` that returns URLs for each item in order.

#### 7e — `View()`

Render based on `m.state`:

- **`stateLoading`:** Centered text `"Loading techmeme.com…"`
- **`stateError`:** Error message + `"Press enter to retry."`
- **`stateList`:** Header bar, scrollable headline list, footer help bar
- **`stateDetail`:** Header bar, headline title/source/time, Discussion section, Commentary section, footer help bar

**Styles (lipgloss):**
- Header: bold, full width, with right-aligned key hints
- Selected item: highlighted background (e.g. `lipgloss.Color("62")`)
- Source/time: dimmed (`lipgloss.Color("241")`)
- Section headers ("Discussion:", "Commentary:"): bold, underline
- Footer help: dimmed, full width
- Transient error: red foreground, shown above footer

**Scrolling in list view:** Maintain a viewport offset so the selected item is always visible. Use `m.height` to compute how many items fit.

---

### Step 8 — Create `main.go`

```go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    m := model{state: stateLoading}
    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```

---

### Step 9 — Build and smoke-test

```bash
go build -o techmeme-cli .
./techmeme-cli
```

Verify:
- [ ] App loads and shows headlines list
- [ ] `j`/`k` navigate the list
- [ ] `enter` opens detail view with discussions and commentary
- [ ] `o` opens the selected URL in the browser
- [ ] `O` in detail view opens the headline URL
- [ ] `esc` returns to list
- [ ] `r` re-fetches and updates the list
- [ ] `q` quits cleanly
- [ ] Killing network mid-refresh shows transient error and keeps existing data

---

### Step 10 — Commit and push

```bash
git add .
git commit -m "feat: initial implementation of techmeme-cli"
git push
```
