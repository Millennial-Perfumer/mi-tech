# MI-Tech API Connector Memory

## Existing Integrations
| Platform | Client Location | Auth Method | Config Source |
|----------|----------------|-------------|---------------|
| **Shopify** | `internal/client/shopify/` | Access Token header | `SettingsProvider` (DB) |
| **Amazon SP-API** | `internal/client/amazon/` | STS AssumeRole + SigV4 | `SettingsProvider` (DB → ENV fallback) |
| **WhatsApp** | `internal/automation/whatsapp/` | Meta System User Token | `SettingsProvider` (DB) |
| **Meta Marketing** | `internal/marketing/` | Meta System User Token | `SettingsProvider` (DB) |

## Integration Pattern
All clients follow the same architecture:
1. Accept `*config.SettingsProvider` in constructor (NOT raw `*config.Config`)
2. Retrieve credentials dynamically at call time from `SettingsProvider.Get*()`
3. SettingsProvider checks DB first (`app_configs` table), falls back to `os.Getenv()`
4. HTTP client with 10-30s timeout
5. JSON request/response marshaling

## Adding a New Integration
1. Create `internal/client/<name>/client.go`
2. Add config keys to `app_configs` via new migration (use category name, e.g., `pinnacle`)
3. Add `Get<Name>*()` methods to `SettingsProvider` with ENV fallback
4. Add category to `SettingsTab.tsx` CATEGORY_META + categoryOrder
5. Wire in `server.go` → service → handler
6. Update Swagger annotations

## Config Storage
- Keys stored in `app_configs` table with `category`, `is_secret`, `label`, `sort_order`
- Secret values masked in frontend with toggle-reveal
- Frontend `SettingsTab.tsx` groups configs by category with custom icons/colors
