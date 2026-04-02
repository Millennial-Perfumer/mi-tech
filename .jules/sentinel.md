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

## 2026-04-02 - [Insecure Webhook Verification Patterns]
**Vulnerability:** Webhook handlers for Shopify and WhatsApp lacked essential security controls: missing 1MB body size limits (DoS risk), "security theater" echo-back verification (WhatsApp), and non-constant-time HMAC comparison (Shopify).
**Learning:** Webhook endpoints are public and often highly privileged, as they process data from external providers into the core database. Without strict body size limits and "fail-closed" verification logic (rejecting requests if secrets are not configured), these endpoints are vulnerable to DoS attacks and unauthorized data injection.
**Prevention:** Always implement `http.MaxBytesReader` for public webhook endpoints, use `hmac.Equal` for signature validation, and strictly verify all challenge tokens (like WhatsApp's `hub.verify_token`) against configured secrets before responding.
