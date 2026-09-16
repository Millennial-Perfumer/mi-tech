# Inventory

All inventory endpoints are prefixed with `/api`.

## Adjust stock

`POST /api/inventory/adjust?id={item_id}&delta={delta}`

Adjusts stock by a signed integer delta. The following optional query
parameters are recorded in the inventory audit log:

- `reason` — adjustment reason, default `manual_adjustment`, maximum 50 characters.
- `platform` — source platform, default `internal`, maximum 50 characters.
- `external_order_id` — optional order or reference ID, maximum 100 characters.

Example:

```text
POST /api/inventory/adjust?id=66&delta=-1&reason=sale&platform=flipkart&external_order_id=FK-123
```

Existing callers that provide only `id` and `delta` continue to use the
`internal` / `manual_adjustment` defaults.
