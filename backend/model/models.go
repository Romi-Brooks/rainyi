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
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64     `gorm:"index;default:0" json:"user_id"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SkillNode struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PersonaID   int64     `gorm:"index;not null" json:"persona_id"`
	ParentID    *int64    `gorm:"index" json:"parent_id"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	FileName    string    `gorm:"size:200;not null" json:"file_name"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Source      string    `gorm:"size:10;default:'db'" json:"source"`
	StoragePath string    `gorm:"size:500" json:"storage_path"`
	Priority    int       `gorm:"default:0" json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
}

type SkillKV struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SkillNodeID int64     `gorm:"index;not null" json:"skill_node_id"`
	Key         string    `gorm:"size:200;not null" json:"key"`
	Value       string    `gorm:"type:text;not null" json:"value"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
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
	db.AutoMigrate(&User{}, &Conversation{}, &Message{}, &Persona{}, &SkillNode{}, &SkillKV{}, &FileRecord{})
}
