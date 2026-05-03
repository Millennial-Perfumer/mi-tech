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

## 2026-05-15 - [Pushing Aggregation to Database]
**Learning:** Application-side aggregation of report data (e.g., using maps to group by HSN code and manual loops for tax splits) is less efficient than SQL-level aggregation. Pushing `GROUP BY` and conditional logic (`CASE WHEN`) into the database reduces data transfer and leverages the DB's optimized execution engine.
**Action:** Always evaluate if reporting metrics can be fully computed in a single SQL query before resorting to application-side post-processing. Use `COALESCE` and `NULLIF` to ensure consistent grouping keys when dealing with nullable or empty string columns.
