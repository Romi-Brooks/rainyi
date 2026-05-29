package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"size:100;not null;default:'用户'" json:"username"`
	Email     string    `gorm:"size:200;uniqueIndex;not null" json:"email"`
	Avatar    string    `gorm:"size:500" json:"avatar"`
	Password  string    `gorm:"size:200;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Conversation struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64      `gorm:"index;not null" json:"user_id"`
	Title       string     `gorm:"size:200;not null;default:'情感陪伴'" json:"title"`
	AINickname  string     `gorm:"size:100;not null;default:'RainYi'" json:"ai_nickname"`
	AIAvatar    string     `gorm:"size:500" json:"ai_avatar"`
	PersonaID   *int64     `gorm:"index" json:"persona_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastMessage *LastMessage `gorm:"-" json:"last_message,omitempty"`
}

type LastMessage struct {
	Content   string    `json:"content"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID int64     `gorm:"index;not null" json:"conversation_id"`
	Role           string    `gorm:"size:20;not null" json:"role"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	HasAttachment  bool      `gorm:"default:false" json:"has_attachment"`
	AttachmentType string    `gorm:"size:20" json:"attachment_type"`
	AttachmentURL  string    `gorm:"size:500" json:"attachment_url"`
	CreatedAt      time.Time `json:"created_at"`
	IsDeleted      bool      `json:"is_deleted"`
}

type Persona struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64          `gorm:"index;default:0" json:"user_id"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Nickname    string         `gorm:"size:200" json:"nickname"`
	Description string         `gorm:"size:500" json:"description"`
	DirName     string         `gorm:"size:200" json:"dir_name"`
	Avatar      string         `gorm:"size:500" json:"avatar"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type PersonaFile struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PersonaID     int64     `gorm:"index;not null" json:"persona_id"`
	FileName      string    `gorm:"size:255;not null" json:"file_name"`
	MinioPath     string    `gorm:"size:500;not null" json:"minio_path"`
	Priority      int       `gorm:"default:0" json:"priority"`
	ModuleCategory string   `gorm:"size:100" json:"module_category"`
	FileSize      int64     `json:"file_size"`
	CreatedAt     time.Time `json:"created_at"`
}

type FileRecord struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64          `gorm:"index;not null" json:"user_id"`
	FileType      string         `gorm:"size:30;not null;index" json:"file_type"`
	ReferenceID   int64          `gorm:"index" json:"reference_id"`
	ReferenceType string         `gorm:"size:30;index" json:"reference_type"`
	OriginalName  string         `gorm:"size:255" json:"original_name"`
	StoragePath   string         `gorm:"size:500;not null" json:"storage_path"`
	URL           string         `gorm:"size:500" json:"url"`
	Size          int64          `json:"size"`
	MimeType      string         `gorm:"size:100" json:"mime_type"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{}, &Conversation{}, &Message{}, &Persona{}, &PersonaFile{}, &FileRecord{})
}
