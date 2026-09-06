# Mi Tech Frontend v2

This is the replacement frontend workspace for Mi Tech.

The v2 rebuild will preserve the existing backend contracts and product capabilities while moving to one responsive application shell. `frontend-feedback` remains a separate customer feedback application and is intentionally unchanged.

## Design foundation

- UI font: `Manrope`
- Monospace font: `Geist Mono`
- Canvas: `#F8F8F6`
- Surface: `#FFFFFF`
- Soft surface: `#F4F4F1`
- Text: `#121212`, `#383838`, `#73736F`, `#A4A49E`
- Borders: `#E6E6E1`, `#CFCFC9`
- Status: `#4C4C48`

The current shell is deliberately small and data-free. Feature modules will be migrated in vertical slices after the shared layout, tokens, navigation, and accessibility primitives are validated.

## Local development

```bash
npm install
npm run dev
```
