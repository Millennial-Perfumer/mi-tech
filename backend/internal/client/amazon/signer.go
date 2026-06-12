package amazon

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// SignV4 signs an HTTP request using AWS Signature Version 4.
func SignV4(req *http.Request, body []byte, accessKey, secretKey, region, service string, signTime time.Time) error {
	// 1. Create a canonical request
	t := signTime.UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	req.Header.Set("X-Amz-Date", amzDate)

	// Canonical Headers
	headers := map[string]string{
		"host":       req.URL.Host,
		"x-amz-date": amzDate,
	}

	// Add other headers that should be signed
	var sortedHeaders []string
	for k := range headers {
		sortedHeaders = append(sortedHeaders, strings.ToLower(k))
	}
	sort.Strings(sortedHeaders)

	var canonicalHeaders strings.Builder
	for _, k := range sortedHeaders {
		canonicalHeaders.WriteString(k)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(headers[k])
		canonicalHeaders.WriteString("\n")
	}

	signedHeaders := strings.Join(sortedHeaders, ";")

	// Payload Hash
	h := sha256.New()
	h.Write(body)
	payloadHash := hex.EncodeToString(h.Sum(nil))

	path := req.URL.Path
	if path == "" {
		path = "/"
	}

	canonicalRequest := strings.Join([]string{
		req.Method,
		path,
		req.URL.RawQuery,
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")

	// 2. Create string to sign
	h = sha256.New()
	h.Write([]byte(canonicalRequest))
	hashedCanonicalRequest := hex.EncodeToString(h.Sum(nil))

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		hashedCanonicalRequest,
	}, "\n")

	// 3. Calculate signature
	signingKey := getSignatureKey(secretKey, dateStamp, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// 4. Add Authorization header
	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)

	return nil
}

func hmacSHA256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func getSignatureKey(key, dateStamp, regionName, serviceName string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+key), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(regionName))
	kService := hmacSHA256(kRegion, []byte(serviceName))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}

// SignRequest is a helper that reads the body, signs the request, and restores the body.
func SignRequest(req *http.Request, accessKey, secretKey, region, service string) error {
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	return SignV4(req, body, accessKey, secretKey, region, service, time.Now())
}
