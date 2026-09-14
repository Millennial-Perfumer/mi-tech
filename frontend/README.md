# Mi Tech Operations UI

This is the primary operations UI for Mi Tech. It contains the v2 responsive application shell and is served on the main operations hostname.

The previous operations UI is preserved in `frontend-legacy/` for rollback and local comparison. `frontend-feedback` remains a separate customer feedback application and is intentionally unchanged.

## Design foundation

- UI font: `Manrope`
- Monospace font: `Geist Mono`
- Canvas: `#F8F8F6`
- Surface: `#FFFFFF`
- Soft surface: `#F4F4F1`
- Text: `#121212`, `#383838`, `#73736F`, `#A4A49E`
- Borders: `#E6E6E1`, `#CFCFC9`
- Status: `#4C4C48`

The UI uses the existing backend contracts and product capabilities through one responsive application shell.

## Local development

```bash
npm install
npm run dev
```
