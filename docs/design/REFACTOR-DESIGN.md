# Product Architecture & Brand Guidelines: Jotform-CLI (Go Edition)

## 1. Project Vision & Core Philosophy
`jotform-cli` is a next-generation command-line interface built in Go. It goes beyond standard text output, bringing the premium, motion-driven, and highly geometric brand identity of Jotform 2026 directly into the terminal environment. 

The CLI must feel:
* **Premium but approachable:** Utilizing truecolor (24-bit) terminal capabilities.
* **Technical but expressive:** Using TUI (Terminal User Interface) paradigms over static text.
* **Motion-driven:** Implementing subtle terminal animations (spinners, staggering list entries, color pulses).

## 2. Technical Stack (Strict Requirements)
The agent must strictly use the following Go ecosystem tools to achieve the required UX and visual fidelity:
* **Core Routing:** `spf13/cobra` (Command structure) & `spf13/viper` (Config/State).
* **TUI Architecture:** `charmbracelet/bubbletea` (Elm-style architecture for interactive panes).
* **Styling & Theming:** `charmbracelet/lipgloss` (Crucial for implementing the strict brand color palette).
* **Components:** `charmbracelet/bubbles` (Lists, text inputs, progress bars, spinners).
* **API Client:** Standard `net/http` with asynchronous execution (Goroutines/Channels) to keep the TUI non-blocking.
* **Image/QR Rendering:** Utilize custom ANSI block rendering or `go-qrcode` for visual outputs.

---

## 3. Brand Identity & CLI Color Mapping

The CLI must strictly adhere to the official 2026 Jotform Color Palette. `lipgloss` variables must be created globally for these exact hex codes.

### Color Palette Implementation (`lipgloss` Tokens)
* **Brand Navy (`#0A1551`):** * *CLI Role:* Base structure, primary text (if using inverted backgrounds), table borders, panel outlines, and the anchor layer for any ASCII/Block logo rendering.
    * *Feel:* Stable, trusted, premium.
* **Brand Orange (`#FF6100`):** * *CLI Role:* Primary CTA, highlighted list items, selected menu options, error states (with high contrast), and the most dominant visual accent in headers.
* **Brand Blue (`#0099FF`):** * *CLI Role:* Technical accents, progress bars (`bubbles/progress`), loading spinners (`bubbles/spinner`), and secondary motion cues.
* **Brand Yellow (`#FFB629`):** * *CLI Role:* Highlighted search terms, warning messages, "glow" effects (e.g., coloring the padding/borders of an active input field).

### Terminal Adaptation Rules
* *Avoid* using Orange for large blocks of text; keep it as a punchy accent.
* *Gradients:* Use `charmbracelet/lipgloss` gradient features (Blue to Orange, or Navy to Blue) strictly for welcome banners and key success states.

---

## 4. Logo & Geometry Instructions

The Jotform logo consists of four geometric SVG paths. When displaying the logo in the terminal (e.g., during `jotform login` or `jotform --version`), the agent should attempt to render or approximate these shapes using colored ANSI half-blocks or Braille characters, respecting the specific colors:

* **Orange Path (`#FF6100`):** Main diagonal.
    * ````svg
<path d="M8.5324 21.8782C7.0425 20.398 7.0425 17.9982 8.5324 16.518L20.7592 4.37085C22.2491 2.89067 24.6647 2.89066 26.1546 4.37085C27.6445 5.85104 27.6445 8.2509 26.1546 9.73109L13.9278 21.8782C12.4379 23.3584 10.0223 23.3584 8.5324 21.8782Z" fill="#FF6100"></path>
```

* **Navy Path (`#0A1551`):** Base/Anchor.
    * ```svg
<path d="M7.47266 28.5098C8.03906 29.0589 7.6388 29.9996 6.83452 29.9996H1.80114C0.808059 29.9996 0 29.2163 0 28.2536V23.3741C0 22.5944 0.970426 22.2064 1.53682 22.7555L7.47266 28.5098Z" fill="#0A1551"></path>
```

* **Yellow Path (`#FFB629`):** Floating highlight.
    * ```svg
<path d="M15.3409 28.8897C13.851 27.4096 13.851 25.0097 15.3409 23.5295L20.718 18.1875C22.2079 16.7073 24.6235 16.7073 26.1134 18.1875C27.6033 19.6677 27.6033 22.0676 26.1134 23.5478L20.7363 28.8897C19.2464 30.3699 16.8308 30.3699 15.3409 28.8897Z" fill="#FFB629"></path>
```

* **Blue Path (`#0099FF`):** Tech/Depth layer.
    * ```svg
<path d="M1.135 15.4667C-0.354897 13.9865 -0.354896 11.5867 1.135 10.1065L10.184 1.11014C11.6739 -0.370046 14.0895 -0.370048 15.5794 1.11014C17.0693 2.59033 17.0693 4.99019 15.5794 6.47038L6.5304 15.4667C5.0405 16.9469 2.6249 16.9469 1.135 15.4667Z" fill="#0099FF"></path>
```

---

## 5. Motion & UX Principles (Terminal Context)

The brand dictates motion over stillness. In `bubbletea`, this translates to:
* **Assembly Behavior:** When listing forms (`jotform forms`), do not dump the list instantly. Use `bubbletea` `Tick` commands to cascade/stagger the list items falling into place (e.g., 20ms delay per row).
* **Idle Behavior:** When waiting for an API response, the loading spinner (`Brand Blue`) should pulse smoothly. Use `lipgloss` to gently interpolate border colors of the active pane (e.g., shifting between Navy and Blue).
* **Interaction:** Navigating lists (Up/Down) must feel instantaneous. The active item should pop with `Brand Orange` background and bold text, leaving a trail effect if possible.

---

## 6. Core Command Specifications

### 6.1. Interactive Login (`jotform login`)
* **UX:** Trigger a secure `bubbles/textinput`. Render an ASCII version of the Jotform logo above it. 
* **Brand Rule:** The input prompt cursor should be `Brand Orange`. Upon success, display a subtle gradient success banner.

### 6.2. TUI Dashboard (`jotform dashboard`)
* **UX:** Split the terminal. 
    * *Left Pane (Forms):* Scrollable list. Active item borders in `Brand Orange`.
    * *Right Pane (Details):* Render stats. Use `Brand Blue` for metric numbers. 
* **Brand Rule:** Apply `Brand Navy` borders to inactive panes to maintain a structured, premium UI.

### 6.3. Real-Time Watcher (`jotform watch <form_id>`)
* **UX:** A long-running process that listens for submissions.
* **Motion Rule:** When a new submission arrives, use a sliding animation or a "flash" effect (momentarily highlighting the new data box in `Brand Yellow` before settling to `Brand Navy`).

### 6.4. Share & QR (`jotform share <form_id>`)
* **UX:** Generate a terminal QR code. 
* **Brand Rule:** Render the QR code blocks in pure White/Black for scannability, but wrap the QR code within a beautifully padded `lipgloss` box utilizing the `Brand Navy` background and a `Brand Orange` header.

---

## 7. Agent Execution Strategy
1.  **Phase 1: Foundation & Styling.** Initialize the `cobra` project. Create a `theme.go` file explicitly defining the 4 brand colors using `lipgloss.Color()`.
2.  **Phase 2: Authentication TUI.** Implement `jotform login` using `bubbletea` and `bubbles/textinput`.
3.  **Phase 3: The Dashboard.** Build the complex split-pane UI. Ensure asynchronous API calls so the UI never freezes (maintaining the "motion over stillness" principle).
4.  **Phase 4: Logo Assembly.** Implement a custom function to parse or draw the SVG paths into terminal blocks for the welcome screen.