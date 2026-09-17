# SMM Queue

The SMM Queue accepts a carousel, caption, and hashtags and stores them in the
private Azure Blob Storage container configured by `SMM_QUEUE_CONTAINER`.

## Queue list and item details

`GET /api/smm-queue`

Returns only timestamped folders that contain `_READY.json`. Incomplete Azure
folders are not shown. Queue items initially have the `queued` status; n8n
status callbacks will be added in a later phase.

`GET /api/smm-queue/{folder}`

Returns the manifest for one queue folder, including the shared caption,
hashtags, and carousel order.

## Create or update a queue item

`POST /api/smm-queue`

`PUT /api/smm-queue/{folder}`

`PUT` creates the replacement in a new timestamped folder first and deletes the
old folder only after the replacement is ready.

This admin-only endpoint accepts `multipart/form-data` with:

- `caption` — required post caption, up to 5,000 characters.
- `hashtags` — optional comma- or space-separated hashtags.
- `media` — one or more image files, in carousel order; up to 10 files and
  10 MB per file.

Example:

```bash
curl -X POST https://mi-tech.example.com/api/smm-queue \
  -H "Authorization: Bearer $TOKEN" \
  -F 'caption=New launch carousel' \
  -F 'hashtags=#millennialperfumer,#newlaunch' \
  -F 'media=@first.jpg' \
  -F 'media=@second.jpg'
```

The API uploads the images, then `manifest.json`, and publishes `_READY.json`
last. The virtual folder is UTC timestamp-prefixed, for example:

```text
20260916T050000Z_<queue-id>/
  01-first.jpg
  02-second.jpg
  manifest.json
  _READY.json
```

n8n should select the latest folder containing `_READY.json`, read the manifest
to preserve media order, and then publish the carousel. Partial uploads do not
receive a ready marker and should be ignored by the workflow.

## Delete a queue item

`DELETE /api/smm-queue/{folder}`

Deletes all blobs belonging to a ready queue folder. The UI asks for
confirmation before invoking this operation.

## Configuration

Configure the SMM Queue from **Settings → Connected services → SMM Queue**.
Secret values are masked and can only be edited after an administrator reveals
them for the current session. The API resolves these settings from the
database on each storage operation, so credential changes do not require an
API restart.

Use one credential option. The connection string is preferred; an optional SAS
token with the account name is also supported:

```text
azure_storage_connection_string   # secret, optional
azure_storage_account_name        # e.g. mptechstg
azure_storage_sas_token            # secret, optional
smm_queue_container                # e.g. mp-smm-queue
```

Environment variables with the equivalent `AZURE_STORAGE_*` and
`SMM_QUEUE_CONTAINER` names remain as a compatibility fallback. An OAuth
bearer access token should not be pasted here; use a scoped SAS token or a
managed identity instead.
