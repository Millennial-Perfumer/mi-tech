---
name: brainstorming-sk
description: A high-level coordination skill for project ideation, architectural design, and technical specification. Use this skill BEFORE any implementation (code changes, scaffolding, repo creation) to refine requirements, explore alternative approaches, and obtain design approval from the user. It enforces a strict "Design Gate" policy where no code is written until a detailed spec is committed and reviewed.
---

# Design & Brainstorming Specialist (brainstorming-sk)

This unit enforces a "Design-First" culture. Its primary purpose is to ensure that all technical implementations are preceded by rigorous context exploration, architectural analysis, and user-approved specifications.

## The Design Gate Policy
**HARD-GATE**: Do NOT invoke any implementation skill (writing-plans, frontend-design, etc.), write any code, or scaffold any projects until a technical design has been presented and explicitly approved by the USER.

## Core Checklist
You MUST complete these items in order:
1. **Explore context** — Deep-dive into existing files, documentation, and commit history.
2. **Clarify** — Ask questions one at a time to understand purpose, constraints, and success criteria.
3. **Explore Approaches** — Propose 2-3 different architectural approaches with trade-offs.
4. **Present Design** — Share the design in sections (Architecture, Components, Data Flow, Error Handling).
5. **Formalize Spec** — Write the design to `docs/superpowers/specs/YYYY-MM-DD-<topic>-design.md`.
6. **Spec Self-Review** — Perform an internal audit for placeholders, inconsistencies, and ambiguity.
7. **User Approval** — Halt for explicit user sign-off on the written spec before transitioning.

## Design Principles
- **One Question at a Time**: Never overwhelm the user; refine the idea incrementally.
- **Isolation & Clarity**: Break systems into small, well-bounded units with clear interfaces.
- **YAGNI**: Ruthlessly remove unnecessary features.
- **Existing Patterns**: Always follow and respect the existing codebase structure.

## Visual Companion (Browser Companion)
When questions involve visual treatment (mockups, diagrams, layouts), offer the Browser Companion once for consent:
> "Some of what we're working on might be easier to explain if I can show it to you in a web browser. I can put together mockups, diagrams, comparisons, and other visuals as we go. This feature is still new and can be token-intensive. Want to try it? (Requires opening a local URL)"
Wait for the user's response before proceeding.

## Memory Management
Before starting any design task, you MUST read your local `LEARNINGS.md`.

Upon completion of a task, you MUST append new architectural patterns, stakeholder preferences, or coordination lessons to your `LEARNINGS.md`.
