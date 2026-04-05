# API Reference: WhatsApp Automation

Manage WhatsApp message templates, automation triggers, and real-time customer conversations.

## 🛠 Base Path: `/api/automation/whatsapp`

### 📋 Template Management
| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/templates` | `GET` | ✅ | List all saved WhatsApp templates. |
| `/templates` | `POST` | 🛡️ Admin | Create/Import a new template. |
| `/templates` | `PUT` | 🛡️ Admin | Update template variable mappings. |
| `/templates` | `DELETE` | 🛡️ Admin | Delete a template. |
| `/templates/sync` | `POST` | 🛡️ Admin | Sync template approval statuses from Meta. |
| `/templates/sync-all`| `POST` | 🛡️ Admin | Full import of all templates from Meta. |
| `/templates/fetch` | `GET` | 🛡️ Admin | Preview a template directly from Meta by name. |
| `/templates/upload`| `POST` | 🛡️ Admin | Upload media (images/PDFs) for template headers. |

### ⚡ Trigger Management
| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/triggers` | `GET` | ✅ | List all automation triggers (Webhook -> Template). |
| `/triggers` | `POST` | 🛡️ Admin | Create a new automation trigger. |
| `/triggers` | `PUT` | 🛡️ Admin | Enable/Disable a trigger. |
| `/triggers` | `DELETE` | 🛡️ Admin | Delete a trigger. |

### 💬 Messaging & Chat
| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/messages` | `GET` | ✅ | List logs of all sent/received messages. |
| `/conversations` | `GET` | ✅ | List active customer chat threads. |
| `/chat` | `GET` | ✅ | Retrieve message history for a specific conversation. |
| `/send-message` | `POST` | 🛡️ Admin | Send a free-text message to a customer. |
| `/send-manual` | `POST` | 🛡️ Admin | Send a template message for a specific order. |
| `/send-bulk` | `POST` | 🛡️ Admin | Send bulk marketing messages to selected customers. |
| `/conversations/mode`| `PUT` | 🛡️ Admin | Switch between `auto` (Bot) and `human` (Agent) mode. |

### 📊 Analytics & Integration
| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/metrics` | `GET` | ✅ | Get automation stats (Delivery/Read rates). |
| `/sync-metrics` | `POST` | 🛡️ Admin | Sync delivery metrics directly from Meta Insight API. |
| `/webhook` | `POST` | 🔐 HMAC | Meta Webhook endpoint for status updates and incoming messages. |

## 📖 Key Endpoint Details

### Create Trigger
`POST /api/automation/whatsapp/triggers`
Links a Shopify webhook event to a specific WhatsApp template.

**Request Body:**
```json
{
  "webhook_topic": "orders/fulfilled",
  "template_id": 42
}
```

### Send Bulk Marketing
`POST /api/automation/whatsapp/send-bulk`
Sends marketing messages to multiple customers. Only templates with category `MARKETING` or a specific suffix (configurable) are allowed.

**Request Body:**
```json
{
  "customer_ids": [101, 102, 105],
  "template_id": 42
}
```

### Conversation Mode
`PUT /api/automation/whatsapp/conversations/mode`
Prevents the bot from auto-responding when an agent is manually chatting with a customer.

**Request Body:**
```json
{
  "id": 5,
  "mode": "human"
}
```

---
> [!IMPORTANT]
> The `/webhook` endpoint requires `hub.verify` for the initial Meta handshake and `X-Hub-Signature-256` for ongoing message validation. Ensure `WHATSAPP_APP_SECRET` is configured.
