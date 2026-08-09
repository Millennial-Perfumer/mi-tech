package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"mi-tech/internal/domain/marketing/service"
)

type SMMHandler struct {
	socialService service.SocialService
}

func NewSMMHandler(socialService service.SocialService) *SMMHandler {
	return &SMMHandler{socialService: socialService}
}

// GetOverview returns historical insights for a platform.
// @Summary Social Media Overview
// @Description Fetch historical engagement metrics and growth insights for FB, IG, or Threads.
// @Tags social
// @Security Bearer
// @Produce json
// @Param platform query string true "Platform (facebook, instagram, threads)"
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} map[string]interface{}
// @Router /marketing/smm/overview [get]
func (h *SMMHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if platform == "" {
		http.Error(w, "platform is required", http.StatusBadRequest)
		return
	}

	overview, err := h.socialService.GetOverview(platform, startDate, endDate)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"overview": overview,
	})
}

// CheckHealth performs a real-time audit of Meta asset alignment.
// @Summary SMM Health Check
// @Description Audit the visibility and linkage of configured Meta Page and Instagram IDs.
// @Tags social
// @Security Bearer
// @Produce json
// @Success 200 {object} service.AssetHealth
// @Router /marketing/smm/health [get]
func (h *SMMHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.socialService.CheckAssetHealth()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"health":  health,
	})
}

// PostContent handles cross-platform posting.
// @Summary Cross-platform Post
// @Description Publish content to FB, IG, or Threads.
// @Tags social
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body map[string]string true "Post content"
// @Success 200 {object} map[string]interface{}
// @Router /marketing/smm/post [post]
func (h *SMMHandler) PostContent(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	platform := body["platform"]
	if platform == "" {
		http.Error(w, "platform is required", http.StatusBadRequest)
		return
	}

	postID, err := h.socialService.PostContent(platform, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"post_id": postID,
	})
}

// Sync triggers a manual data fetch from Meta.
// @Summary Sync Social Data
// @Description Manually trigger a deep sync of metrics and post history.
// @Tags social
// @Security Bearer
// @Produce json
// @Param platform query string true "Platform (facebook, instagram, threads)"
// @Success 200 {object} map[string]interface{}
// @Router /marketing/smm/sync [post]
func (h *SMMHandler) Sync(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	if platform == "" {
		http.Error(w, "platform is required", http.StatusBadRequest)
		return
	}

	// 1. Sync daily daily metrics
	if err := h.socialService.SyncPlatformMetrics(platform); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Sync historical posts (async-like but sequential for simplicity here)
	if err := h.socialService.SyncHistoricalInsights(platform); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Sync completed successfully",
	})
}

// GetPostInsights returns granular insights for a specific post.
func (h *SMMHandler) GetPostInsights(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "post id is required", http.StatusBadRequest)
		return
	}
	mediaType := r.URL.Query().Get("media_type")

	insights, err := h.socialService.GetPostInsights(postID, mediaType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"insights": insights,
	})
}

// QueuePost handles creating a new Google Drive queue post item.
// @Summary Add to Social Queue
// @Description Queue a photo, carousel, or video with caption/hashtags for auto-publishing via Google Drive and n8n.
// @Tags social
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body service.CreateQueueInput true "Queue Post Input"
// @Success 200 {object} map[string]interface{}
// @Router /marketing/smm/queue [post]
func (h *SMMHandler) QueuePost(w http.ResponseWriter, r *http.Request) {
	var input service.CreateQueueInput
	var fileHeaders []*multipart.FileHeader

	contentType := r.Header.Get("Content-Type")
	log.Printf("[SMM Queue Log] Incoming request Content-Type: %s", contentType)

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(50 << 20); err == nil {
			input.Caption = r.FormValue("caption")
			input.Hashtags = r.FormValue("hashtags")
			input.PostType = r.FormValue("post_type")

			if targetPlatforms := r.FormValue("target_platforms"); targetPlatforms != "" {
				_ = json.Unmarshal([]byte(targetPlatforms), &input.TargetPlatforms)
			}

			if r.MultipartForm != nil && r.MultipartForm.File != nil {
				files := r.MultipartForm.File["files"]
				for _, f := range files {
					input.MediaFilenames = append(input.MediaFilenames, f.Filename)
					fileHeaders = append(fileHeaders, f)
				}
			}
		} else {
			log.Printf("[SMM Queue Log] Error parsing multipart form: %v", err)
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			log.Printf("[SMM Queue Log] Error decoding JSON body: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	log.Printf("[SMM Queue Log] Processing post - type: %s, caption: %q, hashtags: %q, files count: %d", input.PostType, input.Caption, input.Hashtags, len(fileHeaders))

	post, err := h.socialService.CreateQueueItem(input)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("[SMM Queue Log] Failed to insert DB post record: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Stream files directly from RAM memory into Google Drive (Zero Disk Writes)
	if post != nil {
		var inMemoryFiles []service.InMemoryFile
		for _, fHeader := range fileHeaders {
			file, err := fHeader.Open()
			if err == nil {
				if seeker, ok := file.(io.Seeker); ok {
					_, _ = seeker.Seek(0, io.SeekStart)
				}
				data, err := io.ReadAll(file)
				file.Close()
				if err == nil {
					inMemoryFiles = append(inMemoryFiles, service.InMemoryFile{
						Name:     fHeader.Filename,
						Data:     data,
						MimeType: fHeader.Header.Get("Content-Type"),
					})
				}
			}
		}
		log.Printf("[SMM Queue Log] Read %d media files into RAM memory for direct GDrive streaming", len(inMemoryFiles))

		// Fetch Drive Refresh Token, Client ID, Secret, Service Account JSON, or OAuth token from DB / Env
		refreshToken, _ := h.socialService.GetAppConfig("gdrive_refresh_token")
		if refreshToken == "" {
			refreshToken = os.Getenv("GDRIVE_REFRESH_TOKEN")
		}

		clientID, _ := h.socialService.GetAppConfig("gdrive_client_id")
		if clientID == "" {
			clientID = os.Getenv("GDRIVE_CLIENT_ID")
		}

		clientSecret, _ := h.socialService.GetAppConfig("gdrive_client_secret")
		if clientSecret == "" {
			clientSecret = os.Getenv("GDRIVE_CLIENT_SECRET")
		}

		saJSON, _ := h.socialService.GetAppConfig("gdrive_service_account_json")
		if saJSON == "" {
			saJSON = os.Getenv("GDRIVE_SERVICE_ACCOUNT_JSON")
		}

		driveToken, _ := h.socialService.GetAppConfig("gdrive_access_token")
		if driveToken == "" {
			driveToken = os.Getenv("GDRIVE_ACCESS_TOKEN")
		}

		n8nWebhookURL, _ := h.socialService.GetAppConfig("n8n_webhook_url")
		if n8nWebhookURL == "" {
			n8nWebhookURL = os.Getenv("N8N_WEBHOOK_URL")
		}

		parentFolderID := "1djXkok8cuP3efyurTd2nOwoKRo-HpEC3"

		// Execute Direct Google Drive REST API Streaming in Background (100% In-Memory)
		go func(rToken, cID, cSecret, sa, token, pID, fName, caption, hashtags string, memFiles []service.InMemoryFile) {
			activeToken := token

			// 1. Try Refresh Token Exchange (Highest Priority - 24/7 User Account Upload)
			if rToken != "" {
				freshToken, err := service.GetAccessTokenFromRefreshToken(cID, cSecret, rToken)
				if err == nil && freshToken != "" {
					activeToken = freshToken
					log.Printf("[GDrive Auth Success] Generated fresh Google OAuth access token using Refresh Token")
				} else {
					log.Printf("[GDrive Refresh Token Error] %v", err)
				}
			}

			// 2. Fallback to Service Account JSON
			if activeToken == "" && sa != "" {
				freshToken, err := service.GetAccessTokenFromServiceAccountJSON(sa)
				if err != nil {
					log.Printf("[GDrive Service Account Error] Failed to generate access token: %v", err)
					return
				}
				activeToken = freshToken
			}

			if activeToken == "" {
				log.Printf("[SMM Queue Log] Notice: Neither gdrive_refresh_token, gdrive_service_account_json nor gdrive_access_token is configured in Settings. Direct upload skipped.")
				return
			}

			log.Printf("[SMM Queue Log] Starting 100%% in-memory Google Drive stream for folder %s into parent %s...", fName, pID)
			folderID, err := service.UploadInMemoryPackageToDrive(activeToken, pID, fName, caption, hashtags, memFiles)
			if err != nil {
				log.Printf("[GDrive Stream Error] %v", err)
			} else {
				log.Printf("[GDrive Stream Success] Streamed subfolder %s (ID: %s) directly to Google Drive (Zero Disk Writes)", fName, folderID)
			}
		}(refreshToken, clientID, clientSecret, saJSON, driveToken, parentFolderID, post.FolderName, post.Caption, post.Hashtags, inMemoryFiles)

		// 2. n8n Webhook Trigger
		if n8nWebhookURL == "" {
			log.Printf("[SMM Queue Log] Notice: n8n_webhook_url is empty in Settings/env. Webhook dispatch skipped.")
		} else {
			go dispatchQueueToN8n(n8nWebhookURL, post.FolderName, "")
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"post":    post,
	})
}

// dispatchQueueToN8n sends an asynchronous webhook notification to n8n for auto GDrive upload
func dispatchQueueToN8n(webhookURL string, folderName string, folderPath string) {
	log.Printf("[SMM Webhook Log] Dispatching queue folder %s to n8n webhook: %s", folderName, webhookURL)
	req, err := http.NewRequest("POST", webhookURL, strings.NewReader(fmt.Sprintf(`{"folder_name":"%s","path":"%s"}`, folderName, folderPath)))
	if err != nil {
		log.Printf("[SMM Webhook Log] Failed to build webhook request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[SMM Webhook Log] Webhook call failed: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[SMM Webhook Log] n8n Webhook response status: %s", resp.Status)
}

// GetQueue returns queued posts from the database.
// @Summary List Social Queue Posts
// @Description Fetch active queued posts and their publishing status.
// @Tags social
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /marketing/smm/queue [get]
func (h *SMMHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	posts, err := h.socialService.GetQueueItems(20)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"posts":   posts,
	})
}

