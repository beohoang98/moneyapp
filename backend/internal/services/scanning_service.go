package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
)

var (
	ErrScanningDisabled  = fmt.Errorf("scanning_disabled")
	ErrInvalidFileType   = fmt.Errorf("invalid_file_type")
	ErrInvalidStorageKey = fmt.Errorf("invalid_storage_key")
	ErrExtractionFailed  = fmt.Errorf("extraction_failed")
	ErrScanTimeout       = fmt.Errorf("scan_timeout")
	ErrTooManyScans      = fmt.Errorf("too_many_scans")
)

var scanAllowedMimes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

type ScanningService struct {
	db    *gorm.DB
	store storage.ObjectStore
	sem   chan struct{}
}

func NewScanningService(db *gorm.DB, store storage.ObjectStore) *ScanningService {
	return &ScanningService{
		db:    db,
		store: store,
		sem:   make(chan struct{}, 2),
	}
}

func (s *ScanningService) GetSettings(ctx context.Context) (*models.ScanningSettings, error) {
	var settings models.ScanningSettings
	if err := s.db.WithContext(ctx).First(&settings, 1).Error; err != nil {
		return nil, fmt.Errorf("get scanning settings: %w", err)
	}
	return &settings, nil
}

func (s *ScanningService) UpdateSettings(ctx context.Context, update *models.ScanningSettings) error {
	if update.BaseURL != "" {
		lower := strings.ToLower(update.BaseURL)
		if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			return fmt.Errorf("base_url must use http or https scheme")
		}
	}

	existing, err := s.GetSettings(ctx)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"enabled":    update.Enabled,
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}

	if update.BaseURL != "" {
		updates["base_url"] = update.BaseURL
	} else {
		updates["base_url"] = existing.BaseURL
	}

	if update.Model != "" {
		updates["model"] = update.Model
	} else {
		updates["model"] = existing.Model
	}

	if update.APIKey != "" {
		updates["api_key"] = update.APIKey
	} else {
		updates["api_key"] = existing.APIKey
	}

	return s.db.WithContext(ctx).Model(&models.ScanningSettings{}).
		Where("id = 1").Updates(updates).Error
}

func (s *ScanningService) TestConnection(ctx context.Context, baseURL, model, apiKey string) (bool, string) {
	if baseURL == "" {
		return false, "base_url is required"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	url := strings.TrimRight(baseURL, "/") + "/models"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Sprintf("Cannot reach %s", baseURL)
	}

	effectiveKey := apiKey
	if effectiveKey == "" {
		effectiveKey = os.Getenv("OPENAI_API_KEY")
	}
	if effectiveKey != "" {
		req.Header.Set("Authorization", "Bearer "+effectiveKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Cannot reach %s", baseURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return false, "Invalid API key"
	}

	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("Unexpected status %d from %s", resp.StatusCode, baseURL)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, "Failed to read response from API"
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return true, fmt.Sprintf("Connected to %s (could not parse model list)", baseURL)
	}

	for _, m := range result.Data {
		if m.ID == model {
			return true, fmt.Sprintf("Connected to %s", model)
		}
	}
	return false, fmt.Sprintf("Model %s not found. Run `ollama pull %s`", model, model)
}

func (s *ScanningService) CheckHealth(ctx context.Context) (bool, string) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return false, "Cannot read scanning settings"
	}
	if !settings.Enabled {
		return false, "Scanning is disabled. Enable it in Settings."
	}

	apiKey := settings.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	return s.TestConnection(ctx, settings.BaseURL, settings.Model, apiKey)
}

func (s *ScanningService) ScanImage(ctx context.Context, file multipart.File, filename string, contentType string) (*models.ScanResult, string, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, "", err
	}
	if !settings.Enabled {
		return nil, "", ErrScanningDisabled
	}

	if !scanAllowedMimes[contentType] {
		return nil, "", ErrInvalidFileType
	}

	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	default:
		return nil, "", ErrTooManyScans
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("read file: %w", err)
	}

	effectiveContentType := contentType
	if len(fileBytes) > 0 {
		head := fileBytes
		if len(head) > 512 {
			head = head[:512]
		}
		effectiveContentType = http.DetectContentType(head)
	}
	if !scanAllowedMimes[effectiveContentType] {
		return nil, "", ErrInvalidFileType
	}

	slog.Info("scanning image", "file_size_bytes", len(fileBytes), "content_type", contentType)

	storageKey := fmt.Sprintf("scan-tmp/%s_%s", uuid.New().String(), sanitizeFilename(filename))
	if err := s.store.Upload(ctx, storageKey, bytes.NewReader(fileBytes), int64(len(fileBytes)), contentType); err != nil {
		return nil, "", fmt.Errorf("upload temp: %w", err)
	}

	if ctx.Err() != nil {
		_ = s.store.Delete(context.Background(), storageKey)
		return nil, "", ctx.Err()
	}

	visionBytes, visionContentType, err := resizeForVision(fileBytes, effectiveContentType, 600)
	if err != nil {
		_ = s.store.Delete(context.Background(), storageKey)
		return nil, "", fmt.Errorf("resize for vision: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(visionBytes)
	dataURL := fmt.Sprintf("data:%s;base64,%s", visionContentType, b64)

	scanCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	result, err := s.callVisionAPI(scanCtx, settings, dataURL)
	if err != nil {
		_ = s.store.Delete(context.Background(), storageKey)
		if scanCtx.Err() != nil && ctx.Err() == nil {
			return nil, "", ErrScanTimeout
		}
		return nil, "", err
	}

	return result, storageKey, nil
}

func resizeForVision(src []byte, contentType string, maxLongEdge int) ([]byte, string, error) {
	if maxLongEdge <= 0 {
		return src, contentType, nil
	}

	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, "", err
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, "", fmt.Errorf("invalid image dimensions")
	}
	if int64(w)*int64(h) > 20_000_000 {
		return nil, "", fmt.Errorf("image too large")
	}

	longEdge := w
	if h > longEdge {
		longEdge = h
	}
	if longEdge <= maxLongEdge {
		// Still normalize to JPEG to keep payloads smaller and consistent for the vision model.
		var out bytes.Buffer
		if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, "", err
		}
		return out.Bytes(), "image/jpeg", nil
	}

	scale := float64(maxLongEdge) / float64(longEdge)
	newW := int(float64(w)*scale + 0.5)
	newH := int(float64(h)*scale + 0.5)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, dst, &jpeg.Options{Quality: 85}); err != nil {
		return nil, "", err
	}
	return out.Bytes(), "image/jpeg", nil
}

func (s *ScanningService) callVisionAPI(ctx context.Context, settings *models.ScanningSettings, imageDataURL string) (*models.ScanResult, error) {
	apiKey := settings.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	prompt := `Analyze this receipt/invoice image and extract the following information as JSON:
{
  "vendor": "vendor/store name",
  "date": "YYYY-MM-DD format",
  "currency": "3-letter currency code",
  "total_amount": integer amount in minor currency units (smallest denomination; e.g. cents for USD; VND has no cents),
  "line_items": [{"description": "item name", "amount": integer in minor units}],
  "confidence": {"vendor": "high|medium|low", "date": "high|medium|low", "total_amount": "high|medium|low"}
}
Return ONLY the JSON object, no other text.`

	reqBody := map[string]interface{}{
		"model":      settings.Model,
		"max_tokens": 2000,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": prompt},
					{"type": "image_url", "image_url": map[string]string{"url": imageDataURL}},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimRight(settings.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vision API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, ErrExtractionFailed
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read vision response: %w", err)
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, ErrExtractionFailed
	}

	if len(chatResp.Choices) == 0 {
		return nil, ErrExtractionFailed
	}

	content := chatResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var scanResult models.ScanResult
	if err := json.Unmarshal([]byte(content), &scanResult); err != nil {
		return nil, ErrExtractionFailed
	}

	if scanResult.TotalAmount < 0 {
		return nil, ErrExtractionFailed
	}

	if scanResult.Date != "" {
		if _, err := time.Parse("2006-01-02", scanResult.Date); err != nil {
			return nil, ErrExtractionFailed
		}
	}

	return &scanResult, nil
}

func (s *ScanningService) DeleteTempScan(ctx context.Context, storageKey string) error {
	if !strings.HasPrefix(storageKey, "scan-tmp/") {
		return ErrInvalidStorageKey
	}
	return s.store.Delete(ctx, storageKey)
}
