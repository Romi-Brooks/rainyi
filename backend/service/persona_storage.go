package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"rain-yi-backend/config"
	"rain-yi-backend/model"
	"rain-yi-backend/repository"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const PersonaPrefix = "ai-persona"

type PersonaStorage struct {
	client *minio.Client
	bucket string
	pfRepo *repository.PersonaFileRepository
}

func NewPersonaStorage(pfRepo *repository.PersonaFileRepository) *PersonaStorage {
	if config.AppConfig.MinioEndpoint == "" {
		log.Println("MinIO 未配置，人格文件存储不可用")
		return nil
	}

	client, err := minio.New(config.AppConfig.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AppConfig.MinioAccessKey, config.AppConfig.MinioSecretKey, ""),
		Secure: config.AppConfig.MinioUseSSL,
	})
	if err != nil {
		log.Printf("警告: PersonaStorage MinIO 客户端初始化失败: %v", err)
		return nil
	}

	bucket := config.AppConfig.MinioBucket
	if bucket == "" {
		bucket = "rain-yi"
	}

	log.Printf("PersonaStorage 已连接: %s/%s (前缀: %s)", config.AppConfig.MinioEndpoint, bucket, PersonaPrefix)
	return &PersonaStorage{
		client: client,
		bucket: bucket,
		pfRepo: pfRepo,
	}
}

func (ps *PersonaStorage) personaDir(persona *model.Persona) string {
	dirName := persona.DirName
	if dirName == "" {
		dirName = SanitizeDirName(persona.Name)
	}
	return fmt.Sprintf("ai-persona/%s/%d", dirName, persona.ID)
}

func (ps *PersonaStorage) objectPath(persona *model.Persona, fileName string) string {
	return fmt.Sprintf("%s/%s", ps.personaDir(persona), fileName)
}

func (ps *PersonaStorage) UploadMD(persona *model.Persona, fileName string, content []byte, priority int) (*model.PersonaFile, error) {
	ctx := context.Background()
	objectName := ps.objectPath(persona, fileName)

	_, err := ps.client.PutObject(ctx, ps.bucket, objectName,
		bytes.NewReader(content), int64(len(content)),
		minio.PutObjectOptions{ContentType: "text/markdown; charset=utf-8"})
	if err != nil {
		return nil, fmt.Errorf("MinIO 上传 MD 文件失败: %w", err)
	}

	category := detectModuleCategory(fileName)
	pf := &model.PersonaFile{
		PersonaID:      persona.ID,
		FileName:       fileName,
		MinioPath:      objectName,
		Priority:       priority,
		ModuleCategory: category,
		FileSize:       int64(len(content)),
	}
	if err := ps.pfRepo.Create(pf); err != nil {
		return nil, fmt.Errorf("PersonaFile 入库失败: %w", err)
	}

	return pf, nil
}

func (ps *PersonaStorage) DownloadMD(pf *model.PersonaFile) ([]byte, error) {
	ctx := context.Background()
	obj, err := ps.client.GetObject(ctx, ps.bucket, pf.MinioPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("MinIO 获取 MD 文件失败: %w", err)
	}
	defer obj.Close()

	_, err = obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("MinIO 文件不存在: %s", pf.MinioPath)
	}

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("读取 MD 文件内容失败: %w", err)
	}

	return data, nil
}

func (ps *PersonaStorage) UploadAvatar(personaID int64, fileName string, data []byte) (string, error) {
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".jpg"
	}
	objectName := fmt.Sprintf("avatar/persona/%d/%s%s", personaID, uuid.New().String(), ext)

	contentType := "image/jpeg"
	switch strings.ToLower(ext) {
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	ctx := context.Background()
	_, err := ps.client.PutObject(ctx, ps.bucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("MinIO 上传人格头像失败: %w", err)
	}

	url := fmt.Sprintf("/storage/%s/%s", ps.bucket, objectName)
	return url, nil
}

func (ps *PersonaStorage) DeleteMD(pf *model.PersonaFile) error {
	ctx := context.Background()
	err := ps.client.RemoveObject(ctx, ps.bucket, pf.MinioPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("MinIO 删除 MD 文件失败: %w", err)
	}
	return ps.pfRepo.Delete(pf.ID)
}

func (ps *PersonaStorage) DeleteAllByPersona(persona *model.Persona) error {
	files, err := ps.pfRepo.FindByPersonaID(persona.ID)
	if err != nil {
		return err
	}
	for _, pf := range files {
		ps.client.RemoveObject(context.Background(), ps.bucket, pf.MinioPath, minio.RemoveObjectOptions{})
	}
	return ps.pfRepo.DeleteByPersonaID(persona.ID)
}

func (ps *PersonaStorage) ListPersonaDirs() ([]string, error) {
	ctx := context.Background()
	objCh := ps.client.ListObjects(ctx, ps.bucket, minio.ListObjectsOptions{
		Prefix:    "ai-persona/",
		Recursive: false,
	})

	dirSet := make(map[string]bool)
	for obj := range objCh {
		if obj.Err != nil {
			continue
		}
		trimmed := strings.TrimPrefix(obj.Key, "ai-persona/")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) >= 1 {
			dirSet[parts[0]] = true
		}
	}

	var dirs []string
	for d := range dirSet {
		dirs = append(dirs, d)
	}
	return dirs, nil
}

func (ps *PersonaStorage) Bucket() string {
	return ps.bucket
}

func detectModuleCategory(fileName string) string {
	name := strings.ToLower(fileName)
	switch {
	case strings.Contains(name, "persona-base"), strings.Contains(name, "persona_base"):
		return "persona_base"
	case strings.Contains(name, "persona-tone"), strings.Contains(name, "persona_tone"):
		return "persona_tone"
	case strings.Contains(name, "forbidden"), strings.Contains(name, "rule"):
		return "forbidden_rules"
	case strings.Contains(name, "emotion"), strings.Contains(name, "companion"):
		return "emotion_companion"
	case strings.Contains(name, "professional"), strings.Contains(name, "skill"):
		return "professional_skills"
	case strings.Contains(name, "style"):
		return "style_switch"
	case strings.Contains(name, "trigger"), strings.Contains(name, "pet"):
		return "trigger_rules"
	default:
		return "general"
	}
}

func SanitizeDirName(name string) string {
	sanitized := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	if sanitized == "" {
		sanitized = "persona"
	}
	return strings.ToLower(sanitized)
}

func (ps *PersonaStorage) Client() *minio.Client {
	return ps.client
}
