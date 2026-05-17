# ADR-0001: Autonomous Discovery-First AI Analysis

**Date**: 2026-05-16
**Status**: accepted
**Deciders**: Antigravity, USER

## Context

The Millennial Perfumer chatbot needs to perform complex, ad-hoc data analysis (e.g., revenue trends, customer segmentation, inventory mapping) on a production PostgreSQL database. 

The primary challenges were:
1. **Schema Evolution**: Hardcoding every table and column name into the AI prompt is unscalable and brittle if the schema changes.
2. **Security**: Allowing the AI to generate and execute SQL carries high risks of data mutation or unauthorized access.
3. **Accuracy**: The LLM often "hallucinates" column names (e.g., assuming `user_id` exists instead of `customer_phone`) if it doesn't have a way to verify the schema.

## Decision

We have implemented a **Discovery-First AI Analysis** architecture consisting of three layers:

1. **Discovery Layer**: Tools provided to the AI (`list_database_tables`, `describe_database_table`) that query the `information_schema`. This gives the AI "eyes" to verify the exact spelling and type of columns before writing queries.
2. **Execution Layer**: A strictly read-only `execute_sql_query` tool that acts as a bridge to the database.
3. **Safety Layer (`QueryGuard`)**: A middleware repository that intercepts all SQL strings. It enforces a "SELECT-only" rule and uses regex to block mutation keywords (INSERT, UPDATE, DELETE, DROP, etc.).

## Alternatives Considered

### Alternative 1: Hardcoded Aggregate Methods Only
- **Pros**: 100% safe, high performance.
- **Cons**: Every new business question requires a developer to write a new Go method and handler. Too rigid for a dynamic BI assistant.
- **Why not**: Failed the requirement for "autonomous" and "dynamic" intelligence.

### Alternative 2: Direct SQL without Discovery
- **Pros**: Fastest implementation.
- **Cons**: Frequent errors when the AI guesses column names wrong. High security risk without a guard layer.
- **Why not**: Resulted in poor user experience and dangerous security posture.

## Consequences

### Positive
- **Self-Healing**: The AI can now fix its own "column does not exist" errors by inspecting the table structure.
- **Flexibility**: Can perform complex ad-hoc joins (e.g., linking Amazon SKUs to internal inventory) without backend code changes.
- **Security**: The `QueryGuard` provides a hard defense against accidental or malicious data mutation.

### Negative
- **Latency**: Discovery steps (listing tables) add 1-2 extra LLM turns to the initial request.
- **Complexity**: Requires careful prompt engineering to ensure the AI knows *when* to use discovery tools vs. hardcoded abstractions.

### Risks
- **Query Complexity**: Very complex `SELECT` queries might be incorrectly flagged by `QueryGuard` regex (though word boundaries mitigate this).
- **Mitigation**: The AI has "fallback" hardcoded aggregate tools for common metrics like `get_revenue_summary`.
