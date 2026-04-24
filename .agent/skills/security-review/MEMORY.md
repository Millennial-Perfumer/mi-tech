# MI-Tech Security Review Memory

## Authentication
- **Method**: JWT Bearer tokens via `Authorization: Bearer <token>` header
- **Middleware**: `internal/handler/auth_middleware.go`
- **Roles**: `admin` (full access), `read` (view-only)
- **JWT Secret**: From `SettingsProvider.GetJWTSecret()` (ENV fallback)
- **OTP Login**: WhatsApp-based OTP delivery for authentication

## Secrets Management
- **Database-backed**: All integration credentials stored in `app_configs` table with `is_secret = true`
- **ENV fallback**: `SettingsProvider` checks DB first, then `os.Getenv()`
- **Never committed**: `.env` excluded via `.gitignore`
- **Frontend masking**: Secret fields masked with toggle-reveal in `SettingsTab.tsx`

## Webhook Security
- **Shopify**: HMAC-SHA256 signature verification on incoming webhooks
- **WhatsApp**: Verify token for webhook subscription validation
- **Meta Marketing**: Webhook verify token

## Key Surface Areas
- `/api/webhooks/shopify` — Public endpoint, must verify HMAC
- `/api/automation/webhook` — WhatsApp webhook, verify token
- `/api/configs` — Admin-only (sensitive credentials)
- All other `/api/*` endpoints require valid JWT

## Known Patterns
- Admin-gated features: `userRole === 'admin'` in frontend, role check middleware in backend
- Rate limiting: Not implemented yet
- CORS: Handled by middleware
- SQL injection: Prevented by GORM parameterized queries
