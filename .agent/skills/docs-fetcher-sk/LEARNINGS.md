# Documentation Retrieval Learnings (docs-fetcher-sk)

This file tracks persistent patterns, high-value documentation URLs, and extraction quirks discovered during documentation research.

## Documentation Entry Points

| Service | Documentation Root URL | Extraction Pattern |
| :--- | :--- | :--- |
| **Meta Marketing API** | `https://developers.facebook.com/docs/marketing-api/` | Standard Graph API structure. Focus on 'Insights' and 'Reference' folders. |
| **Shopify API** | `https://shopify.dev/docs/api/admin-graphql` | GraphQL structure. Prioritize 'Objects' and 'Queries'. |

## Extraction Patterns & Quirks

- **Meta (Facebook)**: Often uses complex JavaScript for navigation, but the actual content is available in the initial HTML for `read_url_content`.
- **Shopify**: GraphQL documentation is extremely nested. Prefer searching for specific field names (e.g., `MoneyV2`) directly via `search_web` to find the exact sub-page.

## Versioning Patterns

- **Meta**: URLs typically include the version (e.g., `/v19.0/`). Always check for the latest version in the global settings or use the version documented in the codebase as a primary target.
