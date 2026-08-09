package service

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetAccessTokenFromRefreshToken uses a Google OAuth Refresh Token to generate a fresh Access Token automatically
func GetAccessTokenFromRefreshToken(clientID, clientSecret, refreshToken string) (string, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return "", fmt.Errorf("refresh token is empty")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", strings.TrimSpace(refreshToken))
	if clientID != "" {
		data.Set("client_id", strings.TrimSpace(clientID))
	}
	if clientSecret != "" {
		data.Set("client_secret", strings.TrimSpace(clientSecret))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("refresh token exchange network error: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	_ = json.Unmarshal(respBody, &tokenResp)

	return tokenResp.AccessToken, nil
}

type ServiceAccountCredentials struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// GetAccessTokenFromServiceAccountJSON exchanges a Service Account JSON for a fresh 1-hour Google API access token
func GetAccessTokenFromServiceAccountJSON(jsonContent string) (string, error) {
	if strings.TrimSpace(jsonContent) == "" {
		return "", fmt.Errorf("service account json content is empty")
	}

	var sa ServiceAccountCredentials
	if err := json.Unmarshal([]byte(jsonContent), &sa); err != nil {
		return "", fmt.Errorf("invalid service account json: %v", err)
	}

	if sa.ClientEmail == "" || sa.PrivateKey == "" {
		return "", fmt.Errorf("service account json missing client_email or private_key")
	}

	tokenURI := sa.TokenURI
	if tokenURI == "" {
		tokenURI = "https://oauth2.googleapis.com/token"
	}

	now := time.Now().Unix()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims := map[string]interface{}{
		"iss":   sa.ClientEmail,
		"scope": "https://www.googleapis.com/auth/drive https://www.googleapis.com/auth/drive.file",
		"aud":   tokenURI,
		"exp":   now + 3600,
		"iat":   now,
	}
	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	unsignedToken := fmt.Sprintf("%s.%s", header, payload)

	// Parse private key
	block, _ := pem.Decode([]byte(sa.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("failed to decode private key PEM")
	}

	var parsedKey interface{}
	var err error
	parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %v", err)
		}
	}

	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	// Sign SHA256 hash
	h := sha256.New()
	h.Write([]byte(unsignedToken))
	hashed := h.Sum(nil)

	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("failed to sign token assertion: %v", err)
	}

	signature := base64.RawURLEncoding.EncodeToString(signatureBytes)
	signedJWT := fmt.Sprintf("%s.%s", unsignedToken, signature)

	// Post JWT assertion to token endpoint
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", signedJWT)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(tokenURI, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("token request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	_ = json.Unmarshal(respBody, &tokenResp)

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("service account token exchange error (%s): %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return tokenResp.AccessToken, nil
}

type InMemoryFile struct {
	Name     string
	Data     []byte
	MimeType string
}

// UploadInMemoryPackageToDrive streams caption, hashtags, and media files directly from RAM memory into Google Drive (Zero Disk Writes)
func UploadInMemoryPackageToDrive(accessToken string, parentFolderID string, folderName string, caption string, hashtags string, files []InMemoryFile) (string, error) {
	if accessToken == "" || parentFolderID == "" {
		return "", fmt.Errorf("access token or parent folder ID missing")
	}

	// 1. Create subfolder in Google Drive
	createFolderURL := "https://www.googleapis.com/drive/v3/files?supportsAllDrives=true"
	folderMeta := map[string]interface{}{
		"name":     folderName,
		"mimeType": "application/vnd.google-apps.folder",
		"parents":  []string{parentFolderID},
	}
	folderBody, _ := json.Marshal(folderMeta)

	req, err := http.NewRequest("POST", createFolderURL, bytes.NewBuffer(folderBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var folderResp struct {
		ID    string `json:"id"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&folderResp)

	if folderResp.ID == "" {
		return "", fmt.Errorf("drive folder creation error: %s", folderResp.Error.Message)
	}

	createdFolderID := folderResp.ID
	log.Printf("[GDrive Direct Stream] Created Drive subfolder %s (ID: %s)", folderName, createdFolderID)

	// 2. Upload caption.txt in-memory
	if strings.TrimSpace(caption) != "" {
		_ = uploadInMemoryBytesToDrive(client, accessToken, createdFolderID, "caption.txt", []byte(strings.TrimSpace(caption)), "text/plain; charset=utf-8")
	}

	// 3. Upload hashtags.txt in-memory
	if strings.TrimSpace(hashtags) != "" {
		_ = uploadInMemoryBytesToDrive(client, accessToken, createdFolderID, "hashtags.txt", []byte(strings.TrimSpace(hashtags)), "text/plain; charset=utf-8")
	}

	// 4. Upload binary media files directly from RAM
	for _, f := range files {
		_ = uploadInMemoryBytesToDrive(client, accessToken, createdFolderID, f.Name, f.Data, f.MimeType)
	}

	return createdFolderID, nil
}

func uploadInMemoryBytesToDrive(client *http.Client, token string, parentID string, fileName string, fileData []byte, mimeType string) error {
	boundary := "---GoogleDriveBoundary" + fmt.Sprintf("%d", time.Now().UnixNano())

	if mimeType == "" {
		mimeType = "application/octet-stream"
		ext := strings.ToLower(filepath.Ext(fileName))
		switch ext {
		case ".txt":
			mimeType = "text/plain; charset=utf-8"
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".webp":
			mimeType = "image/webp"
		case ".mp4":
			mimeType = "video/mp4"
		case ".mov":
			mimeType = "video/quicktime"
		}
	}

	metaJSON, _ := json.Marshal(map[string]interface{}{
		"name":     fileName,
		"parents":  []string{parentID},
		"mimeType": mimeType,
	})

	var body bytes.Buffer
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Type: application/json; charset=UTF-8\r\n\r\n")
	body.Write(metaJSON)
	body.WriteString("\r\n")

	body.WriteString("--" + boundary + "\r\n")
	body.WriteString(fmt.Sprintf("Content-Type: %s\r\n\r\n", mimeType))
	body.Write(fileData)
	body.WriteString("\r\n")
	body.WriteString("--" + boundary + "--\r\n")

	uploadURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&supportsAllDrives=true"
	req, err := http.NewRequest("POST", uploadURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "multipart/related; boundary="+boundary)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[GDrive Direct Stream Error] Upload request failed for %s: %v", fileName, err)
		return err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var uploadResp struct {
		ID    string `json:"id"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(respBytes, &uploadResp)

	if uploadResp.ID == "" {
		log.Printf("[GDrive Direct Stream Error] Failed uploading file %s: %s", fileName, uploadResp.Error.Message)
		return fmt.Errorf("upload error: %s", uploadResp.Error.Message)
	}

	log.Printf("[GDrive Direct Stream Success] Streamed %s (%d bytes) directly to Drive subfolder %s (File ID: %s)", fileName, len(fileData), parentID, uploadResp.ID)
	return nil
}

// UploadPackageToGoogleDrive creates folder in Google Drive and uploads caption.txt + media files
func UploadPackageToGoogleDrive(accessToken string, parentFolderID string, folderName string, targetDir string) (string, error) {
	if accessToken == "" || parentFolderID == "" {
		return "", fmt.Errorf("access token or parent folder ID missing")
	}

	// 1. Create subfolder in Google Drive
	createFolderURL := "https://www.googleapis.com/drive/v3/files?supportsAllDrives=true"
	folderMeta := map[string]interface{}{
		"name":     folderName,
		"mimeType": "application/vnd.google-apps.folder",
		"parents":  []string{parentFolderID},
	}
	folderBody, _ := json.Marshal(folderMeta)

	req, err := http.NewRequest("POST", createFolderURL, bytes.NewBuffer(folderBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var folderResp struct {
		ID    string `json:"id"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&folderResp); err != nil {
		return "", err
	}

	if folderResp.ID == "" {
		return "", fmt.Errorf("drive folder creation error: %s", folderResp.Error.Message)
	}

	createdFolderID := folderResp.ID
	log.Printf("[GDrive Direct Upload] Created Drive subfolder %s (ID: %s)", folderName, createdFolderID)

	// 2. Upload files in targetDir to createdFolderID
	files, err := os.ReadDir(targetDir)
	if err == nil {
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			filePath := filepath.Join(targetDir, f.Name())
			_ = uploadSingleFileToDrive(client, accessToken, createdFolderID, f.Name(), filePath)
		}
	}

	return createdFolderID, nil
}

func uploadSingleFileToDrive(client *http.Client, token string, parentID string, fileName string, filePath string) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("[GDrive Direct Upload Error] Cannot read local file %s: %v", filePath, err)
		return err
	}

	if len(fileData) == 0 {
		log.Printf("[GDrive Direct Upload Warning] File %s is 0 bytes on local disk!", fileName)
	}

	boundary := "---GoogleDriveBoundary" + fmt.Sprintf("%d", time.Now().UnixNano())

	mimeType := "application/octet-stream"
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".txt":
		mimeType = "text/plain; charset=utf-8"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".webp":
		mimeType = "image/webp"
	case ".mp4":
		mimeType = "video/mp4"
	case ".mov":
		mimeType = "video/quicktime"
	}

	metaJSON, _ := json.Marshal(map[string]interface{}{
		"name":     fileName,
		"parents":  []string{parentID},
		"mimeType": mimeType,
	})

	var body bytes.Buffer
	// Part 1: Metadata JSON
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Type: application/json; charset=UTF-8\r\n\r\n")
	body.Write(metaJSON)
	body.WriteString("\r\n")

	// Part 2: Media Binary Content
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString(fmt.Sprintf("Content-Type: %s\r\n\r\n", mimeType))
	body.Write(fileData)
	body.WriteString("\r\n")
	body.WriteString("--" + boundary + "--\r\n")

	uploadURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&supportsAllDrives=true"
	req, err := http.NewRequest("POST", uploadURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "multipart/related; boundary="+boundary)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[GDrive Direct Upload Error] Upload request failed for %s: %v", fileName, err)
		return err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var uploadResp struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Size  string `json:"size"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(respBytes, &uploadResp)

	if uploadResp.ID == "" {
		log.Printf("[GDrive Direct Upload Error] Failed uploading file %s: %s (Raw: %s)", fileName, uploadResp.Error.Message, string(respBytes))
		return fmt.Errorf("upload error: %s", uploadResp.Error.Message)
	}

	log.Printf("[GDrive Direct Upload Success] Successfully uploaded %s (%d bytes) to Drive subfolder %s (File ID: %s)", fileName, len(fileData), parentID, uploadResp.ID)
	return nil
}
