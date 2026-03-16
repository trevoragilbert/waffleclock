# WaffleClock — Remaining Tasks

## 19c. DNS — add CNAME record in Cloudflare *(manual)*

- [ ] In Cloudflare for `trevoragilbert.com`, add a CNAME record for the `waffles` subdomain only — do not touch the apex or `www` records:
  - **Name**: `waffles`
  - **Target**: `trevoragilbert.github.io`
  - **Proxy status**: set to **DNS only** (gray cloud, not orange) — Cloudflare proxying conflicts with GitHub Pages SSL cert provisioning
- [ ] Confirm the site loads at `https://waffles.trevoragilbert.com` once GitHub verifies the domain

---

## 20. Visual Plate Counter

Replace the plain text waffle count with an illustrated plate that accumulates mini waffle icons.

### 20a. Plate element

- [ ] Replace `#waffle-count` with a `#plate-area` containing:
  - `#plate` — CSS ellipse representing a plate (white/cream fill, subtle warm-gray border)
  - `#plate-waffles` — absolutely-positioned container inside `#plate` where mini waffle icons are appended
- [ ] Plate is always visible; starts empty

### 20b. Mini waffle icon

- [ ] Define the waffle SVG as a reusable JS function that returns an SVG element
- [ ] Mini waffle size: ~28–36px, same Noun Project SVG paths, fixed golden-brown fill (`#E4BC80`)
- [ ] Each mini waffle is a `<div class="mini-waffle">` wrapping the SVG, positioned absolutely within `#plate-waffles`

### 20c. Placement logic

- [ ] On each DONE transition, call `addWaffleToPlate()`:
  - Random position within the plate bounds (constrained so icon doesn't overflow the plate edge)
  - Random rotation: `-20deg` to `+20deg`
  - Append new mini waffle to `#plate-waffles`
- [ ] Waffles stack naturally via DOM order — no collision detection needed

### 20d. Drop animation

- [ ] Mini waffle enters with a short "plop" animation:
  - Starts slightly above final position, `scale(0.6)`, `opacity: 0`
  - Animates to final position, `scale(1)`, `opacity: 1` over ~300ms ease-out
  - CSS `@keyframes waffleDrop`

### 20e. Numeric label

- [ ] Small `#waffle-count-label` below the plate showing `"N waffles"` — small, muted, Lexend

---

## 21. "Stop Waffling Around" — Session End & Summary

### 21a. Stop button

- [ ] Add `#stop-btn` — small, muted text/ghost button, visible after first "Start" tap, hidden in IDLE
- [ ] Label: **"Stop Waffling Around"**

### 21b. SUMMARY state

- [ ] Add `SUMMARY` phase — tapping stop at any active phase cancels timer and transitions here
- [ ] Summary screen shows:
  - Total waffle count large (Space Mono)
  - Singular/plural label
  - The filled plate as a recap visual
  - **"Start Over"** button — resets count to 0, clears plate, returns to IDLE

### 21c. Edge case — zero waffles

- [ ] If stopped before any waffle completes, show `"No waffles yet."` and skip the plate visual

---

## 22. Footer Credits & Links

- [ ] Add favorite waffle recipe link next to Noun Project credit, separated by ` · `
- [ ] Link: "Favorite recipe" → `https://cooking.nytimes.com/recipes/1017409-waffles`, opens in new tab
- [ ] Both links inherit the muted `#credit` style (no bright blue, no underline unless hovered)

---

## 23. Final File

- [ ] Validate HTML (no unclosed tags)
- [ ] Remove any debug `console.log` statements
- [ ] Confirm single `index.html` — no external files except Google Fonts CDN
