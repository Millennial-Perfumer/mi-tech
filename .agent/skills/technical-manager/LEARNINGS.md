# Skill Memory (LEARNINGS.md)

This file contains the technical memory and lessons learned for technical-manager.

## Technical Insights
- **2026-04-03 Meta Drill-Down**: When building hierarchical Marketing dashboards (Campaign -> Ad Set -> Ad), the backend MUST fetch insights at the child level (level=campaign, adset, or ad) for each specific view. This allows the frontend to populate individual rows and aggregate them for the top total cards.
- **2026-04-03 Multi-Level Mapping**: Use a lookup helper (`getInsightForId`) in the frontend to map a flat insights array to diverse object lists.

## Bug Fixes & Red Flags
- **2026-04-03 404 Route Shadowing**: High-priority or deeply nested routes in `router.go` should be registered at the TOP of the block to prevent being shadowed by generic prefix matches.
- **2026-04-03 Silent OAuth Failures**: Meta API errors (like expired tokens) must be explicitly propagated as 401s with clear "Session Expired" messages. Silently returning `null` leads to confusing "0 metrics" reports.
- **2026-04-03 Port 8080 Bind Failure**: If the backend fails to reflect changes after a rebuild, check for stale processes with `lsof -i :8080` and use `kill -9` if `Air` is blocked.

## Architectural Context
- **2026-04-03 Marketing Client**: Uses a GORM-backed settings repo. Tokens are pulled via `c.settings.GetMetaMarketingAccessToken()`.
- **2026-04-03 System User Tokens**: Prefer these for "permanent" access in Business-level integrations.
