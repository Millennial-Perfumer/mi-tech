# Skill Memory (LEARNINGS.md) - Data Strategist

This file contains the technical memory and lessons learned for metrics-sk. Use this as a reference before starting any task to avoid repeating past mistakes.

## Technical Insights
- **2026-04-03 Marketing Console Funnel**: Detailed stage definitions (Lead -> Order) are critical for calculating Conversion Rates at each level.

## Bug Fixes & Red Flags
- **2026-04-03 Marketing Data Missing**: Metrics showed zero for spend/RoAS.
    - **Root Cause**: Missing or broken backend mapping for Shopify `MoneyV2` entities and related Ad Insights.
    - **Resolution**: Always verify numerical mapping logic and use explicit float conversion for Currency fields.

## Architectural Context
- **2026-04-03 Marketing Console**: The dashboard depends on data from Shopify and Meta Ad accounts. Ensure cross-platform data consistency.
