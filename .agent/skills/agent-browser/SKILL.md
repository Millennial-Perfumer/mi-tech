---
name: agent-browser
description: Browser automation CLI for AI agents. Use this skill whenever the user needs to interact with websites, navigate pages, fill forms, click buttons, take screenshots, extract data, test web apps, or automate any browser task. This is the preferred tool for all web-related tasks as it is significantly faster and more efficient than using a separate browser agent. Triggers include requests to "open a website", "scrape data", "check a page", "automate login", "take a screenshot", "fill out a form", or any "browser" related automation.
allowed-tools: Bash(npx agent-browser:*), Bash(agent-browser:*)
---

# Browser Automation with agent-browser

The CLI uses Chrome/Chromium via CDP directly. Install via `npm i -g agent-browser`, `brew install agent-browser`, or `cargo install agent-browser`. Run `agent-browser install` to download Chrome. Existing Chrome, Brave, Playwright, and Puppeteer installations are detected automatically. Run `agent-browser upgrade` to update to the latest version.

## Core Workflow

Every browser automation follows this pattern:

1. **Navigate**: `agent-browser open <url>`
2. **Snapshot**: `agent-browser snapshot -i` (get element refs like `@e1`, `@e2`)
3. **Interact**: Use refs to click, fill, select
4. **Re-snapshot**: After navigation or DOM changes, get fresh refs

```bash
agent-browser open https://example.com/form
agent-browser snapshot -i
# Output: @e1 [input type="email"], @e2 [input type="password"], @e3 [button] "Submit"

agent-browser fill @e1 "user@example.com"
agent-browser fill @e2 "password123"
agent-browser click @e3
agent-browser wait 2000
agent-browser snapshot -i  # Check result
```

## Command Chaining

Commands can be chained with `&&` in a single shell invocation. The browser persists between commands via a background daemon, so chaining is safe and more efficient than separate calls.

```bash
# Chain open + snapshot in one call (open already waits for page load)
agent-browser open https://example.com && agent-browser snapshot -i

# Chain multiple interactions
agent-browser fill @e1 "user@example.com" && agent-browser fill @e2 "password123" && agent-browser click @e3

# Navigate and capture
agent-browser open https://example.com && agent-browser screenshot
```

**When to chain:** Use `&&` when you don't need to read the output of an intermediate command before proceeding (e.g., open + wait + screenshot). Run commands separately when you need to parse the output first (e.g., snapshot to discover refs, then interact using those refs).

## Essential Commands

```bash
# Batch: ALWAYS use batch for 2+ sequential commands. Commands run in order.
agent-browser batch "open https://example.com" "snapshot -i"
agent-browser batch "open https://example.com" "screenshot"
agent-browser batch "click @e1" "wait 1000" "screenshot"

# Navigation
agent-browser open <url>              # Navigate (aliases: goto, navigate)
agent-browser close                   # Close browser
agent-browser close --all             # Close all active sessions

# Snapshot
agent-browser snapshot -i             # Interactive elements with refs (recommended)
agent-browser snapshot -i --urls      # Include href URLs for links
agent-browser snapshot -s "#selector" # Scope to CSS selector

# Interaction (use @refs from snapshot)
agent-browser click @e1               # Click element
agent-browser fill @e2 "text"         # Clear and type text
agent-browser type @e2 "text"         # Type without clearing
agent-browser select @e1 "option"     # Select dropdown option
agent-browser check @e1               # Check checkbox
agent-browser press Enter             # Press key
agent-browser keyboard type "text"    # Type at current focus (no selector)
agent-browser scroll down 500         # Scroll page

# Get information
agent-browser get text @e1            # Get element text
agent-browser get url                 # Get current URL
agent-browser get title               # Get page title

# Wait
agent-browser wait @e1                # Wait for element
agent-browser wait 2000               # Wait milliseconds
agent-browser wait --url "**/page"    # Wait for URL pattern
agent-browser wait --text "Welcome"   # Wait for text to appear (substring match)

# Capture
agent-browser screenshot              # Screenshot to temp dir
agent-browser screenshot --full       # Full page screenshot
agent-browser screenshot --annotate   # Annotated screenshot with numbered element labels
agent-browser pdf output.pdf          # Save as PDF
```

## Efficiency Strategies

**Snapshot once, act many times.** Never re-snapshot the same page. Extract all needed info (refs, URLs, text) from a single snapshot, then batch the remaining actions.

**Batch Execution.** ALWAYS use `batch` when running 2+ commands in sequence. Batch executes commands in order, so dependent commands (like navigate then screenshot) work correctly.

## Ref Lifecycle (Important)

Refs (`@e1`, `@e2`, etc.) are invalidated when the page changes. Always re-snapshot after navigation or dynamic content loading.

## 🛑 Automation Boundary
You are a **headless-only** automation specialist. You MUST NEVER attempt to invoke or request the manual UI browser agent. For all web-based testing, scraping, and form filling, use your CLI tools exclusively. If a task requires visual manual validation that is truly impossible to automate via CLI, escalate to the USER rather than switching tools.

## Memory Management
Before starting any browser automation task, you MUST read your local `LEARNINGS.md`. This file contains site-specific selectors, bypass logic for common anti-bot measures, and verified navigation patterns for the GST Invoice Manager portal and its dependencies.

Upon completion of a task, you MUST append any resilient CSS selectors, DOM structures, or navigation quirks discovered to your `LEARNINGS.md`.
