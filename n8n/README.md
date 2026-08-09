# n8n Google Drive Automated Social Media Queuer (Photos, Carousels & Videos)

This repository contains templates and workflows to build a **zero-maintenance Google Drive queue**. You simply drop folders containing photos or videos into Google Drive, and n8n automatically detects the media type, posts to social platforms, moves the folder to an archive, and texts you a **WhatsApp completion report**.

---

## 📁 Google Drive Folder Layout

Create a main folder in Google Drive named `automation/`. Each subfolder inside represents 1 queued post:

```
Google Drive: automation/
├── Post_01_Launch/          (1 Image + caption.txt) -> Single Photo Post
├── Post_02_Carousel/        (3 Images + caption.txt) -> Carousel Post
├── Post_03_PromoVideo/      (1 Video .mp4 + caption.txt) -> Reel / Video Post
└── Published_Archive/       (Automated Archive folder)
```

---

## 🎥 Supported Media Types

- **Single Photo**: 1 `.jpg` / `.png` / `.webp`
- **Carousel**: 2 to 10 `.jpg` / `.png` images
- **Video / Reel**: 1 `.mp4` / `.mov` video file

---

## 🚀 How it Works in n8n

1. **Scan Google Drive**: n8n checks `automation/` on your schedule (e.g. 9 AM, 2 PM, 7 PM).
2. **Read Content**: Reads the media files + `caption.txt`.
3. **Publish**: Auto-selects Single Photo, Carousel, or Reel API endpoint for Facebook, Instagram, Threads, and X.
4. **Archive**: Moves the processed subfolder to `automation/Published_Archive/`.
5. **WhatsApp Notification**: Sends a live report to your WhatsApp phone number.

---

## 📱 WhatsApp Report Example

```
📁 Google Drive Post Published!

📂 Folder: Post_03_PromoVideo
🎥 Media Type: Video / Reel (.mp4)
📝 Caption: Check out our product demo video!

✅ Instagram Reel: Published
✅ Facebook Video: Published
✅ Threads Video: Published
✅ X Video Tweet: Published

📁 Folder moved to Published_Archive/
⏰ Published at 09:00 AM
```
