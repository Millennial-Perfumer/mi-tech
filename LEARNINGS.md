# Technical Manager - Project Learnings

## Knowledge Graph Integration
- **Context Source**: The codebase is fully indexed in the `graphify-out/` directory.
- **Architectural Analysis**: Before every task, the `GRAPH_REPORT.md` and `.graphify_analysis.json` MUST be consulted to identify:
    - **God Nodes**: High-impact components like `SettingsProvider`, `AutomationHandler`, and `MetaMarketingClient`.
    - **Side Effects**: Inferred semantic relationships (e.g., Security risks in `SECURITY_LOG.md` affecting specific `whatsapp_automation` logic).
- **Navigation**: Use the Obsidian vault at `~/Documents/Personal Dev/GST Invoice Manager/mi-tech-obsidian` for rapid conceptual browsing.

## Graph Update Policy
- After any significant feature addition or architectural refactor, run `graphify` (AST mode is sufficient for structural changes) to keep the knowledge base fresh.
- Technical Manager MUST review the `GRAPH_REPORT.md` to identify new "God Nodes" or unexpected coupling.

## Data Integrity & Stickiness Policy
- **Local Metadata Supremacy**: Metadata generated or stamped locally (e.g., `delivered_at`, `feedback_status_id`) must be protected from being overwritten by `nil` values during synchronization with external sources (Shopify, Meta).
- **Sticky Upsert Pattern**: Use `Selective Stamping` and preservation logic in repository `Upsert` methods to ensure existing timestamps are not wiped by partial updates.
- **Status Transition Logic**: All methods that update order status are responsible for stamping relevant lifecycle timestamps to maintain audit consistency.

## Coordination Patterns
- **Extraction Mode**: When `agent-browser` is unavailable, manual semantic reading of docs/images is required to maintain graph integrity.
- **Doc-as-Code**: Strictly follow the mandate in `GEMINI.md` for all documentation updates.

## Performance & Scalability
- **Parallel Processing**: Use Go goroutines for bulk operations (proven in `c714e1ec-088f-4f14-bd53-551bb4d19ec6`).
- **Dependency Flow**: Always verify persistence interfaces in `backend/internal/repository/` before modifying services.
