1. Add test coverage for `BulkDeleteCustomers` to measure baseline performance.
2. Refactor `BulkDeleteCustomers` to avoid N+1 issues.
   - Fetch customers to find Shopify mappings in one `GetByIDs` call. (Note: repo uses `[]uint` for `GetByIDs`, we will need to convert `[]int64` to `[]uint` or update the repo/service appropriately)
   - Perform concurrent external API deletions (`s.shopifyClient.DeleteCustomer`) using `errgroup` or a standard wait group with a worker pool (e.g., concurrency limit 5).
   - Perform single batch DB delete via `s.repo.BulkDelete(ctx, ids)`.
3. Complete pre-commit instructions.
4. Verify tests pass and submit PR.
