package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:50;uniqueIndex" json:"username"`
	Email     string    `gorm:"size:100;uniqueIndex" json:"email"`
	Password  string    `gorm:"size:255" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Images    []Image   `json:"images,omitempty"`
}

type Image struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `json:"userId"`
	OriginalFilename string    `gorm:"size:255" json:"originalFilename"`
	StoredFilename   string    `gorm:"size:255" json:"storedFilename"`
	FilePath         string    `gorm:"size:500" json:"filePath"`
	MimeType         string    `gorm:"size:50" json:"mimeType"`
	FileSize         int64     `json:"fileSize"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Exif             ImageEXIF `json:"exif"`
	Tags             []Tag     `gorm:"many2many:image_tags;" json:"tags"`
	Thumbnail        Thumbnail `json:"thumbnail"`
}

type ImageEXIF struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ImageID       uint      `gorm:"uniqueIndex" json:"imageId"`
	CameraMake    string    `gorm:"size:100" json:"cameraMake"`
	CameraModel   string    `gorm:"size:100" json:"cameraModel"`
	TakenAt       time.Time `json:"takenAt"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	LocationName  string    `gorm:"size:200" json:"locationName"`
	Orientation   int       `json:"orientation"`
	ISO           int       `json:"iso"`
	Aperture      string    `gorm:"size:20" json:"aperture"`
	ShutterSpeed  string    `gorm:"size:20" json:"shutterSpeed"`
	FocalLength   string    `gorm:"size:20" json:"focalLength"`
	Flash         string    `gorm:"size:50" json:"flash"`
	AdditionalRaw string    `gorm:"type:longtext" json:"additionalRaw"`
}

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_tag" json:"userId"`
	Name      string    `gorm:"size:50;uniqueIndex:idx_user_tag" json:"name"`
	Color     string    `gorm:"size:7" json:"color"`
	CreatedAt time.Time `json:"createdAt"`
	Images    []Image   `gorm:"many2many:image_tags;" json:"-"`
}

type ImageTag struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	ImageID uint `gorm:"index" json:"imageId"`
	TagID   uint `gorm:"index" json:"tagId"`
}

type Thumbnail struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ImageID   uint      `gorm:"uniqueIndex" json:"imageId"`
	Data      []byte    `gorm:"type:longblob" json:"-"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}
