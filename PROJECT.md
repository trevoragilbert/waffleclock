# waffleflip

A single-purpose waffle cooking timer. Vanilla HTML/JS/CSS, zero dependencies.

## Concept

Eliminates the start-stop-clear annoyance of timing waffles. One tap to start, automatic phase transitions, audio cues at each transition.

## Timer Flow

```
[Start] → SIDE 1 (1:00 countdown) → beep (subtle) → SIDE 2 (3:00 countdown) → beep (louder) → DONE state → [Next Waffle] → loops back to SIDE 1
```

### States

1. **IDLE** — App just opened. Shows "Start" button. Waffle count: 0.
2. **SIDE 1** — 1:00 countdown. Label: "SIDE 1". Timer ticking.
3. **SIDE 2** — 3:00 countdown. Label: "FLIP — SIDE 2". Timer ticking. Auto-transitions from SIDE 1 with a subtle beep.
4. **DONE** — Timer shows 0:00. Louder beep. Label: "DONE". Shows "Next Waffle" button.

### Transitions

| From | To | Trigger |
|------|----|---------|
| IDLE | SIDE 1 | User taps "Start" |
| SIDE 1 | SIDE 2 | Timer hits 0:00 |
| SIDE 2 | DONE | Timer hits 0:00 |
| DONE | SIDE 1 | User taps "Next Waffle" |

### Waffle Counter

Visible at all times. Increments by 1 each time the timer enters DONE. Starts at 0.

## Audio

Generated programmatically via Web Audio API (no audio files).

- **Subtle beep** (SIDE 1 → SIDE 2): Short, single tone. Low-medium volume.
- **Louder beep** (SIDE 2 → DONE): Two-tone or repeated beep. Noticeably louder/longer.

### Waffle Icon

An inline SVG waffle icon (simple grid/circle shape representing a waffle) displayed above the phase label.

- **Color progression**: Starts as a pale batter color (`#F5E6C8` or similar) at the beginning of SIDE 1. Gradually darkens to a golden-brown (`#C68A2E` or similar) by the end of SIDE 2. The color interpolation spans the full cook time (1min + 3min = 4min total).
- Implemented by calculating total elapsed cook time as a 0→1 ratio and interpolating the fill color in JS.
- Resets to pale batter color on "Next Waffle".
- The icon also participates in the flip animation — it flips along with the timer.

### Flip Animation

When SIDE 1 ends and transitions to SIDE 2, the timer display does a 3D flip around the horizontal axis (like a waffle being flipped on a griddle).

- CSS `transform: rotateX()` with `perspective` on the parent.
- Animation: 0 → 180° rotation over ~600ms, ease-in-out.
- At the 90° midpoint (timer edge-on and not readable), swap the displayed time from 0:00 to 3:00 and update the phase label.
- Second half of rotation (90° → 180°) reveals the new timer value.
- The timer element needs `backface-visibility: hidden` or the text swap timed to the midpoint so there's no visible backwards text.
- Pure CSS animation triggered by adding/removing a class. JS listens for `animationend` to clean up.

## UI

### Layout

Single screen, vertically centered. Mobile-first. No scrolling needed.

Top to bottom:
1. Phase label (e.g. "SIDE 1", "FLIP — SIDE 2", "DONE")
2. Timer display (large, dominant)
3. Primary action button ("Start" / "Next Waffle")
4. Waffle count (small, bottom)

### Design

- **Colors**: Warm palette. Creamy/off-white background (`#FFF8F0` or similar). Dark brown text (`#3E2723`). Accent color for active states — warm amber/golden (`#F59E0B` range).
- **Typography**: Lexend (Google Fonts) for all UI text. Space Mono (Google Fonts) for timer digits. Both loaded via `<link>` from Google Fonts CDN.
- **Style**: Stark, flat. No borders, no shadows, no rounded corners on containers. Buttons are simple solid rectangles with no border-radius. Minimal visual noise.
- **Phase indication**: Background color shifts subtly per phase. SIDE 1 gets a slightly warm tint, SIDE 2 gets a slightly deeper warm tint, DONE returns to base.

### Responsive

- Works on mobile screens 320px+.
- Timer text scales with viewport (`clamp()` or similar).
- Tap targets minimum 48px.
- No landscape-specific layout needed, but shouldn't break.

## Technical

### Stack

- Single `index.html` file containing all HTML, CSS, and JS.
- No build step, no dependencies, no frameworks.

### Timer

- Use `setInterval` at ~100ms or `requestAnimationFrame` with timestamp delta for smooth countdown.
- Display in `M:SS` format.
- Timer should not drift — compare against `Date.now()` or `performance.now()`, don't accumulate interval ticks.

### Audio

- Web Audio API: create `OscillatorNode` for beeps.
- Handle AudioContext resume on user gesture (mobile browsers require this).
- Subtle beep: ~440Hz, ~200ms duration.
- Loud beep: ~600Hz, ~400ms, repeated twice with short gap.

### State management

Plain JS. A single state variable tracks current phase. No over-engineering.

## Files

```
index.html   — everything
```

## Out of Scope

- Settings/customization for times
- Pause/resume
- History or persistence
- PWA/service worker
- Multiple concurrent timers