package amazon

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

// uriEscape escapes a string as described in the AWS SigV4 specification.
func uriEscape(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' {
			buf.WriteByte(c)
		} else {
			buf.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return buf.String()
}

func uriEscapePath(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' || c == '/' {
			buf.WriteByte(c)
		} else {
			buf.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return buf.String()
}

func cleanHeaderValue(v string) string {
	v = strings.TrimSpace(v)
	var buf strings.Builder
	var lastSpace bool
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c == ' ' || c == '\t' {
			if !lastSpace {
				buf.WriteByte(' ')
				lastSpace = true
			}
		} else {
			buf.WriteByte(c)
			lastSpace = false
		}
	}
	return buf.String()
}

// SignV4 signs an HTTP request using AWS Signature Version 4.
func SignV4(req *http.Request, body []byte, accessKey, secretKey, region, service string, signTime time.Time) error {
	// 1. Create a canonical request
	t := signTime.UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	if req.Header.Get("X-Amz-Date") == "" {
		req.Header.Set("X-Amz-Date", amzDate)
	} else {
		amzDate = req.Header.Get("X-Amz-Date")
	}

	// Host header might not be in req.Header
	if req.Header.Get("Host") == "" && req.Host != "" {
		req.Header.Set("Host", req.Host)
	} else if req.Header.Get("Host") == "" && req.URL.Host != "" {
		req.Header.Set("Host", req.URL.Host)
	}

	// Canonical Headers
	headers := make(map[string]string)
	var sortedHeaders []string

	for k, v := range req.Header {
		lk := strings.ToLower(k)

		// AWS skips authorization header from being signed
		if lk == "authorization" {
			continue
		}

		vals := make([]string, len(v))
		for i, val := range v {
			vals[i] = cleanHeaderValue(val)
		}

		valStr := strings.Join(vals, ",")

		headers[lk] = valStr
		sortedHeaders = append(sortedHeaders, lk)
	}

	sort.Strings(sortedHeaders)

	var canonicalHeaders strings.Builder
	var signedHeadersList []string
	for _, k := range sortedHeaders {
		canonicalHeaders.WriteString(k)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(headers[k])
		canonicalHeaders.WriteString("\n")
		signedHeadersList = append(signedHeadersList, k)
	}

	signedHeaders := strings.Join(signedHeadersList, ";")

	// Payload Hash
	h := sha256.New()
	h.Write(body)
	payloadHash := hex.EncodeToString(h.Sum(nil))

	// Path
	reqPath := req.URL.Path
	if req.URL.RawPath != "" {
		reqPath = req.URL.RawPath
	}
	if reqPath == "" {
		reqPath = "/"
	}

	// Normalize path
	canonicalReqPath := path.Clean(reqPath)
	if strings.HasSuffix(reqPath, "/") && canonicalReqPath != "/" {
		canonicalReqPath += "/"
	}
	canonicalPath := uriEscapePath(canonicalReqPath)

	// Query string
	query := req.URL.RawQuery
	var canonicalQuery strings.Builder
	if query != "" {
		params := strings.Split(query, "&")
		type kv struct {
			k, v string
		}
		var parsed []kv
		for _, p := range params {
			parts := strings.SplitN(p, "=", 2)
			k := parts[0]
			v := ""
			if len(parts) > 1 {
				v = parts[1]
			}

			// We handle cases where '=' is in the key or missing completely
			uk, err := url.QueryUnescape(k)
			if err == nil {
				k = uk
			}
			uv, err := url.QueryUnescape(v)
			if err == nil {
				v = uv
			}

			parsed = append(parsed, kv{uriEscape(k), uriEscape(v)})
		}

		sort.Slice(parsed, func(i, j int) bool {
			if parsed[i].k == parsed[j].k {
				return parsed[i].v < parsed[j].v
			}
			return parsed[i].k < parsed[j].k
		})

		for i, p := range parsed {
			if i > 0 {
				canonicalQuery.WriteString("&")
			}
			canonicalQuery.WriteString(p.k)
			canonicalQuery.WriteString("=")
			canonicalQuery.WriteString(p.v)
		}
	}

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalPath,
		canonicalQuery.String(),
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
