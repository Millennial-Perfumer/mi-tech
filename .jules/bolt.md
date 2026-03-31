## 2025-05-15 - [Safe Batch Processing and Data Integrity]
**Learning:** When refactoring sequential operations into batch processing (e.g., to solve N+1 query problems), it is critical to ensure that error handling remains robust and that data aggregation is preserved. Continuing a batch operation after a lookup failure can lead to data loss. Furthermore, batch processing must account for multiple updates to the same entity within a single batch by aggregating or merging data correctly before persisting.
**Action:** Always return an error or handle partial failures explicitly if a critical step in the batch process fails. Ensure that PII and metrics are merged or aggregated correctly when multiple source records affect the same target entity in a batch.

## 2025-05-16 - [Efficient GORM Batch Upsert with Partial Indexes]
**Learning:** To optimize $O(N)$ iterative upserts into a single $O(1)$ batch operation in GORM, `clause.OnConflict` is the standard approach. However, if the database uses a partial unique index (e.g., `WHERE deleted_at IS NULL`), GORM's `OnConflict` must explicitly target this index using `TargetWhere` (in GORM v1.2x) or `IndexConfig` (in newer versions) to avoid "there is no unique or exclusion constraint matching the ON CONFLICT specification" errors.
**Action:** When performing batch upserts on tables with partial indexes, always use `TargetWhere` to match the index's condition in the `ON CONFLICT` clause.

## 2026-03-27 - [Batching Line Item Upserts]
**Learning:** Iterative database operations within repository methods (like looping over line items to perform individual upserts) create significant overhead due to multiple database roundtrips. Even when using transactions, the per-row execution time adds up. GORM's native batch insert (`tx.Create(&slice)`) reduces this to a single O(1) roundtrip.
**Action:** Always prefer batch operations (`Create`, `Save`) with slices instead of loops when handling child entities or bulk datasets in GORM. Ensure slice elements are updated by index (`for i := range slice`) before the batch call.

## 2026-03-28 - [Optimizing CTEs with Window Functions]
**Learning:** Common Table Expressions (CTEs) that perform global aggregations (like `GROUP BY` on a full table) before joining and filtering are a significant performance bottleneck. PostgreSQL's window functions (e.g., `SUM(...) OVER (PARTITION BY ...)`) applied within a date-filtered subquery allow for localized aggregation on only the relevant subset of data.
**Action:** Replace global aggregation CTEs with window functions inside filtered subqueries to reduce query complexity from O(TotalRows) to O(FilteredRows).
