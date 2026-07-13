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
## 2026-06-12 - [Missing HMAC Signature Verification in Webhooks]
**Vulnerability:** The `MarketingWebhookHandler.handleNotification` endpoint did not verify the `X-Hub-Signature-256` header, allowing unauthenticated attackers to spoof requests.
**Learning:** For endpoints handling payloads from external systems, validation must happen implicitly via signature verification before any business logic executes. Missing these checks leaves the endpoints wide open.
**Prevention:** Implement HMAC verification for all external webhooks. Specifically, utilize `io.LimitReader` (for DoS protection), restore the body via `io.NopCloser(bytes.NewBuffer(body))`, and enforce strict string comparison logic using `hmac.Equal` to defend against timing attacks.

## 2026-06-13 - [Open Redirect in Tracking System]
**Vulnerability:** The tracking link redirect handler blindly redirected users to whatever `trackingURL` was stored in the database without any domain validation.
**Learning:** Even internal data sources (like a database) can contain unvalidated or poisoned URLs. Blindly redirecting to a stored URL acts as an Open Redirect. However, implementing a strict single-domain allowlist can break legitimate multi-provider integrations (e.g. tracking links for various couriers like Delhivery, Shiprocket).
**Prevention:** Always parse the destination URL and validate its hostname against a broad, yet strict allowlist of trusted root domains (including their subdomains) corresponding to the business's actual integrated partners to prevent attackers from redirecting to arbitrary malicious domains.


## 2024-06-14 - [DoS & Fail-Open Vulnerabilities in Webhooks]
**Vulnerability:**
1. Missing request body size limits in `ShopifyWebhookHandler` and `WhatsAppWebhook` (DoS vulnerability via large payloads).
2. Fail-open behavior in `verifyWebhook` and `validateWhatsAppSignature` where missing secrets bypassed validation checks and returned true/valid.
**Learning:** Handlers processing external webhooks MUST employ `io.LimitReader` before invoking `io.ReadAll` to constrain unbounded inputs, and MUST explicitly default to a "fail-closed" state whenever required credentials or configurations are missing, to ensure security bounds are not inadvertently dropped.
**Prevention:** Always bound external request reads using `io.LimitReader` (e.g., 1MB) and ensure cryptographic signature verifications explicitly return `false` or trigger an error path if the corresponding secret is unconfigured or empty.

## 2026-06-15 - [AI Query Guard Stacked Query & Command Injection Bypass]
**Vulnerability:** The AI `execute_sql_query` tool in `QueryGuard` relied on a simplistic regex to block mutation keywords (like `UPDATE`, `DROP`) but missed advanced Postgres execution constructs (`COPY`, `DO`, `CALL`, `EXECUTE`). Furthermore, it was vulnerable to stacked queries (multiple queries separated by `;`), allowing attackers or hallucinatory AI models to prepend a safe `SELECT` query to bypass the prefix check and append arbitrary commands (e.g., `SELECT 1; COPY users TO PROGRAM 'rm -rf /'`).
**Learning:** Regex-based query guards are inherently brittle and must explicitly account for all database-specific execution, procedural, and file-IO keywords. Additionally, blocking stacked queries entirely (by rejecting `;` outside string literals) is crucial to prevent prefix-bypass attacks on endpoints expecting single statements.
**Prevention:** Always explicitly block procedural and system-level keywords (`COPY`, `DO`, `CALL`, `EXECUTE`) in regex guards. Implement strict stacked-query prevention by parsing the SQL to reject `;` tokens outside of string literals, ensuring only single, verified statements execute.

## 2026-06-18 - [Timing Attack in Webhook Signature Verification]
**Vulnerability:** `ShopifyWebhookHandler` used a simple string equality check (`==`) to compare the provided HMAC signature with the expected HMAC signature in `verifyWebhook`.
**Learning:** Standard string equality checks short-circuit on the first mismatched character. An attacker could measure the time taken for the comparison to fail and incrementally deduce the correct signature character by character, bypassing the verification entirely.
**Prevention:** Always use constant-time comparison functions, such as `hmac.Equal`, when validating cryptographic signatures or sensitive tokens to prevent timing attacks.

## 2025-02-27 - B2B Handlers DoS Vulnerability via Unbounded Request Bodies
**Vulnerability:** Several B2B handlers (`HandleCustomers`, `HandleInvoices`, `HandlePaymentTerms`, etc.) read the HTTP request body using `io.ReadAll(r.Body)` or `json.NewDecoder(r.Body)` without first restricting the maximum payload size, presenting a critical unmitigated Denial-of-Service (DoS) memory exhaustion risk.
**Learning:** Even internal or authenticated routes can be vulnerable to resource exhaustion. The Go standard library does not automatically limit request bodies, so manual constraints must be applied before decoding.
**Prevention:** Always bound request body sizes using `r.Body = http.MaxBytesReader(w, r.Body, 1048576)` (1MB limit) at the start of HTTP POST/PUT handlers, before calling `io.ReadAll` or decoding JSON.

## 2024-06-25 - [DoS Vulnerability in Authentication Handlers via Unbounded Request Bodies]
**Vulnerability:** The authentication and user handlers (`Login`, `VerifyOTP`, `CreateUser`) read the HTTP request body using `json.NewDecoder(r.Body)` without first restricting the maximum payload size, presenting a critical unmitigated Denial-of-Service (DoS) memory exhaustion risk.
**Learning:** Handlers processing unauthenticated external requests MUST employ request body limits before reading or decoding. The Go standard library does not automatically limit request bodies, making this an explicit responsibility of the developer.
**Prevention:** Always bound request body sizes using `r.Body = http.MaxBytesReader(w, r.Body, 1048576)` (1MB limit) at the start of HTTP POST/PUT handlers before calling `json.NewDecoder`.

## 2026-07-13 - [Missing Timeout in External HTTP Clients]
**Vulnerability:** External HTTP requests made using `http.DefaultClient` or unconfigured `http.Client` instances lack timeouts by default in Go. This allows malicious or slow external servers to hold connections open indefinitely, potentially leading to resource exhaustion (Denial of Service) via file descriptor or memory depletion.
**Learning:** In Go, the `http.Get`, `http.Post`, and `http.DefaultClient` instances intentionally do not specify request timeouts. Any integration communicating with external APIs (like Meta or Amazon) must use explicit timeouts to prevent the application from hanging.
**Prevention:** Always instantiate custom `http.Client` structures and define an explicit `Timeout` property (e.g., `Timeout: 30 * time.Second`) before making network requests to third-party endpoints.
