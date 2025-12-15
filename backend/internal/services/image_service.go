package services

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"image-manager/internal/config"
	"image-manager/internal/dto"
	"image-manager/internal/models"

	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/rwcarlsen/goexif/exif"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ImageService struct {
	db   *gorm.DB
	cfg  config.Config
	tags *TagService
}

func NewImageService(db *gorm.DB, cfg config.Config, tags *TagService) *ImageService {
	return &ImageService{
		db:   db,
		cfg:  cfg,
		tags: tags,
	}
}

func (s *ImageService) Upload(userID uint, fileHeader *multipart.FileHeader, tagNames []string) (*models.Image, error) {
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

	imgCfg, format, err := image.DecodeConfig(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		return nil, errors.New("无法解析图片")
	}
	mimeType := fmt.Sprintf("image/%s", format)

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(fileHeader.Filename))
	destPath := filepath.Join(s.cfg.StorageDir, "originals", filename)
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return nil, err
	}

	if err := os.WriteFile(destPath, buffer.Bytes(), 0o644); err != nil {
		return nil, err
	}

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

	if err := s.db.Create(imageModel).Error; err != nil {
		return nil, err
	}

	if err := s.extractAndSaveEXIF(imageModel.ID, bytes.NewReader(buffer.Bytes())); err != nil {
		log.Printf("failed to parse EXIF: %v", err)
	}

	if err := s.generateThumbnail(imageModel.ID, bytes.NewReader(buffer.Bytes())); err != nil {
		log.Printf("failed to generate thumbnail: %v", err)
	}

	if len(tagNames) > 0 {
		if err := s.tags.AssignByNames(userID, imageModel.ID, tagNames); err != nil {
			log.Printf("failed to assign tags: %v", err)
		}
	}

	return imageModel, nil
}

func (s *ImageService) extractAndSaveEXIF(imageID uint, reader io.Reader) error {
	exifData, err := exif.Decode(reader)
	if err != nil {
		return err
	}

	exifModel := models.ImageEXIF{
		ImageID: imageID,
	}

	if tag, err := exifData.Get(exif.DateTimeOriginal); err == nil {
		if ts, err := tag.StringVal(); err == nil {
			if parsed, parseErr := time.Parse("2006:01:02 15:04:05", ts); parseErr == nil {
				exifModel.TakenAt = parsed
			}
		}
	}

	if tag, err := exifData.Get(exif.Model); err == nil {
		exifModel.CameraModel, _ = tag.StringVal()
	}
	if tag, err := exifData.Get(exif.Make); err == nil {
		exifModel.CameraMake, _ = tag.StringVal()
	}

	return s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&exifModel).Error
}

func (s *ImageService) generateThumbnail(imageID uint, reader io.Reader) error {
	img, err := imaging.Decode(reader)
	if err != nil {
		return err
	}

	thumb := imaging.Fill(img, s.cfg.ThumbnailWidth, s.cfg.ThumbnailHeight, imaging.Center, imaging.Lanczos)

	buff := &bytes.Buffer{}
	if err := jpeg.Encode(buff, thumb, &jpeg.Options{Quality: 85}); err != nil {
		return err
	}

	thumbnail := models.Thumbnail{
		ImageID: imageID,
		Data:    buff.Bytes(),
		Width:   s.cfg.ThumbnailWidth,
		Height:  s.cfg.ThumbnailHeight,
		Size:    buff.Len(),
	}

	return s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&thumbnail).Error
}

func (s *ImageService) List(userID uint, filters map[string]string, page, pageSize int) ([]models.Image, int64, error) {
	var images []models.Image
	var total int64

	query := s.db.Model(&models.Image{}).Preload("Thumbnail").Preload("Exif").Preload("Tags").Where("user_id = ?", userID)

	if keyword, ok := filters["keyword"]; ok && keyword != "" {
		query = query.Where("original_filename LIKE ?", "%"+keyword+"%")
	}
	if start, ok := filters["start"]; ok && start != "" {
		query = query.Where("created_at >= ?", start)
	}
	if end, ok := filters["end"]; ok && end != "" {
		query = query.Where("created_at <= ?", end)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&images).Error; err != nil {
		return nil, 0, err
	}

	return images, total, nil
}

func (s *ImageService) Get(userID, imageID uint) (*models.Image, error) {
	var imageModel models.Image
	if err := s.db.Preload("Thumbnail").Preload("Exif").Preload("Tags").Where("user_id = ? AND id = ?", userID, imageID).First(&imageModel).Error; err != nil {
		return nil, err
	}
	return &imageModel, nil
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
