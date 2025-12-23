// Package services 提供业务逻辑层的服务实现
// image_service.go 实现了图片相关的业务逻辑，包括上传、查询、编辑、删除等功能
package services

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"  // 注册 PNG 解码器，用于 image.DecodeConfig 解析PNG格式
	_ "image/gif"  // 注册 GIF 解码器，用于 image.DecodeConfig 解析GIF格式
	"io"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"image-manager/internal/config"
	"image-manager/internal/dto"
	"image-manager/internal/models"

	"github.com/disintegration/imaging"  // 图片处理库，用于解码、裁剪、生成缩略图等操作
	"github.com/lucasb-eyer/go-colorful" // 颜色处理库，用于颜色空间转换
	"github.com/rwcarlsen/goexif/exif"   // EXIF数据解析库，用于提取图片元数据
	_ "golang.org/x/image/bmp"           // 注册 BMP 解码器，用于 image.DecodeConfig 解析BMP格式
	_ "golang.org/x/image/tiff"          // 注册 TIFF 解码器，用于 image.DecodeConfig 解析TIFF格式
	_ "golang.org/x/image/webp"          // 注册 WebP 解码器，用于 image.DecodeConfig 解析WebP格式
	"gorm.io/gorm"                       // GORM数据库ORM库
	"gorm.io/gorm/clause"                // GORM子句构建器，用于复杂查询
)

// ImageService 图片服务结构体
// 提供图片相关的业务逻辑处理方法
type ImageService struct {
	db   *gorm.DB       // 数据库连接，使用GORM进行数据库操作
	cfg  config.Config  // 应用配置信息，包含存储路径、缩略图尺寸等
	tags *TagService    // 标签服务，用于处理图片标签相关的操作
	ai   *AIService     // AI服务，用于图片分析和自然语言查询转换
}

// NewImageService 创建图片服务实例
// 参数:
//   - db: GORM数据库连接
//   - cfg: 应用配置
//   - tags: 标签服务实例
//   - ai: AI服务实例
// 返回: ImageService指针
func NewImageService(db *gorm.DB, cfg config.Config, tags *TagService, ai *AIService) *ImageService {
	return &ImageService{
		db:   db,
		cfg:  cfg,
		tags: tags,
		ai:   ai,
	}
}

// Upload 上传图片
// 处理图片上传的完整流程：验证文件大小、解析图片格式、保存文件、提取EXIF信息、生成缩略图、关联标签
// 参数:
//   - userID: 上传用户的ID
//   - fileHeader: 上传的文件头信息，包含文件名、大小等
//   - tagNames: 标签名称列表
//   - useAI: 是否使用AI自动生成标签
// 返回: 创建的图片模型指针和错误信息
func (s *ImageService) Upload(userID uint, fileHeader *multipart.FileHeader, tagNames []string, useAI bool) (*models.Image, error) {
	// 检查文件大小是否超过限制
	if fileHeader.Size > s.cfg.MaxUploadSize {
		return nil, errors.New("文件过大")
	}

	// 打开上传的文件
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 将文件内容读取到内存缓冲区，便于后续多次使用（EXIF提取、缩略图生成都需要读取文件）
	buffer := &bytes.Buffer{}
	if _, err := io.Copy(buffer, src); err != nil {
		return nil, err
	}

	// 使用 image.DecodeConfig 解析图片配置信息（仅读取图片头部信息，不加载完整图片到内存）
	// 这样可以快速获取图片格式、宽度、高度等信息，性能优于完整解码
	reader := bytes.NewReader(buffer.Bytes())
	imgCfg, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, errors.New("无法解析图片，支持的格式：JPEG, PNG, GIF, BMP, TIFF, WebP")
	}

	// 将解析到的格式字符串转换为标准MIME类型
	mimeType := getMimeType(format)

	// 生成唯一文件名：使用纳秒时间戳 + 原始文件名（经过清理处理）
	// 纳秒时间戳确保文件名唯一，避免文件名冲突
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(fileHeader.Filename))
	destPath := filepath.Join(s.cfg.StorageDir, "originals", filename)
	// 确保目标目录存在，os.ModePerm 表示目录权限为 0777
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return nil, err
	}

	// 将文件内容写入磁盘，文件权限为 0644（所有者可读写，其他人只读）
	if err := os.WriteFile(destPath, buffer.Bytes(), 0o644); err != nil {
		return nil, err
	}

	// 创建图片记录到数据库
	imageModel := &models.Image{
		UserID:           userID,
		OriginalFilename: fileHeader.Filename,
		StoredFilename:   filename,
		FilePath:         destPath,
		MimeType:         mimeType,
		FileSize:         fileHeader.Size,
		Width:            imgCfg.Width,
		Height:           imgCfg.Height,
	}

	// 使用GORM的Create方法将记录插入数据库
	if err := s.db.Create(imageModel).Error; err != nil {
		return nil, err
	}

	// 异步提取并保存EXIF信息（如果失败只记录日志，不影响主流程）
	// 使用 bytes.NewReader 重新创建reader，因为之前的reader已被读取
	if err := s.extractAndSaveEXIF(imageModel.ID, bytes.NewReader(buffer.Bytes())); err != nil {
		log.Printf("failed to parse EXIF: %v", err)
	}

	// 异步生成缩略图（如果失败只记录日志，不影响主流程）
	if err := s.generateThumbnail(imageModel.ID, bytes.NewReader(buffer.Bytes())); err != nil {
		log.Printf("failed to generate thumbnail: %v", err)
	}

	// 调用AI分析图片并生成标签（如果失败只记录日志，不影响主流程）
	aiTags := []string{}
	if useAI && s.ai != nil {
		// 先获取用户已有的标签库，让AI优先从中选择
		existingTags, err := s.tags.List(userID)
		existingTagNames := []string{}
		if err == nil {
			for _, tag := range existingTags {
				existingTagNames = append(existingTagNames, tag.Name)
			}
		}
		// 调用AI分析图片，传入已有标签库
		log.Printf("开始调用AI分析图片，已有标签库: %v", existingTagNames)
		analyzedTags, err := s.ai.AnalyzeImage(buffer.Bytes(), mimeType, existingTagNames)
		if err != nil {
			log.Printf("AI分析图片失败: %v", err)
		} else {
			log.Printf("AI分析完成，生成标签: %v", analyzedTags)
			aiTags = analyzedTags
		}
	} else if !useAI {
		log.Printf("用户选择不使用AI自动生成标签")
	}

	// 合并用户提供的标签和AI生成的标签，去重
	allTags := append(tagNames, aiTags...)
	log.Printf("合并后的所有标签（用户标签: %v, AI标签: %v）: %v", tagNames, aiTags, allTags)
	uniqueTags := make(map[string]bool)
	finalTags := []string{}
	for _, tag := range allTags {
		tag = strings.TrimSpace(tag)
		if tag != "" && !uniqueTags[tag] {
			uniqueTags[tag] = true
			finalTags = append(finalTags, tag)
		}
	}
	log.Printf("去重后的最终标签列表: %v (共%d个)", finalTags, len(finalTags))

	// 如果有关键词，则关联标签到图片
	if len(finalTags) > 0 {
		if err := s.tags.AssignByNames(userID, imageModel.ID, finalTags); err != nil {
			log.Printf("failed to assign tags: %v", err)
		}
	}

	return imageModel, nil
}

// extractAndSaveEXIF 提取并保存图片的EXIF信息
// 从图片文件中提取EXIF元数据（拍摄时间、相机信息等）并保存到数据库
// 参数:
//   - imageID: 图片ID
//   - reader: 图片文件的读取器
// 返回: 错误信息
func (s *ImageService) extractAndSaveEXIF(imageID uint, reader io.Reader) error {
	// 使用 goexif 库解析EXIF数据
	// exif.Decode 会从图片文件的EXIF段中读取所有元数据
	exifData, err := exif.Decode(reader)
	if err != nil {
		return err
	}

	exifModel := models.ImageEXIF{
		ImageID: imageID,
	}

	// 提取拍摄时间（DateTimeOriginal标签）
	// EXIF标准中使用 "2006:01:02 15:04:05" 格式存储时间
	if tag, err := exifData.Get(exif.DateTimeOriginal); err == nil {
		if ts, err := tag.StringVal(); err == nil {
			// 解析时间字符串为time.Time类型
			// 使用指针类型以便支持NULL值（如果图片没有拍摄时间）
			if parsed, parseErr := time.Parse("2006:01:02 15:04:05", ts); parseErr == nil {
				exifModel.TakenAt = &parsed
			}
		}
	}

	// 提取相机型号（Model标签）
	if tag, err := exifData.Get(exif.Model); err == nil {
		if model, err := tag.StringVal(); err == nil && model != "" {
			exifModel.CameraModel = model
		}
	}
	// 提取相机制造商（Make标签）
	if tag, err := exifData.Get(exif.Make); err == nil {
		if make, err := tag.StringVal(); err == nil && make != "" {
			exifModel.CameraMake = make
		}
	}

	// 构建更新字段映射，用于处理数据库冲突（如果EXIF记录已存在则更新）
	// 只在有值时更新 TakenAt，避免用nil覆盖已有的有效时间
	updates := map[string]interface{}{
		"camera_make":  exifModel.CameraMake,
		"camera_model": exifModel.CameraModel,
	}
	if exifModel.TakenAt != nil {
		updates["taken_at"] = exifModel.TakenAt
	}

	// 使用GORM的OnConflict子句处理冲突：如果image_id已存在则更新，否则插入
	// clause.OnConflict 实现 UPSERT（INSERT ... ON DUPLICATE KEY UPDATE）语义
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "image_id"}},  // 冲突检测的列
		DoUpdates: clause.Assignments(updates),          // 发生冲突时执行的更新操作
	}).Create(&exifModel).Error
}

// generateThumbnail 生成图片缩略图
// 使用imaging库将原图缩放并裁剪到指定尺寸，然后保存到数据库
// 参数:
//   - imageID: 图片ID
//   - reader: 图片文件的读取器
// 返回: 错误信息
func (s *ImageService) generateThumbnail(imageID uint, reader io.Reader) error {
	// 使用imaging库解码图片（支持多种格式：JPEG、PNG、GIF等）
	// imaging.Decode 会将图片完整加载到内存中
	img, err := imaging.Decode(reader)
	if err != nil {
		return err
	}

	// 使用imaging.Fill方法生成缩略图
	// Fill会按比例缩放图片，然后裁剪到指定尺寸，保持图片中心部分
	// imaging.Center: 裁剪时保持中心对齐
	// imaging.Lanczos: 使用Lanczos重采样算法，提供较好的缩放质量
	thumb := imaging.Fill(img, s.cfg.ThumbnailWidth, s.cfg.ThumbnailHeight, imaging.Center, imaging.Lanczos)

	// 将缩略图编码为JPEG格式并写入缓冲区
	// Quality: 85 是JPEG压缩质量，平衡文件大小和图片质量
	buff := &bytes.Buffer{}
	if err := jpeg.Encode(buff, thumb, &jpeg.Options{Quality: 85}); err != nil {
		return err
	}

	// 创建缩略图记录
	thumbnail := models.Thumbnail{
		ImageID: imageID,
		Data:    buff.Bytes(),               // 缩略图二进制数据
		Width:   s.cfg.ThumbnailWidth,       // 缩略图宽度（配置中定义）
		Height:  s.cfg.ThumbnailHeight,      // 缩略图高度（配置中定义）
		Size:    buff.Len(),                 // 缩略图文件大小（字节）
	}

	// 使用OnConflict处理冲突：如果缩略图已存在则更新所有字段
	return s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&thumbnail).Error
}

func (s *ImageService) List(userID uint, filters map[string]string, page, pageSize int) ([]models.Image, int64, error) {
	var images []models.Image
	var total int64

	baseQuery := s.db.Model(&models.Image{}).Where("images.user_id = ?", userID)
	
	// 检查是否有keyword
	hasKeyword := false
	keyword := ""
	if k, ok := filters["keyword"]; ok && k != "" {
		hasKeyword = true
		keyword = k
	}
	
	// 检查是否有其他筛选条件（除了keyword）
	hasOtherFilters := false
	hasOtherFilters = hasOtherFilters || (filters["start"] != "" || filters["end"] != "")
	hasOtherFilters = hasOtherFilters || (filters["width_min"] != "" || filters["width_max"] != "")
	hasOtherFilters = hasOtherFilters || (filters["height_min"] != "" || filters["height_max"] != "")
	hasOtherFilters = hasOtherFilters || (filters["size_min"] != "" || filters["size_max"] != "")
	hasOtherFilters = hasOtherFilters || (filters["taken_start"] != "" || filters["taken_end"] != "")
	hasOtherFilters = hasOtherFilters || (filters["tags"] != "")
	
	// 获取keyword_mode，默认为"or"
	keywordMode := filters["keyword_mode"]
	if keywordMode != "and" && keywordMode != "or" {
		keywordMode = "or"  // 默认使用OR模式
	}
	
	// 构建查询：根据keyword_mode决定keyword和其他条件的关系
	// 正确的逻辑：
	// - keyword OR/AND (其他所有条件的组合)
	// - 标签内部根据tag_mode使用OR/AND，但标签作为整体与其他条件用AND连接
	// 重要：空的筛选条件被视为"不限制"（匹配所有），而不是"不匹配"
	var query *gorm.DB
	
	// 如果两者都有，根据keyword_mode决定连接方式
	if hasKeyword && hasOtherFilters {
		if keywordMode == "and" {
			// AND模式：关键词 AND (其他所有条件的组合)
			// 检查是否包含标签筛选（标签筛选会使用GROUP BY，可能影响Preload）
			hasTagFilter := filters["tags"] != "" && strings.TrimSpace(filters["tags"]) != ""
			if hasTagFilter {
				// 如果包含标签筛选，先获取符合条件的图片ID列表，然后使用ID列表进行最终查询
				// 这样可以避免GROUP BY对Preload的影响
				tempQuery := baseQuery.Where("images.original_filename LIKE ?", "%"+keyword+"%")
				tempQuery = s.buildOtherFiltersQuery(tempQuery, userID, filters)
				var imageIDs []uint
				if err := tempQuery.Pluck("images.id", &imageIDs).Error; err != nil {
					return nil, 0, err
				}
				if len(imageIDs) == 0 {
					return []models.Image{}, 0, nil
				}
				// 使用ID列表创建干净的查询，避免GROUP BY等子句影响Preload
				query = s.db.Model(&models.Image{}).Where("images.user_id = ? AND images.id IN ?", userID, imageIDs)
			} else {
				// 没有标签筛选，可以直接使用buildOtherFiltersQuery的结果
				query = baseQuery.Where("images.original_filename LIKE ?", "%"+keyword+"%")
				query = s.buildOtherFiltersQuery(query, userID, filters)
			}
		} else {
			// OR模式：关键词 OR (其他所有条件的组合)
			// 使用子查询或分别查询然后合并ID的方式
			// 构建keyword查询（只包含keyword条件）
			keywordQuery := s.db.Model(&models.Image{}).
				Where("images.user_id = ?", userID).
				Where("images.original_filename LIKE ?", "%"+keyword+"%")
			
			// 构建其他条件查询（作为整体，不包含keyword）
			otherQuery := s.buildOtherFiltersQuery(
				s.db.Model(&models.Image{}).Where("images.user_id = ?", userID),
				userID,
				filters,
			)
			
			// 获取keyword查询的图片ID
			var keywordImageIDs []uint
			if err := keywordQuery.Pluck("images.id", &keywordImageIDs).Error; err != nil {
				return nil, 0, err
			}
			
			// 获取其他条件查询的图片ID
			var otherImageIDs []uint
			if err := otherQuery.Pluck("images.id", &otherImageIDs).Error; err != nil {
				return nil, 0, err
			}
			
			// 合并去重
			allImageIDs := make(map[uint]bool)
			for _, id := range keywordImageIDs {
				allImageIDs[id] = true
			}
			for _, id := range otherImageIDs {
				allImageIDs[id] = true
			}
			
			// 转换为slice
			finalImageIDs := make([]uint, 0, len(allImageIDs))
			for id := range allImageIDs {
				finalImageIDs = append(finalImageIDs, id)
			}
			
			if len(finalImageIDs) == 0 {
				// 没有匹配的结果
				return []models.Image{}, 0, nil
			}
			
			// 最终查询：只根据合并后的ID列表查询，不包含任何WHERE条件（除了user_id和id IN）
			// 创建一个全新的查询，避免之前查询中的JOIN、GROUP BY等子句影响Preload
			query = s.db.Model(&models.Image{}).Where("images.user_id = ? AND images.id IN ?", userID, finalImageIDs)
		}
	} else if hasKeyword {
		// 只有keyword，没有其他条件
		// 无论keyword_mode是什么，都只查询keyword匹配的（因为其他条件为空，视为true，但单独的关键词查询应该只返回匹配的）
		query = baseQuery.Where("images.original_filename LIKE ?", "%"+keyword+"%")
	} else if hasOtherFilters {
		// 只有其他条件，没有keyword
		// 检查是否包含标签筛选（标签筛选会使用GROUP BY，可能影响Preload）
		hasTagFilter := filters["tags"] != "" && strings.TrimSpace(filters["tags"]) != ""
		if hasTagFilter {
			// 如果包含标签筛选，先获取图片ID列表，然后使用ID列表进行最终查询
			// 这样可以避免GROUP BY对Preload的影响
			otherQuery := s.buildOtherFiltersQuery(baseQuery, userID, filters)
			var otherImageIDs []uint
			if err := otherQuery.Pluck("images.id", &otherImageIDs).Error; err != nil {
				return nil, 0, err
			}
			if len(otherImageIDs) == 0 {
				return []models.Image{}, 0, nil
			}
			// 使用ID列表创建干净的查询，避免GROUP BY等子句影响Preload
			query = s.db.Model(&models.Image{}).Where("images.user_id = ? AND images.id IN ?", userID, otherImageIDs)
		} else {
			// 没有标签筛选，可以直接使用buildOtherFiltersQuery的结果
			query = s.buildOtherFiltersQuery(baseQuery, userID, filters)
		}
	} else {
		// 没有任何筛选条件，返回所有图片
		query = baseQuery
	}
	
	// 添加Preload
	query = query.Preload("Thumbnail").Preload("Exif").Preload("Tags")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("images.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&images).Error; err != nil {
		return nil, 0, err
	}

	return images, total, nil
}

// parseTagString 解析标签字符串，支持中英文逗号分隔
// 参数:
//   - tagStr: 标签字符串，可以用中文逗号（，）或英文逗号（,）分隔
// 返回: 标签名称列表（已去重、去除空格）
func parseTagString(tagStr string) []string {
	// 先统一将中文逗号替换为英文逗号，然后按英文逗号分割
	// 这样可以在一个字符串中同时使用中英文逗号
	normalized := strings.ReplaceAll(tagStr, "，", ",")
	
	tagNames := []string{}
	for _, t := range strings.Split(normalized, ",") {
		name := strings.TrimSpace(t)
		if name != "" {
			tagNames = append(tagNames, name)
		}
	}
	
	return tagNames
}

// buildOtherFiltersQuery 构建除keyword外的其他筛选条件查询
func (s *ImageService) buildOtherFiltersQuery(baseQuery *gorm.DB, userID uint, filters map[string]string) *gorm.DB {
	query := baseQuery
	
	if start, ok := filters["start"]; ok && start != "" {
		query = query.Where("images.created_at >= ?", start)
	}
	if end, ok := filters["end"]; ok && end != "" {
		query = query.Where("images.created_at <= ?", end)
	}

	// 分辨率筛选
	if wMinStr, ok := filters["width_min"]; ok && wMinStr != "" {
		if wMin, err := strconv.Atoi(wMinStr); err == nil {
			query = query.Where("images.width >= ?", wMin)
		}
	}
	if wMaxStr, ok := filters["width_max"]; ok && wMaxStr != "" {
		if wMax, err := strconv.Atoi(wMaxStr); err == nil {
			query = query.Where("images.width <= ?", wMax)
		}
	}
	if hMinStr, ok := filters["height_min"]; ok && hMinStr != "" {
		if hMin, err := strconv.Atoi(hMinStr); err == nil {
			query = query.Where("images.height >= ?", hMin)
		}
	}
	if hMaxStr, ok := filters["height_max"]; ok && hMaxStr != "" {
		if hMax, err := strconv.Atoi(hMaxStr); err == nil {
			query = query.Where("images.height <= ?", hMax)
		}
	}

	// 文件大小筛选（单位：MB，支持小数）
	// 前端传入MB值（可以是小数，如1.5），后端将其转换为字节进行比较
	if sizeMinStr, ok := filters["size_min"]; ok && sizeMinStr != "" {
		if sizeMinMB, err := strconv.ParseFloat(sizeMinStr, 64); err == nil && sizeMinMB > 0 {
			// 将MB转换为字节：MB * 1024 * 1024，然后转换为int64（向下取整）
			sizeMinBytes := int64(sizeMinMB * 1024 * 1024)
			query = query.Where("images.file_size >= ?", sizeMinBytes)
		}
	}
	if sizeMaxStr, ok := filters["size_max"]; ok && sizeMaxStr != "" {
		if sizeMaxMB, err := strconv.ParseFloat(sizeMaxStr, 64); err == nil && sizeMaxMB > 0 {
			// 将MB转换为字节：MB * 1024 * 1024，然后转换为int64（向下取整）
			sizeMaxBytes := int64(sizeMaxMB * 1024 * 1024)
			query = query.Where("images.file_size <= ?", sizeMaxBytes)
		}
	}

	// 拍摄时间（EXIF）
	hasTakenFilter := false
	takenStart, hasStart := filters["taken_start"]
	takenEnd, hasEnd := filters["taken_end"]
	if hasStart && takenStart != "" {
		hasTakenFilter = true
	}
	if hasEnd && takenEnd != "" {
		hasTakenFilter = true
	}
	
	if hasTakenFilter {
		query = query.Joins("LEFT JOIN image_exifs ON images.id = image_exifs.image_id")
		if hasStart && takenStart != "" {
			query = query.Where("image_exifs.taken_at >= ?", takenStart)
		}
		if hasEnd && takenEnd != "" {
			query = query.Where("image_exifs.taken_at <= ?", takenEnd)
		}
	}

	// 标签筛选，支持多个标签（用逗号分隔，支持中英文逗号）
	// 获取tag_mode，默认为"or"
	tagMode := filters["tag_mode"]
	if tagMode != "and" && tagMode != "or" {
		tagMode = "or"  // 默认使用OR模式
	}
	
	// 同时只使用标签库里确实存在的标签
	if tagStr, ok := filters["tags"]; ok && strings.TrimSpace(tagStr) != "" {
		tagNames := parseTagString(tagStr)
		// 过滤：只保留标签库里确实存在的标签
		if len(tagNames) > 0 {
			var existingTags []models.Tag
			s.db.Where("user_id = ? AND name IN ?", userID, tagNames).Find(&existingTags)
			validTagNames := []string{}
			for _, tag := range existingTags {
				validTagNames = append(validTagNames, tag.Name)
			}
			// 如果用户提供了标签筛选条件，但所有标签都不存在于标签库，应该返回空结果（false条件）
			if len(validTagNames) == 0 {
				// 使用一个不可能满足的条件，确保返回空结果
				query = query.Where("1 = 0")
				return query
			}
			// 存在有效标签时进行筛选
			if tagMode == "and" {
				// AND模式：图片必须同时拥有所有标签
				// 对每个标签进行JOIN，确保图片包含所有标签
				for i, tagName := range validTagNames {
					alias := fmt.Sprintf("image_tags_%d", i)
					tagAlias := fmt.Sprintf("tags_%d", i)
					query = query.
						Joins(fmt.Sprintf("JOIN image_tags AS %s ON images.id = %s.image_id", alias, alias)).
						Joins(fmt.Sprintf("JOIN tags AS %s ON %s.id = %s.tag_id AND %s.user_id = ? AND %s.name = ?", 
							tagAlias, tagAlias, alias, tagAlias, tagAlias), userID, tagName)
				}
				query = query.Group("images.id").Distinct("images.id")
			} else {
				// OR模式：图片有其中任意一个标签即可匹配
				query = query.
					Joins("JOIN image_tags ON images.id = image_tags.image_id").
					Joins("JOIN tags ON tags.id = image_tags.tag_id AND tags.user_id = ?", userID).
					Where("tags.name IN ?", validTagNames).
					Group("images.id").
					Distinct("images.id")
			}
		}
	}
	
	return query
}

func (s *ImageService) Get(userID, imageID uint) (*models.Image, error) {
	var imageModel models.Image
	if err := s.db.Preload("Thumbnail").Preload("Exif").Preload("Tags").Where("user_id = ? AND id = ?", userID, imageID).First(&imageModel).Error; err != nil {
		return nil, err
	}
	return &imageModel, nil
}

func (s *ImageService) Update(userID, imageID uint, fileHeader *multipart.FileHeader) (*models.Image, error) {
	imageModel, err := s.Get(userID, imageID)
	if err != nil {
		return nil, err
	}

	if fileHeader.Size > s.cfg.MaxUploadSize {
		return nil, errors.New("文件过大")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	buffer := &bytes.Buffer{}
	if _, err := io.Copy(buffer, src); err != nil {
		return nil, err
	}

	// 先使用 image.DecodeConfig 获取格式和尺寸（需要导入相应的解码器）
	reader := bytes.NewReader(buffer.Bytes())
	imgCfg, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, errors.New("无法解析图片，支持的格式：JPEG, PNG, GIF, BMP, TIFF, WebP")
	}
	
	// 标准化 MIME 类型
	mimeType := getMimeType(format)

	// 删除旧文件
	if err := os.Remove(imageModel.FilePath); err != nil && !os.IsNotExist(err) {
		log.Printf("failed to remove old file: %v", err)
	}

	// 保存新文件
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(fileHeader.Filename))
	destPath := filepath.Join(s.cfg.StorageDir, "originals", filename)
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return nil, err
	}

	if err := os.WriteFile(destPath, buffer.Bytes(), 0o644); err != nil {
		return nil, err
	}
	
	// 更新数据库记录
	imageModel.StoredFilename = filename
	imageModel.FilePath = destPath
	imageModel.MimeType = mimeType
	imageModel.FileSize = fileHeader.Size
	imageModel.Width = imgCfg.Width
	imageModel.Height = imgCfg.Height

	if err := s.db.Save(imageModel).Error; err != nil {
		return nil, err
	}

	// 更新缩略图
	if err := s.generateThumbnail(imageModel.ID, bytes.NewReader(buffer.Bytes())); err != nil {
		log.Printf("failed to generate thumbnail: %v", err)
	}

	// 更新 EXIF
	if err := s.extractAndSaveEXIF(imageModel.ID, bytes.NewReader(buffer.Bytes())); err != nil {
		log.Printf("failed to parse EXIF: %v", err)
	}

	return imageModel, nil
}

func (s *ImageService) Delete(userID, imageID uint) error {
	imageModel, err := s.Get(userID, imageID)
	if err != nil {
		return err
	}

	if err := os.Remove(imageModel.FilePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Thumbnail{}, "image_id = ?", imageID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.ImageEXIF{}, "image_id = ?", imageID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.ImageTag{}, "image_id = ?", imageID).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Image{}, "id = ?", imageID).Error
	})
}

func (s *ImageService) GetThumbnail(imageID uint) (*models.Thumbnail, error) {
	var thumb models.Thumbnail
	if err := s.db.Where("image_id = ?", imageID).First(&thumb).Error; err != nil {
		return nil, err
	}
	return &thumb, nil
}

func (s *ImageService) GetFile(imageID uint) (*models.Image, []byte, error) {
	imageModel, err := s.GetRaw(imageID)
	if err != nil {
		return nil, nil, err
	}

	data, err := os.ReadFile(imageModel.FilePath)
	if err != nil {
		return nil, nil, err
	}

	return imageModel, data, nil
}

func (s *ImageService) GetRaw(imageID uint) (*models.Image, error) {
	var img models.Image
	if err := s.db.Where("id = ?", imageID).First(&img).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	lower := strings.ToLower(base)
	clean := strings.ReplaceAll(lower, " ", "_")
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			return r
		}
		return '-'
	}, clean)
}

func (s *ImageService) Crop(userID, imageID uint, req dto.CropRequest) (*models.Image, error) {
	imageModel, err := s.Get(userID, imageID)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(imageModel.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := imaging.Decode(file)
	if err != nil {
		return nil, err
	}

	cropped := imaging.Crop(img, image.Rect(req.X, req.Y, req.X+req.Width, req.Y+req.Height))
	filename := fmt.Sprintf("%d_crop_%s", time.Now().UnixNano(), imageModel.StoredFilename)
	destPath := filepath.Join(s.cfg.StorageDir, "originals", filename)

	if err := imaging.Save(cropped, destPath); err != nil {
		return nil, err
	}

	info, err := os.Stat(destPath)
	if err != nil {
		return nil, err
	}

	newImage := models.Image{
		UserID:           userID,
		OriginalFilename: "crop_" + imageModel.OriginalFilename,
		StoredFilename:   filename,
		FilePath:         destPath,
		MimeType:         imageModel.MimeType,
		FileSize:         info.Size(),
		Width:            cropped.Bounds().Dx(),
		Height:           cropped.Bounds().Dy(),
	}

	if err := s.db.Create(&newImage).Error; err != nil {
		return nil, err
	}

	return &newImage, nil
}

func (s *ImageService) Adjust(userID, imageID uint, req dto.AdjustRequest) (*models.Image, error) {
	imageModel, err := s.Get(userID, imageID)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(imageModel.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := imaging.Decode(file)
	if err != nil {
		return nil, err
	}

	var adjusted image.Image = imaging.AdjustBrightness(img, float64(req.Brightness)/100)
	adjusted = imaging.AdjustContrast(adjusted, float64(req.Contrast)/100)
	adjusted = imaging.AdjustSaturation(adjusted, float64(req.Saturation)/100)
	adjusted = adjustHue(adjusted, float64(req.Hue))

	filename := fmt.Sprintf("%d_adjust_%s", time.Now().UnixNano(), imageModel.StoredFilename)
	destPath := filepath.Join(s.cfg.StorageDir, "originals", filename)

	if err := imaging.Save(adjusted, destPath); err != nil {
		return nil, err
	}

	info, err := os.Stat(destPath)
	if err != nil {
		return nil, err
	}

	newImage := models.Image{
		UserID:           userID,
		OriginalFilename: "adjust_" + imageModel.OriginalFilename,
		StoredFilename:   filename,
		FilePath:         destPath,
		MimeType:         imageModel.MimeType,
		FileSize:         info.Size(),
		Width:            adjusted.Bounds().Dx(),
		Height:           adjusted.Bounds().Dy(),
	}

	if err := s.db.Create(&newImage).Error; err != nil {
		return nil, err
	}

	return &newImage, nil
}

func adjustHue(img image.Image, degrees float64) image.Image {
	if degrees == 0 {
		return img
	}

	bounds := img.Bounds()
	dst := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			colorVal, ok := colorful.MakeColor(img.At(x, y))
			if !ok {
				continue
			}
			h, s, l := colorVal.Hsl()
			h = math.Mod(h+degrees, 360)
			if h < 0 {
				h += 360
			}
			newColor := colorful.Hsl(h, s, l)
			_, _, _, alpha := img.At(x, y).RGBA()
			dst.Set(x, y, color.NRGBA{
				R: uint8(newColor.R * 255),
				G: uint8(newColor.G * 255),
				B: uint8(newColor.B * 255),
				A: uint8(alpha >> 8),
			})
		}
	}

	return dst
}

// getMimeType 将 imaging 格式转换为标准 MIME 类型
func getMimeType(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "bmp":
		return "image/bmp"
	case "tiff", "tif":
		return "image/tiff"
	case "webp":
		return "image/webp"
	default:
		return fmt.Sprintf("image/%s", format)
	}
}

// GetOtherUserImages 获取其他用户的所有图片（用于导入功能）
// 参数:
//   - sourceUserID: 源用户ID
// 返回: 图片列表和错误信息
func (s *ImageService) GetOtherUserImages(sourceUserID uint) ([]models.Image, error) {
	var images []models.Image
	if err := s.db.Preload("Thumbnail").Preload("Tags").Where("user_id = ?", sourceUserID).Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

// ImportImages 导入图片
// 从源用户复制图片到目标用户，同时复制标签、EXIF和缩略图
// 参数:
//   - targetUserID: 目标用户ID（当前用户）
//   - sourceUserID: 源用户ID（被导入的用户）
//   - imageIDs: 要导入的图片ID列表
//   - tagService: 标签服务，用于创建和关联标签
// 返回: 导入的图片列表和错误信息
func (s *ImageService) ImportImages(targetUserID, sourceUserID uint, imageIDs []uint, tagService *TagService) ([]models.Image, error) {
	// 1. 获取源用户的所有标签，以便在导入时保留标签颜色
	var sourceTags []models.Tag
	if err := s.db.Where("user_id = ?", sourceUserID).Find(&sourceTags).Error; err != nil {
		return nil, fmt.Errorf("获取源用户标签失败: %v", err)
	}
	sourceTagMap := make(map[string]models.Tag)
	for _, tag := range sourceTags {
		sourceTagMap[tag.Name] = tag
	}

	// 2. 获取目标用户现有的标签，找出需要创建的标签
	var targetTags []models.Tag
	if err := s.db.Where("user_id = ?", targetUserID).Find(&targetTags).Error; err != nil {
		return nil, fmt.Errorf("获取目标用户标签失败: %v", err)
	}
	targetTagMap := make(map[string]models.Tag)
	for _, tag := range targetTags {
		targetTagMap[tag.Name] = tag
	}

	// 3. 获取要导入的图片（包含完整信息：Tags, Exif, Thumbnail）
	var sourceImages []models.Image
	if err := s.db.Preload("Tags").Preload("Exif").Preload("Thumbnail").
		Where("user_id = ? AND id IN ?", sourceUserID, imageIDs).Find(&sourceImages).Error; err != nil {
		return nil, fmt.Errorf("获取源图片失败: %v", err)
	}

	if len(sourceImages) == 0 {
		return []models.Image{}, nil
	}

	// 4. 创建目标用户缺失的标签（保留源用户的标签颜色）
	for _, sourceImg := range sourceImages {
		for _, sourceTag := range sourceImg.Tags {
			if _, exists := targetTagMap[sourceTag.Name]; !exists {
				// 标签不存在，需要创建，使用源用户的标签颜色
				newTag := models.Tag{
					UserID: targetUserID,
					Name:   sourceTag.Name,
					Color:  sourceTag.Color, // 保留源用户的标签颜色
				}
				if err := s.db.Create(&newTag).Error; err != nil {
					log.Printf("创建标签失败 %s: %v", sourceTag.Name, err)
					continue
				}
				targetTagMap[sourceTag.Name] = newTag
			}
		}
	}

	// 5. 导入每张图片
	importedImages := []models.Image{}
	for _, sourceImg := range sourceImages {
		// 读取源图片文件
		fileData, err := os.ReadFile(sourceImg.FilePath)
		if err != nil {
			log.Printf("读取源图片文件失败 %s: %v", sourceImg.FilePath, err)
			continue
		}

		// 生成新的文件名
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(sourceImg.OriginalFilename))
		destPath := filepath.Join(s.cfg.StorageDir, "originals", filename)
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			log.Printf("创建目录失败: %v", err)
			continue
		}

		// 复制文件
		if err := os.WriteFile(destPath, fileData, 0o644); err != nil {
			log.Printf("复制文件失败: %v", err)
			continue
		}

		// 创建新的图片记录（除了CreatedAt和UpdatedAt，其他信息原样保留）
		newImage := models.Image{
			UserID:           targetUserID,
			OriginalFilename: sourceImg.OriginalFilename,
			StoredFilename:   filename,
			FilePath:         destPath,
			MimeType:         sourceImg.MimeType,
			FileSize:         sourceImg.FileSize,
			Width:            sourceImg.Width,
			Height:           sourceImg.Height,
			// CreatedAt 和 UpdatedAt 会自动设置为当前时间
		}

		// 保存图片记录
		if err := s.db.Create(&newImage).Error; err != nil {
			log.Printf("创建图片记录失败: %v", err)
			os.Remove(destPath) // 清理已复制的文件
			continue
		}

		// 复制EXIF数据（如果存在）
		if sourceImg.Exif.ID != 0 {
			newExif := sourceImg.Exif
			newExif.ID = 0 // 重置ID，让数据库自动生成
			newExif.ImageID = newImage.ID
			if err := s.db.Create(&newExif).Error; err != nil {
				log.Printf("复制EXIF数据失败: %v", err)
			}
		}

		// 复制缩略图（如果存在）
		if sourceImg.Thumbnail.ID != 0 {
			newThumbnail := sourceImg.Thumbnail
			newThumbnail.ID = 0 // 重置ID
			newThumbnail.ImageID = newImage.ID
			if err := s.db.Create(&newThumbnail).Error; err != nil {
				log.Printf("复制缩略图失败: %v", err)
			}
		}

		// 关联标签（使用目标用户的标签，如果标签不存在则已经创建）
		tagNames := []string{}
		for _, sourceTag := range sourceImg.Tags {
			tagNames = append(tagNames, sourceTag.Name)
		}
		if len(tagNames) > 0 {
			if err := tagService.AssignByNames(targetUserID, newImage.ID, tagNames); err != nil {
				log.Printf("关联标签失败: %v", err)
			}
		}

		importedImages = append(importedImages, newImage)
	}

	return importedImages, nil
}
