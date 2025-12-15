package handlers

import (
	"net/http"
	"strconv"

	"image-manager/internal/dto"
	"image-manager/internal/services"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageService *services.ImageService
	tagService   *services.TagService
}

func NewImageHandler(imageService *services.ImageService, tagService *services.TagService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
		tagService:   tagService,
	}
}

func (h *ImageHandler) Upload(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请选择要上传的图片"})
		return
	}

	tags := ctx.PostFormArray("tags[]")
	image, err := h.imageService.Upload(userID, file, tags)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, image)
}

func (h *ImageHandler) List(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	page := parseInt(ctx.DefaultQuery("page", "1"))
	pageSize := parseInt(ctx.DefaultQuery("pageSize", "20"))
	filters := map[string]string{
		"keyword": ctx.Query("keyword"),
		"start":   ctx.Query("start_date"),
		"end":     ctx.Query("end_date"),
	}

	images, total, err := h.imageService.List(userID, filters, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"items":    images,
	})
}

func (h *ImageHandler) Detail(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	imageID := parseUint(ctx.Param("id"))

	image, err := h.imageService.Get(userID, imageID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "图片不存在"})
		return
	}

	ctx.JSON(http.StatusOK, image)
}

func (h *ImageHandler) Delete(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	imageID := parseUint(ctx.Param("id"))

	if err := h.imageService.Delete(userID, imageID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *ImageHandler) Thumbnail(ctx *gin.Context) {
	imageID := parseUint(ctx.Param("id"))

	thumb, err := h.imageService.GetThumbnail(imageID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "缩略图不存在"})
		return
	}

	ctx.Data(http.StatusOK, "image/jpeg", thumb.Data)
}

func (h *ImageHandler) Original(ctx *gin.Context) {
	imageID := parseUint(ctx.Param("id"))

	imageModel, data, err := h.imageService.GetFile(imageID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "图片不存在"})
		return
	}

	ctx.Data(http.StatusOK, "image/"+imageModel.MimeType, data)
}

func (h *ImageHandler) Crop(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	imageID := parseUint(ctx.Param("id"))

	var req dto.CropRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	newImage, err := h.imageService.Crop(userID, imageID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, newImage)
}

func (h *ImageHandler) Adjust(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	imageID := parseUint(ctx.Param("id"))

	var req dto.AdjustRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	newImage, err := h.imageService.Adjust(userID, imageID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, newImage)
}

func parseInt(value string) int {
	i, _ := strconv.Atoi(value)
	if i <= 0 {
		return 1
	}
	return i
}

func parseUint(value string) uint {
	i, _ := strconv.Atoi(value)
	if i < 0 {
		return 0
	}
	return uint(i)
}
