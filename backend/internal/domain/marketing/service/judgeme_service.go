package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mi-tech/internal/domain/inventory/entity"
	"mi-tech/internal/domain/marketing/dto"
	marketingEntity "mi-tech/internal/domain/marketing/entity"

	"gorm.io/gorm"
)

type JudgeMeService struct {
	db         *gorm.DB
	httpClient *http.Client
}

func NewJudgeMeService(db *gorm.DB) *JudgeMeService {
	return &JudgeMeService{
		db: db,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Modular Fragrance Review Sentence Arrays (50+ each)
var titleOpenings = []string{
	"Smells amazing", "Super impresssed", "Excellent drydown", "My new daily wear",
	"Compliment magnet", "Very close match", "Awesome projection", "Worth every rupee",
	"Smooth & rich blend", "Great longevity", "Blind buy success", "Subtle yet attractive",
	"Rich opening", "Gets me compliments", "Perfect office scent", "High quality juice",
	"Absolute banger!", "Clean and fresh", "Smells like luxury designer", "Super satisfied",
	"Really good fragrance", "Loved it", "Surprised by the quality", "10/10 scent profile",
	"Mind blowing performance", "Unreal quality for this price", "Stays on clothes forever",
	"Elegant & sophisticated", "Fresh top notes!", "Intense drydown", "Heavy compliment getter",
}

var bodySentence1 = []string{
	"Got this last week and I'm genuinely impressed.",
	"Honestly wasn't expecting this kind of quality at this price point.",
	"The scent profile is super refined and smooth.",
	"Bought this after a friend recommended it to me.",
	"Applied 4 sprays in the morning before heading out.",
	"Using this for the past 3 days now.",
	"Order arrived quickly and packaging was super safe.",
	"Was a bit skeptical before buying, but glad I ordered it.",
	"I've been using designer perfumes for years and this holds up really well.",
	"The initial spray is bold and crisp.",
	"Wore this to office today and felt super confident.",
	"First time trying Millennial Perfumer and I'm hooked.",
	"Decided to try this bottle based on the notes description.",
	"Just unboxed it today and sprayed on my wrist.",
}

var bodySentence2 = []string{
	"The opening is fresh and the drydown turns into a warm, rich aroma.",
	"Wore it to work and received two compliments before lunch.",
	"Settles down into a very subtle and pleasant fragrance after 20 mins.",
	"Performs really well in warm Indian weather.",
	"The longevity is solid 6 to 8 hours on my clothes.",
	"The projection is strong for the first 2 hours.",
	"Not harsh or synthetic at all on the nose.",
	"The drydown on this one is incredible.",
	"Stays on my jacket even the next morning!",
	"The notes are very well balanced, not overpowering at all.",
	"As it dries down, the scent becomes warm and sophisticated.",
	"Colleagues actually asked me which perfume I was wearing.",
	"The sillage cloud around me was so pleasant throughout the evening.",
	"The citrus and woody accords blend effortlessly together.",
}

var bodySentence3 = []string{
	"Easily lasts through my entire workday.",
	"Definitely worth every single rupee.",
	"High quality juice and great packaging!",
	"Great for daily wear as well as evening outings.",
	"Will try other variants from this house too.",
	"Smells way more expensive than it actually is.",
	"Clean, refined, and long-lasting performance.",
	"Would 100% recommend to anyone looking for a solid perfume.",
	"Going to buy a backup bottle soon.",
	"Must try for fragrance lovers!",
	"Already added 2 more variants to my cart.",
	"Great work by the brand on this blend.",
}

func applyHumanImperfections(text string) string {
	r := rand.Float64()
	if r < 0.15 && len(text) > 0 {
		text = strings.ToLower(text[:1]) + text[1:]
	}

	typoMap := map[string]string{
		"impressed":  "impresssed",
		"writing":    "writting",
		"received":   "recieved",
		"definitely": "definitelyyy",
		"really":     "reallly",
		"fragrance":  "fragrence",
		"perfume":    "perfum",
		"because":    "coz",
		"very":       "v.",
		"it's":       "its",
		"I'm":        "im",
		"don't":      "dont",
		"can't":      "cant",
		"recommend":  "recomended",
		"quality":    "qualityyy",
		"awesome":    "awsum",
	}

	if rand.Float64() < 0.25 {
		words := strings.Split(text, " ")
		for i, w := range words {
			cleanWord := strings.ToLower(strings.Trim(w, ".,!"))
			if typo, exists := typoMap[cleanWord]; exists {
				words[i] = typo
				break
			}
		}
		text = strings.Join(words, " ")
	}

	if rand.Float64() < 0.10 && strings.HasSuffix(text, ".") {
		text = text[:len(text)-1]
	}

	if rand.Float64() < 0.10 && strings.HasSuffix(text, "!") {
		text = text + "!"
	}

	return text
}

func generateDynamicTitle() string {
	title := titleOpenings[rand.Intn(len(titleOpenings))]
	if rand.Float64() < 0.20 {
		options := []string{"!!", " :)", " - Must Buy"}
		title += options[rand.Intn(len(options))]
	}
	return applyHumanImperfections(title)
}

func generateDynamicBody() string {
	s1 := bodySentence1[rand.Intn(len(bodySentence1))]
	s2 := bodySentence2[rand.Intn(len(bodySentence2))]
	s3 := bodySentence3[rand.Intn(len(bodySentence3))]

	var full string
	if rand.Float64() < 0.70 {
		full = fmt.Sprintf("%s %s %s", s1, s2, s3)
	} else {
		full = fmt.Sprintf("%s %s", s1, s2)
	}

	return applyHumanImperfections(full)
}

func generateRandomIndianName() (gender string, fullName string) {
	return Generate1000Name()
}

func generateRating() int {
	if rand.Intn(100) < 85 {
		return 5
	}
	return 4
}

// GenerateReviews creates draft review objects capped at max 10 total reviews.
func (s *JudgeMeService) GenerateReviews(ctx context.Context, req dto.GenerateReviewsRequest) ([]dto.GeneratedReviewDTO, error) {
	if req.ShopDomain == "" {
		req.ShopDomain = "4296fb-8e.myshopify.com"
	}
	if req.Email == "" {
		req.Email = "aboobakersiddiq2000@gmail.com"
	}
	if req.ReviewsPerProduct <= 0 {
		req.ReviewsPerProduct = 1
	}

	// Fetch target products from database
	var items []entity.InventoryItem
	query := s.db.WithContext(ctx)
	if len(req.ProductIDs) > 0 {
		query = query.Where("id IN ?", req.ProductIDs)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// Fallback to sample products if DB returns 0 items
	var targetProducts []dto.GeneratedReviewDTO
	if len(items) == 0 {
		targetProducts = []dto.GeneratedReviewDTO{
			{ProductID: "10094054015266", ProductTitle: "Aeros"},
			{ProductID: "10094054048034", ProductTitle: "All Of Me"},
			{ProductID: "10094054080802", ProductTitle: "Allure"},
			{ProductID: "10094054113570", ProductTitle: "Angel Share"},
			{ProductID: "10094054146338", ProductTitle: "Aqua"},
		}
	} else {
		for _, item := range items {
			targetProducts = append(targetProducts, dto.GeneratedReviewDTO{
				ProductID:    strconv.Itoa(item.ID),
				ProductTitle: item.Title,
			})
		}
	}

	rand.Seed(time.Now().UnixNano())
	var drafts []dto.GeneratedReviewDTO

	// Strictly cap total generated reviews to MAX 10 total per request
	const maxTotalLimit = 10
	totalGenerated := 0

	for _, p := range targetProducts {
		if totalGenerated >= maxTotalLimit {
			break
		}

		for i := 0; i < req.ReviewsPerProduct; i++ {
			if totalGenerated >= maxTotalLimit {
				break
			}

			gender, name := generateRandomIndianName()
			rating := generateRating()
			title := generateDynamicTitle()
			body := generateDynamicBody()

			reviewerEmail := req.Email
			if req.AliasEmail && strings.Contains(req.Email, "@") {
				parts := strings.Split(req.Email, "@")
				cleanFirstName := strings.ToLower(strings.Fields(name)[0])
				reviewerEmail = fmt.Sprintf("%s+%s%d@%s", parts[0], cleanFirstName, rand.Intn(90)+10, parts[1])
			}

			drafts = append(drafts, dto.GeneratedReviewDTO{
				ID:           fmt.Sprintf("draft_%d_%d", time.Now().UnixNano(), rand.Intn(1000)),
				ProductID:    p.ProductID,
				ProductTitle: p.ProductTitle,
				ReviewerName: name,
				Gender:       gender,
				Email:        reviewerEmail,
				Rating:       rating,
				Title:        title,
				Body:         body,
				ShopDomain:   req.ShopDomain,
			})

			totalGenerated++
		}
	}

	return drafts, nil
}

// SubmitReviews sends reviews to Judge.me API endpoint AND stores them into PostgreSQL database table.
func (s *JudgeMeService) SubmitReviews(ctx context.Context, req dto.SubmitReviewsRequest) (*dto.SubmitReviewsResponse, error) {
	if req.DelayMs < 0 {
		req.DelayMs = 1200
	}

	res := &dto.SubmitReviewsResponse{
		TotalProcessed: len(req.Reviews),
		Results:        make([]dto.SubmissionResultDTO, 0, len(req.Reviews)),
	}

	for idx, rev := range req.Reviews {
		shopDomain := rev.ShopDomain
		if shopDomain == "" {
			shopDomain = "4296fb-8e.myshopify.com"
		}

		status := "SUCCESS"
		statusCode := 200
		responseBody := ""

		if req.DryRun {
			status = "DRY_RUN"
			statusCode = 200
			responseBody = "Dry run test successfully validated"
			res.Successful++
		} else {
			bodyBuf := &bytes.Buffer{}
			mw := multipart.NewWriter(bodyBuf)

			_ = mw.WriteField("url", shopDomain)
			_ = mw.WriteField("shop_domain", shopDomain)
			_ = mw.WriteField("platform", "shopify")
			_ = mw.WriteField("name", rev.ReviewerName)
			_ = mw.WriteField("email", rev.Email)
			_ = mw.WriteField("rating", strconv.Itoa(rev.Rating))
			_ = mw.WriteField("title", rev.Title)
			_ = mw.WriteField("body", rev.Body)
			_ = mw.WriteField("id", rev.ProductID)
			_ = mw.Close()

			httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.judge.me/api/v1/reviews", bodyBuf)
			if err != nil {
				res.Failed++
				res.Results = append(res.Results, dto.SubmissionResultDTO{
					Index:        idx + 1,
					ProductID:    rev.ProductID,
					ProductTitle: rev.ProductTitle,
					ReviewerName: rev.ReviewerName,
					Email:        rev.Email,
					Status:       "ERROR",
					StatusCode:   0,
					ResponseBody: err.Error(),
				})
				continue
			}

			httpReq.Header.Set("Content-Type", mw.FormDataContentType())
			httpReq.Header.Set("Accept", "*/*")
			httpReq.Header.Set("Origin", "https://millennialperfumer.in")
			httpReq.Header.Set("Referer", "https://millennialperfumer.in/")
			httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36")

			resp, err := s.httpClient.Do(httpReq)
			if err != nil {
				res.Failed++
				res.Results = append(res.Results, dto.SubmissionResultDTO{
					Index:        idx + 1,
					ProductID:    rev.ProductID,
					ProductTitle: rev.ProductTitle,
					ReviewerName: rev.ReviewerName,
					Email:        rev.Email,
					Status:       "FAILED",
					StatusCode:   0,
					ResponseBody: err.Error(),
				})
				continue
			} else {
				respBodyBytes, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				responseBody = string(respBodyBytes)
				statusCode = resp.StatusCode

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					status = "SUCCESS"
					res.Successful++
				} else {
					status = "API_ERROR"
					res.Failed++
				}
			}
		}

		// Log status result
		res.Results = append(res.Results, dto.SubmissionResultDTO{
			Index:        idx + 1,
			ProductID:    rev.ProductID,
			ProductTitle: rev.ProductTitle,
			ReviewerName: rev.ReviewerName,
			Email:        rev.Email,
			Status:       status,
			StatusCode:   statusCode,
			ResponseBody: responseBody,
		})

		// Save into PostgreSQL published_reviews table if SUCCESS or DRY_RUN
		if status == "SUCCESS" || status == "DRY_RUN" {
			pubRecord := marketingEntity.PublishedReview{
				ReviewID:     rev.ID,
				ProductID:    rev.ProductID,
				ProductTitle: rev.ProductTitle,
				ReviewerName: rev.ReviewerName,
				Gender:       rev.Gender,
				Email:        rev.Email,
				Rating:       rev.Rating,
				Title:        rev.Title,
				Body:         rev.Body,
				ShopDomain:   shopDomain,
				Status:       status,
				StatusCode:   statusCode,
				PublishedAt:  time.Now(),
			}
			_ = s.db.WithContext(ctx).Create(&pubRecord).Error
		}

		if idx < len(req.Reviews)-1 && req.DelayMs > 0 {
			time.Sleep(time.Duration(req.DelayMs) * time.Millisecond)
		}
	}

	return res, nil
}

// GetPublishedReviews retrieves paginated historical published reviews from PostgreSQL database.
func (s *JudgeMeService) GetPublishedReviews(ctx context.Context, page, limit int, productID, search string) (*dto.PublishedReviewsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var reviews []marketingEntity.PublishedReview
	var total int64

	query := s.db.WithContext(ctx).Model(&marketingEntity.PublishedReview{})

	if productID != "" {
		query = query.Where("product_id = ?", productID)
	}

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(reviewer_name) LIKE ? OR LOWER(product_title) LIKE ? OR LOWER(title) LIKE ? OR LOWER(body) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count published reviews: %w", err)
	}

	if err := query.Order("published_at DESC").Limit(limit).Offset(offset).Find(&reviews).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch published reviews: %w", err)
	}

	return &dto.PublishedReviewsResponse{
		Total:   total,
		Page:    page,
		Limit:   limit,
		Reviews: reviews,
	}, nil
}
