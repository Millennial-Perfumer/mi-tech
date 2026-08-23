# Top / Heart / Base Notes Workflow

The raw Shopify export stores each product's description in
`custom.product_description` (Lexical rich-text). It comes in two formats:

1. **Structured** (~10 products): labeled list items
   `Top Notes: ... / Heart Notes: ... / Base Notes: ...`
2. **Prose** (~71 products): notes embedded in sentences
   (`opens with ...`, `At its heart ...`, `The base settles into ...`)

Parsing prose reliably with regex is brittle, so the extractor does **not**
try. Notes are filled with an AI pass over the description and merged in.

## Workflow

1. **Generate the notes override stub** with a blank template for every SKU
   that lacks notes:

   ```bash
   python3 listings/make_notes_template.py
   ```

2. **Fill the override with AI** — pass the description of each SKU and capture
   the three note lines (`top_notes`, `heart_notes`, `base_notes`) into
   `listings/notes_override.json`.

3. **Rebuild the catalog** merging the override on top of the previous catalog
   (so existing curated notes are kept unless the override replaces them):

   ```bash
   python3 listings/shopify_to_catalog.py --write --notes notes_override.json
   ```

   Verify nothing is missing — the script prints a `Missing any note` list and
   should report `none`.

## Override JSON shape

```json
{
  "F013": {
    "top_notes": "Mandarin Orange, Pear, Bergamot",
    "heart_notes": "Sea Notes, Lavender, Cardamom",
    "base_notes": "Tonka Bean, Amberwood, Musk"
  },
  ...
}
```

Notes already present in the previous catalog are kept automatically; the
override only needs to add/repair entries.
