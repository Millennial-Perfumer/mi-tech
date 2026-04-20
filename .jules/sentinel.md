## 2026-04-20 - [Broken Webhook Verification Authentication]
**Vulnerability:** The WhatsApp Webhook verification endpoint (GET) was returning the `hub.challenge` without any token validation, and the Meta Marketing handler was using non-timing-safe string comparison.
**Learning:** Returning a challenge blindly allows unauthorized actors to register themselves as a webhook target for a Meta App, potentially intercepting sensitive data if they can influence the Meta App settings or if they are testing for exposed endpoints. Webhook verification must be treated as an authentication phase.
**Prevention:** Always validate `hub.mode == "subscribe"` and verify `hub.verify_token` using `subtle.ConstantTimeCompare` before responding with the challenge.
