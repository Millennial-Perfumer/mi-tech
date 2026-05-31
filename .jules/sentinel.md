## 2025-05-15 - [Permissive CORS Configuration]
**Vulnerability:** Permissive CORS policy echoing back any `Origin` header and setting `Access-Control-Allow-Credentials: true`.
**Learning:** This "reflect-all" pattern effectively bypasses CORS protections, allowing any third-party domain to make authenticated requests to the API on behalf of a user.
**Prevention:** Implement a strict allowlist for the `Origin` header and only set `Access-Control-Allow-Credentials: true` for trusted origins.

## 2026-03-26 - [Authentication Identity Mismatch in Config Reveal]
**Vulnerability:** `ConfigsHandler.RevealConfigs` verified the admin password against the first user in the `users` table instead of the authenticated user performing the request.
**Learning:** Relying on `First()` for sensitive credential verification without an explicit user identifier allows any authenticated user with access to the endpoint to succeed by providing the password of the "first" user (likely the original admin), bypassing individual accountability and potentially allowing horizontal or vertical privilege escalation if the "first" user has different permissions.
**Prevention:** Always extract and use the authenticated user's identity from the request context to fetch their specific credentials for verification.

## 2026-03-28 - [SQL Injection in Dynamic Order Clause]
**Vulnerability:** `CustomerRepository.List` passed the `sortBy` parameter from user input directly to GORM's `Order()` method without validation.
**Learning:** ORM methods like GORM's `Order()` often do not parameterize identifiers (like column names) and instead concatenate them into the raw SQL query. This makes them a direct sink for SQL injection if the input is not strictly validated against an allowlist of permitted columns.
**Prevention:** Always validate dynamic sorting parameters against a hardcoded allowlist map of valid column names before passing them to the database query layer.

## 2026-05-31 - [CORS Wildcard and Credentials Conflict]
**Vulnerability:** `CORSMiddleware` incorrectly set `Access-Control-Allow-Credentials: true` when a wildcard `*` was present in `ALLOWED_ORIGINS`.
**Learning:** Browsers block requests where `Access-Control-Allow-Origin` is `*` and `Access-Control-Allow-Credentials` is `true`. The code attempted to bypass this by echoing the `Origin` header even when the "allowance" came from a wildcard, which reintroduced a vulnerability where any site could make authenticated requests.
**Prevention:** If a wildcard `*` is used, the response must set `Access-Control-Allow-Origin: *` and omit `Access-Control-Allow-Credentials`. Credentials must only be allowed for explicitly listed origins.
