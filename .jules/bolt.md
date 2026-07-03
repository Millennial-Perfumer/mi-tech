## 2025-05-15 - [Safe Batch Processing and Data Integrity]
**Learning:** When refactoring sequential operations into batch processing (e.g., to solve N+1 query problems), it is critical to ensure that error handling remains robust and that data aggregation is preserved. Continuing a batch operation after a lookup failure can lead to data loss. Furthermore, batch processing must account for multiple updates to the same entity within a single batch by aggregating or merging data correctly before persisting.
**Action:** Always return an error or handle partial failures explicitly if a critical step in the batch process fails. Ensure that PII and metrics are merged or aggregated correctly when multiple source records affect the same target entity in a batch.

## 2025-05-16 - [Efficient GORM Batch Upsert with Partial Indexes]
**Learning:** To optimize $O(N)$ iterative upserts into a single $O(1)$ batch operation in GORM, `clause.OnConflict` is the standard approach. However, if the database uses a partial unique index (e.g., `WHERE deleted_at IS NULL`), GORM's `OnConflict` must explicitly target this index using `TargetWhere` (in GORM v1.2x) or `IndexConfig` (in newer versions) to avoid "there is no unique or exclusion constraint matching the ON CONFLICT specification" errors.
**Action:** When performing batch upserts on tables with partial indexes, always use `TargetWhere` to match the index's condition in the `ON CONFLICT` clause.

## 2026-03-27 - [Batching Line Item Upserts]
**Learning:** Iterative database operations within repository methods (like looping over line items to perform individual upserts) create significant overhead due to multiple database roundtrips. Even when using transactions, the per-row execution time adds up. GORM's native batch insert (`tx.Create(&slice)`) reduces this to a single O(1) roundtrip.
**Action:** Always prefer batch operations (`Create`, `Save`) with slices instead of loops when handling child entities or bulk datasets in GORM. Ensure slice elements are updated by index (`for i := range slice`) before the batch call.

## 2026-03-27 - [Optimizing Reporting with SQL Aggregations and Window Functions]
**Learning:** Combining multiple metrics into a single SQL query using `FILTER` and `CASE` (conditional aggregation) eliminates redundant database roundtrips and application-side processing. For HSN/line-item reports, replacing global CTE scans with window functions (`SUM(...) OVER (PARTITION BY order_id)`) within date-filtered JOINs ensures the database only processes relevant rows, significantly improving performance as the table grows.
**Action:** Always prefer conditional SQL aggregation over multiple repository calls for dashboard/reporting logic. Use window functions for per-group aggregations within filtered result sets to avoid full table scans.

## 2026-03-28 - [Batch Inventory Synchronization in UpsertBatch]
**Learning:** Sequential inventory synchronization within a batch order upsert creates an N+1 bottleneck. Aggregating SKU deltas across the entire batch allows for fetching mappings in a single tuple `IN` query (`WHERE (platform, sku) IN ?`) and consolidating stock updates by `InventoryItemID`. Batching the final status flags (e.g., `inventory_deducted`) further reduces overhead.
**Action:** When implementing batch operations that involve related entities or secondary updates (like inventory or status flags), always aggregate requirements and perform bulk queries/updates instead of iterating over the primary entities.

## 2026-04-30 - [Regex Pre-compilation and static sort allowlisting]
**Learning:** Avoid repeatedly calling `regexp.MustCompile` or allocating the same map within high-frequency functions (like request handlers or search parsers). Hoisting these to package-level variables reduces CPU cycles and memory allocations per request.
**Action:** Always check for repeated regex compilation or constant map allocations in hot paths and move them to package-level variables.

## 2026-05-14 - [Consolidating Flexible ID Lookups]
**Learning:** Flexible ID lookups (supporting both internal numeric and external string IDs) often lead to N+1 database roundtrips if handlers first resolve the ID and then call a separate service method to fetch the full object or DTO. Consolidating this into a single service-side "GetByFlexibleID" flow that returns the final DTO (including associations like line items) reduces per-request latency.
**Action:** When an API supports both internal and external IDs, implement a single service method that performs resolution and fetching of all required data in one path. Re-use resolved entities for subsequent logic (like notifications or status updates) to avoid redundant DB hits.

## 2026-05-18 - [Optimizing Single-Order Inventory Sync]
**Learning:** Even within single-entity operations (like a webhook processing one order), internal child loops (e.g., iterating over line items/SKUs) can create N+1 query patterns. Batching lookups for mappings and batching the resulting audit logs reduces database roundtrips from O(3N) to O(N+2).
**Action:** Always look for loops within repository methods that perform DB lookups or inserts. Replace them with batch queries (IN clauses) and batch inserts (passing slices to Create).
## 2026-06-12 - [Batching Global Sync for Orders]
**Learning:** Sequential inventory synchronization within `SyncService.Sync` looping over `affectedIDs` to trigger `GlobalSync` causes an N+1 query issue since it internally queries `GetItemByID`.
**Action:** Always batch lookups using an `IN` clause via a repository method like `GetItemsByIDs` and then perform synchronization processing in bulk to eliminate N+1 DB calls.
## 2026-06-12 - [ESLint 9 Flat Config Bug in Vite React TS]
**Learning:** Across frontend projects, trying to extend `reactHooks.configs.flat.recommended` with ESLint 9 Flat Config causes a `TypeError: Cannot read properties of undefined (reading 'recommended')` because `reactHooks.configs` does not natively have a `flat` property yet in some versions.
**Action:** When updating ESLint configs for frontend projects, instead of relying on the broken `flat` property on `eslint-plugin-react-hooks`, manually configure the `plugins: { 'react-hooks': reactHooks }` and spread the rules `...reactHooks.configs.recommended.rules` inside the flat config object.
## 2026-06-14 - [Batching Meta Template Status Sync]
**Learning:** Sequential template synchronization within `SyncStatus` looping over templates to trigger `GetRemoteTemplateByName` causes an N+1 API call issue.
**Action:** Always batch lookups using a single API call like `GetAllRemoteTemplates` and then perform synchronization processing in bulk to eliminate N+1 API calls.
## 2026-06-17 - Optimize BulkDeleteCustomers
**Learning:** Sequential deletion of records with external API calls (Shopify) and database operations creates an N+1 performance bottleneck.
**Action:** Use golang.org/x/sync/errgroup to parallelize external API calls and batch the database deletion with GORM to significantly improve performance.

## 2025-02-18 - ⚡ Bolt: Synchronous Marketing Send in Loop
**Learning:** Sending external API calls in a sequential loop creates an O(N) bottleneck.
**Action:** Replaced sequential loop in `SendBulkMarketing` (`handlers.go`) with concurrent execution using `golang.org/x/sync/errgroup` with a limit of 5 and `sync/atomic` for safe metric aggregation. Benchmark showed an improvement from ~530ms to ~53ms for 100 iterations.

## 2026-06-19 - [Batching GlobalSync calls]
**Learning:** Sequential inventory synchronization within `OrderService` and `ManufacturingService` loops over affected items to trigger `GlobalSync`, causing an N+1 query issue since it internally queries `GetItemByID`.
**Action:** Always batch lookups using a single `GlobalSyncBatch` call to process multiple items efficiently and eliminate N+1 database queries. When adding batched methods to interfaces, remember to update the corresponding local interface definitions in service files.

## 2026-06-25 - [Fixing N+1 Queries in Shopify Price Sync with GORM Associations]
**Learning:** Resolving N+1 query loops by pre-fetching associated local entities (like `InventoryItem` with `Mappings`) and bulk updating their fields (like `Price`) using GORM's `clause.OnConflict` can trigger primary key duplicate violations. Because the entities contain populated association slices, GORM attempts to re-insert them during the `Create` operation.
**Action:** When performing bulk upserts or updates with `gorm.clause.OnConflict` on full entities loaded from the database, always use `.Omit(clause.Associations)` (e.g., `db.Omit(clause.Associations).Clauses(...)`) to restrict the operation to the parent model and prevent association constraint errors. Also, use a map of entity pointers (`map[int]*entity.InventoryItem`) when staging updates to avoid struct copy bugs that lose intermediate changes.

## 2026-06-21 - [Parallelize External API Calls in GlobalSyncBatch]
**Learning:** Performing multiple independent external API calls sequentially (like pushing inventory levels to Shopify and Amazon for an entire batch of items in `GlobalSyncBatch`) creates a severe $O(N)$ performance bottleneck, limited by network latency on every individual request.
**Action:** When iterating over items in a batch that each require external API updates, use `golang.org/x/sync/errgroup` to parallelize the requests. Limit concurrency to `5` (rather than 10) to remain consistent with established rate-limiting guards in the codebase (e.g. `customer_service.go`, `handlers.go`) and prevent triggering rate limit errors from external platforms like Shopify. Always move common dependencies (like fetching a location ID) outside the parallel loop.

## 2026-06-28 - [Batching Meta Template Status Sync updates]
**Learning:** Sequential database updates within `SyncStatus` looping over templates to trigger `UpdateStatus` causes an N+1 query issue since it internally queries the database per template.
**Action:** Always batch updates using a single `BulkUpdateStatuses` call and a map of statuses to eliminate N+1 queries.
