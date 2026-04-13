# Technical Manager - Project Learnings

## Knowledge Graph Integration
- **Context Source**: The codebase is fully indexed in the `graphify-out/` directory.
- **Architectural Analysis**: Before every task, the `GRAPH_REPORT.md` and `.graphify_analysis.json` MUST be consulted to identify:
    - **God Nodes**: High-impact components like `SettingsProvider`, `AutomationHandler`, and `MetaMarketingClient`.
    - **Side Effects**: Inferred semantic relationships (e.g., Security risks in `SECURITY_LOG.md` affecting specific `whatsapp_automation` logic).
- **Navigation**: Use the Obsidian vault at `~/Documents/Personal Dev/GST Invoice Manager/mi-tech-obsidian` for rapid conceptual browsing.

## Graph Update Policy
- **Triggers**: A graph refresh (Partially cached) MUST be triggered after completing:
    - New backend handlers, services, or repository interfaces.
    - Database schema migrations.
    - Major architectural documentation (ADRs).
    - Large-scale refactors.
- **Exclusions**: Trivial UI tweaks, CSS adjustments, or single-line bug fixes do not require an immediate refresh.
- **Efficiency**: Leverage the `.graphify_detect.json` cache to ensure only modified files are re-processed.

## Coordination Patterns
- **Extraction Mode**: When `agent-browser` is unavailable, manual semantic reading of docs/images is required to maintain graph integrity.
- **Doc-as-Code**: Strictly follow the mandate in `GEMINI.md` for all documentation updates.

## Performance & Scalability
- **Parallel Processing**: Use Go goroutines for bulk operations (proven in `c714e1ec-088f-4f14-bd53-551bb4d19ec6`).
- **Dependency Flow**: Always verify persistence interfaces in `backend/internal/repository/` before modifying services.
