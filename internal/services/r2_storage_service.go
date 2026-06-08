package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"graduation-invitation/config"
)

const r2Region = "auto"
const r2Service = "s3"

type R2StorageService struct {
	bucket          string
	accessKeyID     string
	secretAccessKey string
	publicBaseURL   string
	endpoint        string
	client          *http.Client
}

type UploadedObject struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

func NewR2StorageService(cfg config.Config) *R2StorageService {
	endpoint := strings.TrimSpace(cfg.R2Endpoint)
	if endpoint == "" && strings.TrimSpace(cfg.R2AccountID) != "" {
		endpoint = "https://" + strings.TrimSpace(cfg.R2AccountID) + ".r2.cloudflarestorage.com"
	}
	return &R2StorageService{
		bucket:          strings.TrimSpace(cfg.R2Bucket),
		accessKeyID:     strings.TrimSpace(cfg.R2AccessKeyID),
		secretAccessKey: strings.TrimSpace(cfg.R2SecretAccessKey),
		publicBaseURL:   strings.TrimRight(strings.TrimSpace(cfg.R2PublicBaseURL), "/"),
		endpoint:        strings.TrimRight(endpoint, "/"),
		client:          &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *R2StorageService) PutObject(ctx context.Context, key string, contentType string, body []byte) (*UploadedObject, error) {
	if s.bucket == "" || s.accessKeyID == "" || s.secretAccessKey == "" || s.publicBaseURL == "" {
		return nil, errors.New("konfigurasi R2 belum lengkap")
	}
	if s.endpoint == "" {
		return nil, errors.New("R2_ENDPOINT atau R2_ACCOUNT_ID belum diisi")
	}
	key = cleanObjectKey(key)
	if key == "" {
		return nil, errors.New("object key tidak valid")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	endpointURL, err := url.Parse(s.endpoint)
	if err != nil || endpointURL.Scheme == "" || endpointURL.Host == "" {
		return nil, errors.New("R2_ENDPOINT tidak valid")
	}

	objectPath := "/" + path.Join(s.bucket, key)
	targetURL := endpointURL.ResolveReference(&url.URL{Path: objectPath})
	payloadHash := sha256Hex(body)
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, targetURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("Authorization", s.authorization(req, payloadHash, amzDate, dateStamp))

	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal upload ke R2: %w", err)
	}
	defer res.Body.Close()

	responseBody, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("R2 menolak upload (%s): %s", res.Status, strings.TrimSpace(string(responseBody)))
	}

	return &UploadedObject{
		Key: key,
		URL: s.publicBaseURL + "/" + key,
	}, nil
}

func (s *R2StorageService) authorization(req *http.Request, payloadHash string, amzDate string, dateStamp string) string {
	headers := map[string]string{
		"content-type":         req.Header.Get("Content-Type"),
		"host":                 req.URL.Host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)

	var canonicalHeaders strings.Builder
	for _, name := range names {
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(headers[name]))
		canonicalHeaders.WriteString("\n")
	}
	signedHeaders := strings.Join(names, ";")
	canonicalRequest := strings.Join([]string{
		req.Method,
		uriEncodePath(req.URL.EscapedPath()),
		"",
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")

	scope := dateStamp + "/" + r2Region + "/" + r2Service + "/aws4_request"
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(signingKey(s.secretAccessKey, dateStamp), []byte(stringToSign)))

	return "AWS4-HMAC-SHA256 Credential=" + s.accessKeyID + "/" + scope + ", SignedHeaders=" + signedHeaders + ", Signature=" + signature
}

func cleanObjectKey(key string) string {
	key = strings.TrimSpace(strings.ReplaceAll(key, "\\", "/"))
	key = strings.TrimPrefix(path.Clean("/"+key), "/")
	if key == "." {
		return ""
	}
	return key
}

func uriEncodePath(escapedPath string) string {
	parts := strings.Split(escapedPath, "/")
	for i, part := range parts {
		unescaped, err := url.PathUnescape(part)
		if err != nil {
			continue
		}
		parts[i] = strings.ReplaceAll(url.QueryEscape(unescaped), "+", "%20")
	}
	return strings.Join(parts, "/")
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func signingKey(secret string, dateStamp string) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	regionKey := hmacSHA256(dateKey, []byte(r2Region))
	serviceKey := hmacSHA256(regionKey, []byte(r2Service))
	return hmacSHA256(serviceKey, []byte("aws4_request"))
}
