# 🎨 Telegram Perfume AI Poster Generator & Approval Workflow Guide

This workflow transforms any perfume photography screenshot/inspiration sent via Telegram into a high-end luxury poster featuring **your brand's perfume bottle**, complete with inline Telegram approval and feedback loops.

---

## 🏗️ Architecture & Interaction Flow

```
1. 📸 Send Inspiration Photo in Telegram
      ↓
2. 🤖 GPT-4o Vision Deconstructs Style (Lighting, Props, Shadows, Framing)
      ↓
3. 🧪 Injects your specific perfume bottle description / brand DNA
      ↓
4. 🎨 Flux.1 / Replicate generates commercial 8k luxury poster
      ↓
5. 📱 Telegram Bot sends generated poster with buttons:
      [ ✅ Approve & Queue Post ]   [ 🔄 Re-roll ]
      [ ✏️ Adjust Prompt / Feedback ]
      ↓
6. 📁 When Approved: Queues directly into your Social Auto-Publisher workflow!
```

---

## 🚀 Setup & Installation (3 Simple Steps)

### Step 1: Import Workflow into n8n
1. Open your **n8n Canvas**.
2. Click **Workflow Menu (top right) -> Import from File**.
3. Select [`n8n/perfume_ai_poster_workflow.json`](file:///Users/siddiqs_office/Documents/Business/mi-tech/mi-tech/n8n/perfume_ai_poster_workflow.json).

---

### Step 2: Configure Telegram Bot Token
1. Open Telegram and search for `@BotFather`.
2. Send `/newbot` to create your studio bot (e.g. `PerfumeStudioBot`).
3. Copy the HTTP API token provided by BotFather.
4. In n8n, open the **Telegram Trigger** node and create/paste your **Telegram API Credential**.

---

### Step 3: Configure Settings & Product Description
Double-click the node named **`⚙️ Workflow Configurations & Brand Settings`** in n8n and customize:

| Parameter | Description | Example |
| :--- | :--- | :--- |
| `OPENAI_API_KEY` | Your OpenAI API key for Vision analysis | `sk-...` |
| `REPLICATE_API_TOKEN` | API Token for Flux.1 image generation | `r8_...` |
| `PRODUCT_NAME` | Name of your perfume line | `Aura Privée Extrait` |
| `PRODUCT_DESCRIPTION` | Exact details of your bottle, cap, liquid, and label | `a luxury minimalist heavy glass rectangular perfume bottle with amber liquid, sleek cylindrical matte gold cap, labeled with embossed black serif typography 'AURA PRIVÉE'` |
| `GDRIVE_APPROVED_POSTS_FOLDER_ID` | Your Google Drive automation queue folder ID | `1djXkok8cuP3efyurTd2nOwoKRo-HpEC3` |

---

## 🎯 How to Use in Telegram

1. Open a chat with your Telegram Bot.
2. Send **any inspirational perfume photo** or screenshot you found online.
3. The bot will automatically:
   - Reply: *"🎨 Inspiration Received! Analyzing lighting, composition, textures..."*
   - Reply: *"✨ Concept Deconstructed! Generating poster..."*
   - Send the finished high-resolution poster with caption and interactive action buttons.
4. Tap **`✅ Approve & Queue Post`** to confirm or **`✏️ Adjust Prompt`** to tweak the environment!
