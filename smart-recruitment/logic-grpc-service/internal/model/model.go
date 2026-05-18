package model

import (
	"time"
)

type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Email        string    `gorm:"uniqueIndex;size:191;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Role         string    `gorm:"size:16;not null"` // hr | candidate
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

type Job struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	HRUserID    int64     `gorm:"index;not null"`
	Title       string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	Status      string    `gorm:"size:32;not null;default:active"` // active | archived
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type CandidateProfile struct {
	UserID     int64     `gorm:"primaryKey"`
	Name       string    `gorm:"size:128"`
	Phone      string    `gorm:"size:64"`
	Education  string    `gorm:"size:64"`
	School     string    `gorm:"size:128"`
	Experience string    `gorm:"type:text"`
	Skills     string    `gorm:"type:text"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

type Resume struct {
	UserID      int64     `gorm:"primaryKey"`
	ObjectKey   string    `gorm:"size:512;not null"`
	FileName    string    `gorm:"size:255;not null"`
	ContentType string    `gorm:"size:128;not null"`
	SizeBytes   int64     `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type Application struct {
	ID              int64     `gorm:"primaryKey;autoIncrement"`
	JobID           int64     `gorm:"uniqueIndex:idx_job_candidate;index;not null"`
	CandidateUserID int64     `gorm:"uniqueIndex:idx_job_candidate;index;not null"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

type ChatMessage struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	HRUserID  int64     `gorm:"index;not null"`
	Role      string    `gorm:"size:32;not null"` // user | assistant
	Content   string    `gorm:"type:longtext;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime;index"`
}
