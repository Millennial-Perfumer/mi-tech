package amazon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenResponse represents the LWA token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// TokenManager handles LWA token exchange and caching.
type TokenManager struct {
	clientID     string
	clientSecret string
	refreshToken string

	cacheLock   sync.RWMutex
	accessToken string
	expiryTime  time.Time

	httpClient *http.Client
}

func NewTokenManager(clientID, clientSecret, refreshToken string) *TokenManager {
	return &TokenManager{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetAccessToken returns a valid access token, refreshing it if necessary.
func (tm *TokenManager) GetAccessToken() (string, error) {
	tm.cacheLock.RLock()
	if tm.accessToken != "" && time.Now().Before(tm.expiryTime) {
		token := tm.accessToken
		tm.cacheLock.RUnlock()
		return token, nil
	}
	tm.cacheLock.RUnlock()

	tm.cacheLock.Lock()
	defer tm.cacheLock.Unlock()

	// Double-check after acquiring lock
	if tm.accessToken != "" && time.Now().Before(tm.expiryTime) {
		return tm.accessToken, nil
	}

	tokenResp, err := tm.refresh()
	if err != nil {
		return "", fmt.Errorf("failed to refresh LWA token: %w", err)
	}

	tm.accessToken = tokenResp.AccessToken
	// Buffer expiry by 5 minutes
	tm.expiryTime = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return tm.accessToken, nil
}

func (tm *TokenManager) refresh() (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", tm.refreshToken)
	data.Set("client_id", tm.clientID)
	data.Set("client_secret", tm.clientSecret)

	req, err := http.NewRequest("POST", "https://api.amazon.com/auth/o2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("amazon auth error (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
