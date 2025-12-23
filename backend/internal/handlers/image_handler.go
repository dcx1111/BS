package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"image-manager/internal/dto"
	"image-manager/internal/services"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageService *services.ImageService
	tagService   *services.TagService
	authService  *services.AuthService
}

func NewImageHandler(imageService *services.ImageService, tagService *services.TagService, authService *services.AuthService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
		tagService:   tagService,
		authService:  authService,
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
	// 解析是否使用AI的标志，默认为true（保持向后兼容）
	useAI := true
	if useAIStr := ctx.PostForm("use_ai"); useAIStr != "" {
		useAI = useAIStr == "true"
	}
	image, err := h.imageService.Upload(userID, file, tags, useAI)
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
		"keyword":      ctx.Query("keyword"),
		"start":        ctx.Query("start_date"),
		"end":          ctx.Query("end_date"),
		"taken_start":  ctx.Query("taken_start"),
		"taken_end":    ctx.Query("taken_end"),
		"width_min":    ctx.Query("width_min"),
		"width_max":    ctx.Query("width_max"),
		"height_min":   ctx.Query("height_min"),
		"height_max":   ctx.Query("height_max"),
		"size_min":     ctx.Query("size_min"),
		"size_max":     ctx.Query("size_max"),
		"tags":         ctx.Query("tags"),
		"keyword_mode": ctx.Query("keyword_mode"),  // "and" 或 "or"，表示关键词和其他条件的关系
		"tag_mode":     ctx.Query("tag_mode"),      // "and" 或 "or"，表示标签之间的关系
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

func (h *ImageHandler) Update(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	imageID := parseUint(ctx.Param("id"))
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请选择要上传的图片"})
		return
	}

	image, err := h.imageService.Update(userID, imageID, file)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
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

// ImportVerify 验证其他用户的凭据并获取其图片列表
func (h *ImageHandler) ImportVerify(ctx *gin.Context) {
	var req dto.ImportVerifyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// 验证用户凭据
	loginReq := dto.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}
	_, user, err := h.authService.Login(loginReq)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	// 获取该用户的所有图片
	images, err := h.imageService.GetOtherUserImages(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "获取图片列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"images": images,
		"userId": user.ID,
	})
}

// Import 导入其他用户的图片
func (h *ImageHandler) Import(ctx *gin.Context) {
	currentUserID := ctx.GetUint("user_id")
	var req dto.ImportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// 验证源用户凭据
	loginReq := dto.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}
	_, sourceUser, err := h.authService.Login(loginReq)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	// 防止用户导入自己的图片
	if sourceUser.ID == currentUserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "不能导入自己的图片"})
		return
	}

	// 执行导入
	importedImages, err := h.imageService.ImportImages(currentUserID, sourceUser.ID, req.ImageIDs, h.tagService)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("成功导入 %d 张图片", len(importedImages)),
		"importedImages": importedImages,
	})
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
