package services

import (
	"errors"

	"image-manager/internal/dto"
	"image-manager/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TagService struct {
	db *gorm.DB
}

func NewTagService(db *gorm.DB) *TagService {
	return &TagService{db: db}
}

func (s *TagService) Create(userID uint, req dto.CreateTagRequest) (*models.Tag, error) {
	tag := models.Tag{
		UserID: userID,
		Name:   req.Name,
		Color:  req.Color,
	}
	if err := s.db.Create(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (s *TagService) List(userID uint) ([]models.Tag, error) {
	var tags []models.Tag
	if err := s.db.Where("user_id = ?", userID).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (s *TagService) Assign(imageID, tagID uint, userID uint) error {
	var tag models.Tag
	if err := s.db.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		return err
	}

	association := models.ImageTag{
		ImageID: imageID,
		TagID:   tagID,
	}

	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&association).Error
}

func (s *TagService) AssignByNames(userID, imageID uint, names []string) error {
	if len(names) == 0 {
		return nil
	}

	// 先对标签名去重
	uniqueNames := make(map[string]bool)
	deduplicatedNames := []string{}
	for _, name := range names {
		if name != "" && !uniqueNames[name] {
			uniqueNames[name] = true
			deduplicatedNames = append(deduplicatedNames, name)
		}
	}

	for _, name := range deduplicatedNames {
		var tag models.Tag
		// 先查找是否存在该标签
		err := s.db.Where("user_id = ? AND name = ?", userID, name).First(&tag).Error
		if err != nil {
			// 如果不存在，创建新标签，颜色为空（无色）
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tag = models.Tag{
			UserID: userID,
			Name:   name,
					Color:  "", // 自动创建的标签颜色为空
				}
				if err := s.db.Create(&tag).Error; err != nil {
					return err
		}
			} else {
			return err
			}
		}
		// 如果标签已存在，使用现有的标签（包括其颜色）
		if err := s.Assign(imageID, tag.ID, userID); err != nil {
			return err
		}
	}

	// 操作后清理重复的标签关联（确保每个标签只关联一次）
	return s.deduplicateImageTags(imageID)
}

func (s *TagService) AssignBulk(userID, imageID uint, tagIDs []uint) error {
	if len(tagIDs) == 0 {
		return errors.New("标签不能为空")
	}
	for _, tagID := range tagIDs {
		if err := s.Assign(imageID, tagID, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *TagService) Remove(imageID, tagID, userID uint) error {
	return s.db.Where("image_id = ? AND tag_id = ?", imageID, tagID).Delete(&models.ImageTag{}).Error
}

// Delete 删除标签
// 删除标签时，同时删除所有图片与该标签的关联（ImageTag）
// 参数:
//   - userID: 用户ID，确保只能删除自己的标签
//   - tagID: 要删除的标签ID
// 返回: 错误信息
func (s *TagService) Delete(userID, tagID uint) error {
	// 先验证标签是否存在且属于该用户
	var tag models.Tag
	if err := s.db.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		return err
	}

	// 删除该标签与所有图片的关联（ImageTag）
	if err := s.db.Where("tag_id = ?", tagID).Delete(&models.ImageTag{}).Error; err != nil {
		return err
	}

	// 删除标签本身
	if err := s.db.Delete(&tag).Error; err != nil {
		return err
	}

	return nil
}

func (s *TagService) UpdateColor(userID, tagID uint, color string) (*models.Tag, error) {
	var tag models.Tag
	if err := s.db.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		return nil, err
	}
	
	tag.Color = color
	if err := s.db.Save(&tag).Error; err != nil {
		return nil, err
	}
	
	return &tag, nil
}

// UpdateImageTag 修改图片的标签（将旧标签替换为新标签）
// 如果新标签不存在，使用旧标签的颜色创建新标签
// 如果新标签已存在，直接使用该标签
func (s *TagService) UpdateImageTag(userID, imageID uint, oldTagID uint, newTagName string) error {
	// 获取旧标签信息（包括颜色）
	var oldTag models.Tag
	if err := s.db.Where("id = ? AND user_id = ?", oldTagID, userID).First(&oldTag).Error; err != nil {
		return err
	}

	// 查找新标签是否存在
	var newTag models.Tag
	err := s.db.Where("user_id = ? AND name = ?", userID, newTagName).First(&newTag).Error
	if err != nil {
		// 如果新标签不存在，使用旧标签的颜色创建新标签
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newTag = models.Tag{
				UserID: userID,
				Name:   newTagName,
				Color:  oldTag.Color, // 使用旧标签的颜色
			}
			if err := s.db.Create(&newTag).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 删除旧标签关联
	if err := s.db.Where("image_id = ? AND tag_id = ?", imageID, oldTagID).Delete(&models.ImageTag{}).Error; err != nil {
		return err
	}

	// 添加新标签关联
	association := models.ImageTag{
		ImageID: imageID,
		TagID:   newTag.ID,
	}
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&association).Error; err != nil {
		return err
	}

	// 操作后清理重复的标签关联（确保每个标签只关联一次）
	return s.deduplicateImageTags(imageID)
}

// AddImageTagByName 通过标签名给图片添加标签
// 如果标签不存在，创建新标签，颜色为空
func (s *TagService) AddImageTagByName(userID, imageID uint, tagName string) error {
	// 检查该图片是否已经有这个标签
	var existingAssociations []models.ImageTag
	if err := s.db.Where("image_id = ?", imageID).Find(&existingAssociations).Error; err != nil {
		return err
	}

	// 查找该标签是否存在
	var tag models.Tag
	err := s.db.Where("user_id = ? AND name = ?", userID, tagName).First(&tag).Error
	if err != nil {
		// 如果标签不存在，创建新标签，颜色为空
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tag = models.Tag{
				UserID: userID,
				Name:   tagName,
				Color:  "", // 自动创建的标签颜色为空
			}
			if err := s.db.Create(&tag).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 检查该图片是否已经有这个标签（通过tagID检查）
	for _, assoc := range existingAssociations {
		if assoc.TagID == tag.ID {
			// 如果已存在，直接返回（不需要重复添加）
			return nil
		}
	}

	// 添加标签关联
	association := models.ImageTag{
		ImageID: imageID,
		TagID:   tag.ID,
	}
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&association).Error; err != nil {
		return err
	}

	// 操作后清理重复的标签关联（确保每个标签只关联一次）
	return s.deduplicateImageTags(imageID)
}

// deduplicateImageTags 清理图片的重复标签关联，确保每个标签只关联一次
// 保留第一个出现的关联（按ID排序），删除后续重复的关联
func (s *TagService) deduplicateImageTags(imageID uint) error {
	// 查找该图片的所有标签关联，按ID排序以确保一致性
	var associations []models.ImageTag
	if err := s.db.Where("image_id = ?", imageID).Order("id ASC").Find(&associations).Error; err != nil {
		return err
	}

	// 使用map记录已出现的tagID
	seenTags := make(map[uint]bool)
	var toDelete []uint

	for _, assoc := range associations {
		if seenTags[assoc.TagID] {
			// 如果已经见过这个tagID，标记为删除（保留第一个）
			toDelete = append(toDelete, assoc.ID)
		} else {
			// 第一次见到这个tagID，记录下来
			seenTags[assoc.TagID] = true
		}
	}

	// 删除重复的关联
	if len(toDelete) > 0 {
		if err := s.db.Where("id IN ?", toDelete).Delete(&models.ImageTag{}).Error; err != nil {
			return err
		}
	}

	return nil
}
