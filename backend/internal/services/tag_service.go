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

	for _, name := range names {
		tag := models.Tag{
			UserID: userID,
			Name:   name,
		}
		if err := s.db.Where("user_id = ? AND name = ?", userID, name).FirstOrCreate(&tag).Error; err != nil {
			return err
		}
		if err := s.Assign(imageID, tag.ID, userID); err != nil {
			return err
		}
	}

	return nil
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
