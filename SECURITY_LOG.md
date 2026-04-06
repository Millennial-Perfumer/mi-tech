# Project Security Ledger (SECURITY_LOG.md)

This file acts as the persistent memory for the `security-sk` analyst. It tracks identified vulnerabilities, potential risks, and remediations across the project's development lifecycle.

## 🕒 Ongoing Context & Potential Risks
*These are areas that the analyst has flagged but have not yet been fully audited or remediated.*

- **Shared Verify Tokens**: The Meta Marketing and WhatsApp webhooks share a verification token. This has been flagged as **HIGH** severity in current architectural reviews. [Decoupling plan approved].
- **Public Metrics Endpoint**: The `/api/metrics` endpoint is currently unprotected, potentially leaking system metadata. [Flagged during iteration 1].
- **Shared Verify Tokens**: The Meta Marketing and WhatsApp webhooks share a verification token. This has been flagged as **HIGH** severity in current architectural reviews. [Decoupling plan approved].

## ✅ Remediations & Audited Areas
*Resolved issues and verified secure components.*

- **Webhook Integrity & DoS Protection**: Implemented HMAC/Signature verification for Shopify, WhatsApp, and Meta Marketing webhooks. Added 1MB payload limits and constant-time comparisons. [Verified 2026-05-20].
- **Customer Repository Sorting**: Verified safe use of allowlists for dynamic sorting in `List` method. [Verified 2026-04-03].
- **RBAC Coverage**: Standard endpoints in `router.go` are correctly wrapped in `protected` and `adminProtected` middleware.

## 🛡️ Security Posture Summary
- **Primary Auth**: JWT based via middleware.
- **Role Isolation**: Admin roles required for destructive operations.
- **Input Handling**: GORM parameterized queries used for primary filtering.

Skills Involved: `security-sk`, `sa-skill`
