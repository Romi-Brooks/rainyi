package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"rain-yi-backend/config"
	"rain-yi-backend/model"
	"rain-yi-backend/repository"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileStorage interface {
	Save(fileType string, userID int64, referenceType string, referenceID int64, originalName string, data io.Reader, size int64) (*model.FileRecord, error)
	Delete(record *model.FileRecord) error
	GetURL(path string) string
	Get(path string) (io.ReadCloser, error)
	SaveToPath(objectName string, userID int64, fileType string, referenceType string, referenceID int64, originalName string, data io.Reader, size int64) (*model.FileRecord, error)
}

type MinioStorage struct {
	client    *minio.Client
	bucket    string
	publicURL string
	repo      *repository.FileRepository
}

func NewMinioStorage(cfg *config.Config, repo *repository.FileRepository) FileStorage {
	if cfg.MinioEndpoint == "" {
		log.Println("MinIO 未配置，文件存储不可用")
		return nil
	}

	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		log.Printf("警告: MinIO 客户端初始化失败: %v", err)
		return nil
	}

	ctx := context.Background()
	bucket := cfg.MinioBucket
	if bucket == "" {
		bucket = "rain-yi"
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		log.Printf("警告: MinIO 检查存储桶失败: %v", err)
		return nil
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Printf("警告: MinIO 创建存储桶失败: %v", err)
			return nil
		}
		log.Printf("MinIO 存储桶 '%s' 已创建", bucket)
	}

	log.Printf("MinIO 已连接: %s/%s", cfg.MinioEndpoint, bucket)
	return &MinioStorage{
		client:    client,
		bucket:    bucket,
		publicURL: strings.TrimRight(cfg.MinioPublicURL, "/"),
		repo:      repo,
	}
}

func (s *MinioStorage) Save(fileType string, userID int64, referenceType string, referenceID int64, originalName string, data io.Reader, size int64) (*model.FileRecord, error) {
	ext := filepath.Ext(originalName)
	objectName := fmt.Sprintf("%s/%d/%s%s", fileType, userID, uuid.New().String(), ext)

	contentType := detectContentType(ext)

	ctx := context.Background()
	_, err := s.client.PutObject(ctx, s.bucket, objectName, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("MinIO 上传失败: %w", err)
	}

	url := s.GetProxiedURL(objectName)

	record := &model.FileRecord{
		UserID:        userID,
		FileType:      fileType,
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		OriginalName:  originalName,
		StoragePath:   objectName,
		URL:           url,
		Size:          size,
		MimeType:      contentType,
	}

	if err := s.repo.Create(record); err != nil {
		return nil, fmt.Errorf("FileRecord 入库失败: %w", err)
	}

	return record, nil
}

func (s *MinioStorage) Delete(record *model.FileRecord) error {
	ctx := context.Background()
	err := s.client.RemoveObject(ctx, s.bucket, record.StoragePath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("MinIO 删除失败: %w", err)
	}
	return nil
}

func (s *MinioStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucket, path)
}

func (s *MinioStorage) SaveToPath(objectName string, userID int64, fileType string, referenceType string, referenceID int64, originalName string, data io.Reader, size int64) (*model.FileRecord, error) {
	contentType := detectContentType(filepath.Ext(originalName))

	ctx := context.Background()
	_, err := s.client.PutObject(ctx, s.bucket, objectName, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("MinIO 上传失败: %w", err)
	}

	url := fmt.Sprintf("/storage/%s/%s", s.bucket, objectName)

	record := &model.FileRecord{
		UserID:        userID,
		FileType:      fileType,
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		OriginalName:  originalName,
		StoragePath:   objectName,
		URL:           url,
		Size:          size,
		MimeType:      contentType,
	}

	if err := s.repo.Create(record); err != nil {
		return nil, fmt.Errorf("FileRecord 入库失败: %w", err)
	}

	return record, nil
}

func (s *MinioStorage) GetProxiedURL(path string) string {
	return fmt.Sprintf("/storage/%s/%s", s.bucket, path)
}

func (s *MinioStorage) Bucket() string {
	return s.bucket
}

func (s *MinioStorage) Get(path string) (io.ReadCloser, error) {
	ctx := context.Background()
	obj, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	objInfo, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, err
	}
	return &minioObject{obj, objInfo.Size}, nil
}

type minioObject struct {
	obj  *minio.Object
	size int64
}

func (m *minioObject) Read(p []byte) (int, error) {
	return m.obj.Read(p)
}

func (m *minioObject) Close() error {
	return m.obj.Close()
}

func (m *minioObject) Size() int64 {
	return m.size
}

func detectContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}

func SaveUploadedFile(storage FileStorage, fileType string, userID int64, referenceType string, referenceID int64, fileHeader *multipart.FileHeader) (*model.FileRecord, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败: %w", err)
	}

	return storage.Save(fileType, userID, referenceType, referenceID, fileHeader.Filename, bytes.NewReader(data), int64(len(data)))
}
