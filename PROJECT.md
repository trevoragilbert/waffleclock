# techmeme-cli

A terminal UI for browsing Techmeme. Written in Go.

## Overview

`techmeme-cli` scrapes techmeme.com and presents headlines, discussion links, and related content in an interactive TUI. Content is loaded once at startup and cached in memory until the user explicitly refreshes.

## Stack

- **Language:** Go
- **TUI:** [bubbletea](https://github.com/charmbracelet/bubbletea) (with [lipgloss](https://github.com/charmbracelet/lipgloss) for styling, [bubbles](https://github.com/charmbracelet/bubbles) for common components)
- **Scraping:** [goquery](https://github.com/PuYam/goquery) + `net/http`
- **Browser open:** `open` / `xdg-open` via `os/exec`

## Data Model

```
Feed
├── Headline
│   ├── Title        string
│   ├── URL          string
│   ├── Source       string   // e.g. "The Verge", "Reuters"
│   ├── Time         string   // as displayed on Techmeme
│   ├── Discussion[] 
│   │   ├── Title    string
│   │   ├── URL      string
│   │   └── Source   string
│   └── Commentary[]
│       ├── Author   string
│       ├── Text     string   // snippet/quote shown on Techmeme
│       ├── URL      string
│       └── Source   string   // e.g. "X", "Threads", blog name
```

Content is scraped from `https://www.techmeme.com/` and parsed into this structure.

## Views

### 1. Headlines List (default view)

- Scrollable list of headlines.
- Each item shows: title, source, discussion count.
- Keys:
  - `j` / `k` or `↑` / `↓` — navigate
  - `enter` — expand headline (go to Detail View)
  - `o` — open headline URL in default browser
  - `r` — refresh (re-scrape techmeme.com)
  - `q` — quit

Example:

```
 techmeme-cli                                          r: refresh  q: quit

 ▸ OpenAI announces GPT-5 with native computer use       3h ago
   The Verge · 12 discussions · 4 commentary

   Google DeepMind releases Gemini 2.5 benchmarks         5h ago
   Reuters · 8 discussions · 2 commentary

   TSMC begins construction on Arizona fab expansion      6h ago
   Bloomberg · 5 discussions · 1 commentary

   EU reaches deal on landmark AI liability directive     7h ago
   Financial Times · 3 discussions · 3 commentary

   Stripe acquires stablecoin startup for $1.1B           9h ago
   TechCrunch · 6 discussions · 2 commentary



 ↑↓ navigate · enter expand · o open · r refresh · q quit
```

### 2. Headline Detail View

- Shows the full headline, source, and time.
- Two sections: Discussion (linked articles) and Commentary (social posts/reactions).
- Cursor navigates across both sections as a single list.
- Keys:
  - `j` / `k` or `↑` / `↓` — navigate discussion and commentary links
  - `o` — open selected link in browser
  - `enter` — open selected link in browser
  - `esc` / `backspace` — back to Headlines List
  - `O` (shift+o) — open the headline URL itself

Example:

```
 techmeme-cli                                        esc: back  O: open

 OpenAI announces GPT-5 with native computer use
 The Verge · 3h ago

 Discussion:
 ▸ Why GPT-5's computer use changes everything
   Stratechery

   GPT-5 benchmarks show marginal gains on reasoning
   Ars Technica

   OpenAI's new model raises fresh safety questions
   MIT Technology Review

   Microsoft to integrate GPT-5 across Office suite
   The Information

 Commentary:
   @sama: "GPT-5 is our most capable model yet..."
   Sam Altman (X)

   @benedictevans: "The interesting thing about GPT-5 is..."
   Benedict Evans (X)

   "This feels like the moment agents become real"
   John Gruber (Daring Fireball)


 ↑↓ navigate · enter/o open link · O open headline · esc back
```

## Behavior

- **Single fetch:** On startup, scrape techmeme.com once. All navigation is against the cached data.
- **Manual refresh:** `r` re-fetches and replaces the cache. Show a loading indicator during fetch.
- **Error handling:** If the initial fetch fails, show an error message with an option to retry. If a refresh fails, keep the existing data and show a transient error.
- **Browser open:** Use `open` (macOS), `xdg-open` (Linux), or `cmd /c start` (Windows) via `os/exec`.

## Project Structure

```
techmeme-cli/
├── main.go           // entrypoint, initialize model + start bubbletea
├── scraper.go        // fetch + parse techmeme.com HTML into Feed
├── model.go          // bubbletea model, update, view
├── keys.go           // key bindings
└── go.mod
```

## Non-Goals

- Persistent caching (disk, database).
- Configuration files.
- Techmeme River / sidebar content.
- In-terminal article rendering.