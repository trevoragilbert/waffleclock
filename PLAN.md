# waffleflip — Build Plan

## Overview

Single `index.html` file. No build step, no dependencies. Tasks are ordered by dependency — work top to bottom.

---

## 1. HTML Skeleton

- [ ] Create `index.html` with `<!DOCTYPE html>`, `<head>`, `<body>`
- [ ] Add `<meta charset>`, `<meta name="viewport" content="width=device-width, initial-scale=1">`
- [ ] Link Google Fonts: Lexend and Space Mono via `<link>`
- [ ] Add structural elements:
  - `#phase-label` — phase name text
  - `#timer-display` — large countdown digits
  - `#action-btn` — primary action button
  - `#waffle-count` — counter display
  - `#waffle-icon` — SVG placeholder

---

## 2. Waffle SVG Icon

- [ ] Draw inline SVG representing a waffle (grid/circle shape)
- [ ] Set initial fill to pale batter color (`#F5E6C8`)
- [ ] Make fill programmable (JS will update it via `setAttribute` or CSS variable)
- [ ] Position icon above phase label

---

## 3. Base CSS

- [ ] Set background color: `#FFF8F0`
- [ ] Set text color: `#3E2723`
- [ ] Apply Lexend to all UI text, Space Mono to `#timer-display`
- [ ] Vertical centering: flex column, `min-height: 100dvh`, `justify-content: center`, `align-items: center`
- [ ] Timer font size: `clamp(4rem, 20vw, 8rem)` or similar
- [ ] Button style: solid rectangle, no border-radius, no border, no shadow. Minimum height 48px, minimum width 160px. Background: `#F59E0B`, text: dark
- [ ] Phase label: uppercase, Lexend, medium size
- [ ] Waffle count: small, bottom of layout, always visible
- [ ] Responsive: works at 320px viewport width

---

## 4. Phase Background Colors

- [ ] Define CSS custom properties or classes for per-phase background tints:
  - IDLE / DONE: `#FFF8F0` (base)
  - SIDE 1: slightly warm tint (e.g. `#FFF3E0`)
  - SIDE 2: slightly deeper warm tint (e.g. `#FFE0B2`)
- [ ] JS applies the correct class/variable when phase changes

---

## 5. State Machine (JS)

- [ ] Define phase constants: `IDLE`, `SIDE1`, `SIDE2`, `DONE`
- [ ] Single `state` variable tracking current phase
- [ ] `setState(phase)` function that:
  - Updates `state`
  - Calls render functions
  - Starts/stops timer as needed
- [ ] Initial state: `IDLE`

---

## 6. Timer Logic

- [ ] `startTimer(durationMs)` — records `startTime = performance.now()`, sets `endTime = startTime + durationMs`
- [ ] `requestAnimationFrame` loop that:
  - Computes `remaining = endTime - performance.now()`
  - Formats as `M:SS` and writes to `#timer-display`
  - If `remaining <= 0`, stops loop and triggers phase transition
- [ ] `stopTimer()` — cancels the animation frame
- [ ] No drift — always diff against wall clock, never accumulate ticks
- [ ] Display `1:00` on entering SIDE 1, `3:00` on entering SIDE 2, `0:00` on DONE

---

## 7. Waffle Counter

- [ ] Initialize `waffleCount = 0`
- [ ] Increment when state transitions to DONE
- [ ] Update `#waffle-count` text on every change
- [ ] Display format: e.g. `Waffles: 3` or `3 waffles`

---

## 8. Button Behavior

- [ ] IDLE: show "Start" button → clicking transitions to SIDE 1
- [ ] SIDE 1 / SIDE 2: hide button (no pause/resume)
- [ ] DONE: show "Next Waffle" button → clicking transitions to SIDE 1

---

## 9. Audio (Web Audio API)

- [ ] Create `AudioContext` lazily on first user gesture (to satisfy mobile autoplay policy)
- [ ] `playSubtleBeep()`:
  - `OscillatorNode` at ~440Hz
  - Duration: ~200ms
  - Gain: low-medium (e.g. 0.3)
- [ ] `playLoudBeep()`:
  - `OscillatorNode` at ~600Hz
  - Duration: ~400ms, played twice with ~150ms gap between
  - Gain: noticeably louder (e.g. 0.7)
- [ ] Call `playSubtleBeep()` on SIDE 1 → SIDE 2 transition
- [ ] Call `playLoudBeep()` on SIDE 2 → DONE transition

---

## 10. Flip Animation (CSS + JS)

- [ ] Add CSS `@keyframes flipTimer`:
  - `0%`: `rotateX(0deg)`
  - `100%`: `rotateX(180deg)`
  - Duration: 600ms, ease-in-out
- [ ] Set `perspective` on parent container
- [ ] Apply `backface-visibility: hidden` to timer element
- [ ] JS: when SIDE 1 ends:
  1. Add `.flipping` class to trigger CSS animation
  2. At 300ms midpoint (90°), swap displayed time to `3:00` and update phase label to "FLIP — SIDE 2"
  3. Listen for `animationend`, remove `.flipping` class
- [ ] Waffle SVG participates in the flip (is inside the animated container)

---

## 11. Waffle Icon Color Progression

- [ ] Define start color: `#F5E6C8` (pale batter), end color: `#C68A2E` (golden brown)
- [ ] Total cook duration: 240,000ms (1min SIDE 1 + 3min SIDE 2)
- [ ] Each animation frame, compute `elapsed` across both phases:
  - During SIDE 1: `elapsed = performance.now() - side1StartTime`
  - During SIDE 2: `elapsed = 60000 + (performance.now() - side2StartTime)`
- [ ] Interpolate RGB values linearly: `color = lerp(startColor, endColor, elapsed / 240000)`
- [ ] Write interpolated color to SVG fill each frame
- [ ] Reset to `#F5E6C8` when entering SIDE 1 (new waffle)

---

## 12. Render Function

- [ ] `render()` called by `setState()` — updates DOM to match current state:
  - Phase label text
  - Button visibility and label
  - Background tint class
  - Timer display (static value when not running)

---

## 13. Polish & Edge Cases

- [ ] Ensure AudioContext is resumed after user gesture (call `audioCtx.resume()` in button click handlers)
- [ ] Verify timer displays correct static value in IDLE (`--:--` or blank) and DONE (`0:00`)
- [ ] Test that rapid tapping "Next Waffle" doesn't corrupt state (cancel any in-progress animation/timer before starting new one)
- [ ] Verify flip animation doesn't show backwards text at midpoint
- [ ] Test on mobile: tap targets, font scaling, no content overflow at 320px

---

## 14. Final File

- [ ] Validate HTML (no unclosed tags, valid structure)
- [ ] Remove any debug `console.log` statements
- [ ] Confirm single `index.html` contains all HTML, CSS, and JS — no external files except Google Fonts CDN
