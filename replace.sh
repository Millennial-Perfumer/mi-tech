#!/bin/bash
cat << 'DIFF' > patch.diff
--- backend/internal/domain/order/service/customer_service.go
+++ backend/internal/domain/order/service/customer_service.go
@@ -14,6 +14,7 @@
	"strings"
	"time"

+	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
 )

@@ -348,12 +349,29 @@
 }

 func (s *CustomerService) BulkDeleteCustomers(ctx context.Context, ids []int64) error {
-	for _, id := range ids {
-		// Use DeleteCustomer to ensure Shopify sync per customer
-		if err := s.DeleteCustomer(ctx, id); err != nil {
-			log.Printf("BulkDelete: Failed to delete customer %d: %v", id, err)
+	uintIDs := make([]uint, len(ids))
+	for i, id := range ids {
+		uintIDs[i] = uint(id)
+	}
+
+	customers, err := s.repo.GetByIDs(ctx, uintIDs)
+	if err != nil {
+		return fmt.Errorf("failed to fetch customers for bulk delete: %w", err)
+	}
+
+	if s.shopifyClient != nil {
+		g, _ := errgroup.WithContext(ctx)
+		g.SetLimit(5)
+		for _, cust := range customers {
+			cust := cust
+			if cust.ExternalID != nil && *cust.ExternalID != "" {
+				extID, _ := strconv.ParseInt(*cust.ExternalID, 10, 64)
+				if extID > 0 {
+					g.Go(func() error {
+						if err := s.shopifyClient.DeleteCustomer(extID); err != nil {
+							log.Printf("Failed to sync customer deletion to Shopify for ID %d: %v", cust.ID, err)
+						}
+						return nil
+					})
+				}
+			}
		}
+		_ = g.Wait()
	}
-	return nil
+
+	return s.repo.BulkDelete(ctx, ids)
 }

 func (s *CustomerService) ExportMetaCSV(ctx context.Context, boughtOnly bool) ([]byte, error) {
DIFF
patch -p0 < patch.diff
