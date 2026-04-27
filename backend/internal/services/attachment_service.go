package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB

var allowedMimeTypes = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
}

var validEntityTypes = map[string]bool{
	"expense": true,
	"income":  true,
	"invoice": true,
}

type AttachmentService struct {
	db    *gorm.DB
	store storage.ObjectStore
}

func NewAttachmentService(db *gorm.DB, store storage.ObjectStore) *AttachmentService {
	return &AttachmentService{db: db, store: store}
}

func (s *AttachmentService) Upload(ctx context.Context, entityType string, entityID int64, file multipart.File, header *multipart.FileHeader) (*models.Attachment, error) {
	if !validEntityTypes[entityType] {
		return nil, fmt.Errorf("invalid entity type: %s", entityType)
	}

	if header.Size > maxFileSize {
		return nil, fmt.Errorf("file exceeds the 10 MB limit")
	}

	mimeType := detectMimeType(header)
	if !allowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("only PDF, JPEG, and PNG files are allowed")
	}

	storageKey := fmt.Sprintf("%s/%d/%s_%s", entityType, entityID, uuid.New().String(), sanitizeFilename(header.Filename))

	if err := s.store.Upload(ctx, storageKey, file, header.Size, mimeType); err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	att := &models.Attachment{
		EntityType: entityType,
		EntityID:   entityID,
		Filename:   header.Filename,
		MimeType:   mimeType,
		SizeBytes:  header.Size,
		StorageKey: storageKey,
	}

	if err := s.db.WithContext(ctx).Create(att).Error; err != nil {
		if delErr := s.store.Delete(ctx, storageKey); delErr != nil {
			log.Printf("Warning: failed to clean up storage after DB insert failure: %v", delErr)
		}
		return nil, fmt.Errorf("insert attachment: %w", err)
	}
	return att, nil
}

func (s *AttachmentService) GetByID(ctx context.Context, id int64) (*models.Attachment, error) {
	var att models.Attachment
	if err := s.db.WithContext(ctx).First(&att, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query attachment: %w", err)
	}
	return &att, nil
}

func (s *AttachmentService) Download(ctx context.Context, id int64) (*models.Attachment, io.ReadCloser, error) {
	att, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	reader, err := s.store.Download(ctx, att.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("download file: %w", err)
	}

	return att, reader, nil
}

func (s *AttachmentService) ListByEntity(ctx context.Context, entityType string, entityID int64) ([]models.Attachment, error) {
	var attachments []models.Attachment
	if err := s.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&attachments).Error; err != nil {
		return nil, fmt.Errorf("query attachments: %w", err)
	}
	return attachments, nil
}

func (s *AttachmentService) Delete(ctx context.Context, id int64) error {
	att, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.store.Delete(ctx, att.StorageKey); err != nil {
		log.Printf("Warning: failed to delete file from storage (key=%s): %v", att.StorageKey, err)
	}

	result := s.db.WithContext(ctx).Delete(&models.Attachment{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete attachment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *AttachmentService) DeleteByEntity(ctx context.Context, entityType string, entityID int64) error {
	attachments, err := s.ListByEntity(ctx, entityType, entityID)
	if err != nil {
		return err
	}

	for _, att := range attachments {
		if err := s.store.Delete(ctx, att.StorageKey); err != nil {
			log.Printf("Warning: failed to delete storage object %s: %v", att.StorageKey, err)
		}
	}

	if err := s.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Delete(&models.Attachment{}).Error; err != nil {
		return fmt.Errorf("delete attachments by entity: %w", err)
	}
	return nil
}

func (s *AttachmentService) PromoteFromTemp(ctx context.Context, entityType string, entityID int64, sourceStorageKey string) (*models.Attachment, error) {
	if !validEntityTypes[entityType] {
		return nil, fmt.Errorf("invalid entity type: %s", entityType)
	}
	if !strings.HasPrefix(sourceStorageKey, "scan-tmp/") {
		return nil, fmt.Errorf("invalid source_storage_key: must start with scan-tmp/")
	}

	reader, err := s.store.Download(ctx, sourceStorageKey)
	if err != nil {
		return nil, fmt.Errorf("download temp file: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read temp file: %w", err)
	}

	parts := strings.Split(sourceStorageKey, "/")
	originalName := parts[len(parts)-1]

	mimeType := "image/jpeg"
	if strings.HasSuffix(strings.ToLower(originalName), ".png") {
		mimeType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(originalName), ".webp") {
		mimeType = "image/webp"
	}

	permKey := fmt.Sprintf("%s/%d/%s_%s", entityType, entityID, uuid.New().String(), originalName)
	if err := s.store.Upload(ctx, permKey, bytes.NewReader(data), int64(len(data)), mimeType); err != nil {
		return nil, fmt.Errorf("upload permanent: %w", err)
	}

	att := &models.Attachment{
		EntityType: entityType,
		EntityID:   entityID,
		Filename:   originalName,
		MimeType:   mimeType,
		SizeBytes:  int64(len(data)),
		StorageKey: permKey,
	}
	if err := s.db.WithContext(ctx).Create(att).Error; err != nil {
		if delErr := s.store.Delete(ctx, permKey); delErr != nil {
			log.Printf("Warning: failed to clean up promoted file: %v", delErr)
		}
		return nil, fmt.Errorf("insert attachment: %w", err)
	}

	if delErr := s.store.Delete(ctx, sourceStorageKey); delErr != nil {
		log.Printf("Warning: failed to delete temp file after promotion: %v", delErr)
	}

	return att, nil
}

func (s *AttachmentService) CountByEntity(ctx context.Context, entityType string, entityID int64) (int, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Attachment{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Count(&count).Error
	return int(count), err
}

func detectMimeType(header *multipart.FileHeader) string {
	ct := header.Header.Get("Content-Type")
	if ct != "" {
		ct = strings.Split(ct, ";")[0]
		ct = strings.TrimSpace(ct)
		if allowedMimeTypes[ct] {
			return ct
		}
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	}
	return ct
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}
