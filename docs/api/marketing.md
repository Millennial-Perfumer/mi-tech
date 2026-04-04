# API Reference: Meta Marketing

Integration with Meta Ads API for ROI tracking and campaign management.

## 🛠 Base Path: `/api/marketing/meta`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/overview` | `GET` | ✅ | High-level summary of ad accounts and performance. |
| `/campaigns`| `GET` | ✅ | List all campaigns for a specific ad account. |
| `/adsets` | `GET` | ✅ | List ad sets (with insights) for a campaign. |
| `/ads` | `GET` | ✅ | List individual ads (with insights) for an ad set. |
| `/webhook` | `POST` | 🔐 HMAC | Meta Webhook for lead-gen and ad status changes. |

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

---
> [!WARNING]
> If the Meta API token expires, this API will respond with `401 Unauthorized` and a message: `"Meta Session Expired. Please update your API token in Settings."`.
