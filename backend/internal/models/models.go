// Package models 定义了应用程序的数据模型结构
// 包含用户、图片、标签、EXIF数据等核心实体
package models

import (
	"time"
)

// User 用户模型
// 存储用户的基本信息，包括用户名、邮箱和密码
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`                         // 用户ID，主键
	Username  string    `gorm:"size:50;uniqueIndex" json:"username"`          // 用户名，唯一索引，最大50字符
	Email     string    `gorm:"size:100;uniqueIndex" json:"email"`            // 邮箱，唯一索引，最大100字符
	Password  string    `gorm:"size:255" json:"-"`                            // 密码哈希，JSON序列化时排除
	CreatedAt time.Time `json:"createdAt"`                                    // 创建时间
	UpdatedAt time.Time `json:"updatedAt"`                                    // 更新时间
	Images    []Image   `json:"images,omitempty"`                             // 关联的图片列表，一对多关系
}

// Image 图片模型
// 存储图片的基本信息，包括文件名、路径、尺寸、文件大小等
type Image struct {
	ID               uint      `gorm:"primaryKey" json:"id"`                  // 图片ID，主键
	UserID           uint      `json:"userId"`                                // 所属用户ID
	OriginalFilename string    `gorm:"size:255" json:"originalFilename"`      // 原始文件名，最大255字符
	StoredFilename   string    `gorm:"size:255" json:"storedFilename"`        // 存储文件名（经过处理的唯一文件名）
	FilePath         string    `gorm:"size:500" json:"filePath"`              // 文件存储路径，最大500字符
	MimeType         string    `gorm:"size:50" json:"mimeType"`               // MIME类型，如image/jpeg
	FileSize         int64     `json:"fileSize"`                              // 文件大小（字节）
	Width            int       `json:"width"`                                 // 图片宽度（像素）
	Height           int       `json:"height"`                                // 图片高度（像素）
	CreatedAt        time.Time `json:"createdAt"`                             // 创建时间
	UpdatedAt        time.Time `json:"updatedAt"`                             // 更新时间
	Exif             ImageEXIF `json:"exif"`                                  // 关联的EXIF数据，一对一关系
	Tags             []Tag     `gorm:"many2many:image_tags;" json:"tags"`     // 关联的标签列表，多对多关系
	Thumbnail        Thumbnail `json:"thumbnail"`                             // 关联的缩略图，一对一关系
}

// ImageEXIF 图片EXIF数据模型
// 存储图片的EXIF元数据信息，包括相机信息、拍摄时间、地理位置等
type ImageEXIF struct {
	ID            uint       `gorm:"primaryKey" json:"id"`                    // EXIF记录ID，主键
	ImageID       uint       `gorm:"uniqueIndex" json:"imageId"`              // 关联的图片ID，唯一索引
	CameraMake    string     `gorm:"size:100" json:"cameraMake"`              // 相机制造商，如Canon、Nikon
	CameraModel   string     `gorm:"size:100" json:"cameraModel"`             // 相机型号
	TakenAt       *time.Time `gorm:"type:datetime" json:"takenAt,omitempty"`  // 拍摄时间，使用指针类型以支持NULL值
	Latitude      float64    `json:"latitude"`                                // 纬度（GPS坐标）
	Longitude     float64    `json:"longitude"`                               // 经度（GPS坐标）
	LocationName  string     `gorm:"size:200" json:"locationName"`            // 位置名称
	Orientation   int        `json:"orientation"`                             // 图片方向（旋转角度）
	ISO           int        `json:"iso"`                                     // ISO感光度
	Aperture      string     `gorm:"size:20" json:"aperture"`                 // 光圈值，如f/2.8
	ShutterSpeed  string     `gorm:"size:20" json:"shutterSpeed"`             // 快门速度，如1/125
	FocalLength   string     `gorm:"size:20" json:"focalLength"`              // 焦距，如50mm
	Flash         string     `gorm:"size:50" json:"flash"`                    // 闪光灯设置
	AdditionalRaw string     `gorm:"type:longtext" json:"additionalRaw"`      // 其他原始EXIF数据（JSON格式）
}

// Tag 标签模型
// 用户自定义的标签，用于分类和管理图片
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`                         // 标签ID，主键
	UserID    uint      `gorm:"uniqueIndex:idx_user_tag" json:"userId"`       // 所属用户ID，联合唯一索引的一部分
	Name      string    `gorm:"size:50;uniqueIndex:idx_user_tag" json:"name"` // 标签名称，最大50字符，联合唯一索引的一部分
	Color     string    `gorm:"size:7" json:"color"`                          // 标签颜色（十六进制颜色码，如#FF0000），最大7字符
	CreatedAt time.Time `json:"createdAt"`                                    // 创建时间
	Images    []Image   `gorm:"many2many:image_tags;" json:"-"`               // 关联的图片列表，多对多关系，JSON序列化时排除
}

// ImageTag 图片标签关联表
// 多对多关系的中间表，用于关联图片和标签
type ImageTag struct {
	ID      uint `gorm:"primaryKey" json:"id"`      // 关联记录ID，主键
	ImageID uint `gorm:"index" json:"imageId"`      // 图片ID，建立索引以提高查询性能
	TagID   uint `gorm:"index" json:"tagId"`        // 标签ID，建立索引以提高查询性能
}

// Thumbnail 缩略图模型
// 存储图片的缩略图数据，用于快速预览
type Thumbnail struct {
	ID        uint      `gorm:"primaryKey" json:"id"`           // 缩略图ID，主键
	ImageID   uint      `gorm:"uniqueIndex" json:"imageId"`     // 关联的图片ID，唯一索引（每张图片只有一个缩略图）
	Data      []byte    `gorm:"type:longblob" json:"-"`         // 缩略图二进制数据，使用longblob类型存储，JSON序列化时排除
	Width     int       `json:"width"`                          // 缩略图宽度（像素）
	Height    int       `json:"height"`                         // 缩略图高度（像素）
	Size      int       `json:"size"`                           // 缩略图文件大小（字节）
	CreatedAt time.Time `json:"createdAt"`                      // 创建时间
}
