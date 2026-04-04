# API Reference: Meta Marketing

Integration with Meta Ads API for ROI tracking and campaign management.

## 🛠 Base Path: `/api/marketing/meta`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/meta/overview` | `GET` | ✅ | High-level summary of ad accounts and performance. |
| `/smm/overview` | `GET` | ✅ | Social media insights (Instagram/FB) with v22.0 support. |
| `/smm/post/insights` | `GET` | ✅ | Detailed insights for a specific social media post. |
| `/meta/webhook` | `POST` | 🔐 HMAC | Meta Webhook for lead-gen and ad status changes. |

## 📖 Endpoint Details

### Marketing Overview
`GET /api/marketing/meta/overview`

This is the main entry point for the marketing dashboard. It automatically identifies the configured ad account and fetches active campaign insights.

**Query Parameters:**
- `start_date`, `end_date`: YYYY-MM-DD range.

**Sample Response:**
```json
{
  "success": true,
  "accounts": [
    { "id": "act_123456", "name": "Millennial Perfumer Ad Account" }
  ],
  "insights": [
    { "campaign_name": "Perfume Launch 2024", "spend": 5000.0, "impressions": 15000, "clicks": 450 }
  ],
  "summary": [
    { "spend": 12000.5, "impressions": 45000, "clicks": 1200 }
  ],
  "active_id": "act_123456"
}
```

### Lead & Ad Webhooks
`POST /api/marketing/meta/webhook`

Handles incoming real-time notifications from Meta.

**Logic:**
- Specifically configured to process CRM events and Ad account status updates.
- Uses HMAC validation against `META_APP_SECRET`.

### Social Media Overview
`GET /api/marketing/smm/overview`

Retreives real-time follower counts, account reach, and post-level engagement for the configured Instagram Professional account.

**Parameters:**
- `platform`: `instagram` (default)
- `start_date`, `end_date`: YYYY-MM-DD range for the periodic trends.

**Modern Architecture (v22.0):**
- **Unified Views**: Replaces legacy `impressions`.
- **Latency Handling**: Unique reach breakdowns (`follow_type`) take **24-48 hours** to compute by Meta.
- **Carousel Limitation**: Follower breakdowns are categorically unavailable for Carousel posts via the API.

---
> [!IMPORTANT]
> **Account Thresholds**: Follower-level breakdowns and specific demographic insights require the Instagram Professional account to have **100+ followers**.

> [!WARNING]
> If the Meta API token expires, this API will respond with `401 Unauthorized` and a message: `"Meta Session Expired. Please update your API token in Settings."`.
