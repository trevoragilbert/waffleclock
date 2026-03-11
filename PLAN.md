# PLAN.md

## To Do

- [ ] Initialize Go module (`go mod init github.com/trevorgilbert/techmeme-cli`)
- [ ] Add dependencies: bubbletea, lipgloss, bubbles, goquery (`go get`)
- [ ] Create `scraper.go` — fetch `https://www.techmeme.com/`, parse HTML into `Feed` / `Headline` / `Discussion` / `Commentary` structs
- [ ] Create `model.go` — bubbletea `Model` with two states: Headlines List and Headline Detail
- [ ] Create `keys.go` — define key bindings (`j/k`, `↑/↓`, `enter`, `o`, `O`, `r`, `esc`, `backspace`, `q`)
- [ ] Implement Headlines List view — scrollable list showing title, source, discussion count, time
- [ ] Implement Headline Detail view — full headline info, unified Discussion + Commentary list
- [ ] Implement browser open — `open` (macOS), `xdg-open` (Linux), `cmd /c start` (Windows) via `os/exec`
- [ ] Implement `r` refresh — re-scrape with loading indicator; on failure keep existing data and show transient error
- [ ] Implement startup error state — if initial fetch fails, show error with retry option
- [ ] Create `main.go` — entrypoint, initialize model, start bubbletea program
- [ ] Create `go.mod`
