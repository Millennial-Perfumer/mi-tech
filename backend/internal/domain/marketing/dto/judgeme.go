package dto

import "mi-tech/internal/domain/marketing/entity"

// GenerateReviewsRequest contains parameters to generate review drafts.
type GenerateReviewsRequest struct {
	ShopDomain        string   `json:"shop_domain"`
	Email             string   `json:"email"`
	AliasEmail        bool     `json:"alias_email"`
	ReviewsPerProduct int      `json:"reviews_per_product"`
	ProductIDs        []string `json:"product_ids,omitempty"`
}

// GeneratedReviewDTO represents a single draft review item.
type GeneratedReviewDTO struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	ProductTitle string `json:"product_title"`
	ReviewerName string `json:"reviewer_name"`
	Gender       string `json:"gender"`
	Email        string `json:"email"`
	Rating       int    `json:"rating"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	ShopDomain   string `json:"shop_domain"`
}

// SubmitReviewsRequest contains review items to post to Judge.me.
type SubmitReviewsRequest struct {
	Reviews []GeneratedReviewDTO `json:"reviews"`
	DelayMs int                  `json:"delay_ms"`
	DryRun  bool                 `json:"dry_run"`
}

// SubmissionResultDTO contains status log of each submitted review.
type SubmissionResultDTO struct {
	Index        int    `json:"index"`
	ProductID    string `json:"product_id"`
	ProductTitle string `json:"product_title"`
	ReviewerName string `json:"reviewer_name"`
	Email        string `json:"email"`
	Status       string `json:"status"`
	StatusCode   int    `json:"status_code"`
	ResponseBody string `json:"response_body"`
}

// SubmitReviewsResponse contains execution summary.
type SubmitReviewsResponse struct {
	TotalProcessed int                   `json:"total_processed"`
	Successful     int                   `json:"successful"`
	Failed         int                   `json:"failed"`
	Results        []SubmissionResultDTO `json:"results"`
}

// PublishedReviewsResponse wraps paginated published reviews from DB.
type PublishedReviewsResponse struct {
	Total   int64                    `json:"total"`
	Page    int                      `json:"page"`
	Limit   int                      `json:"limit"`
	Reviews []entity.PublishedReview `json:"reviews"`
}
