package handlers

import (
	"net/http"

	"image-manager/internal/dto"
	"image-manager/internal/services"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	tagService *services.TagService
}

func NewTagHandler(tagService *services.TagService) *TagHandler {
	return &TagHandler{tagService: tagService}
}

func (h *TagHandler) Create(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	var req dto.CreateTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	tag, err := h.tagService.Create(userID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tag)
}

func (h *TagHandler) List(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	tags, err := h.tagService.List(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, tags)
}

func (h *TagHandler) Assign(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	imageID := parseUint(ctx.Param("id"))

	var req dto.AssignTagsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.tagService.AssignBulk(userID, imageID, req.TagIDs); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"assigned": true})
}

func (h *TagHandler) Remove(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	imageID := parseUint(ctx.Param("id"))
	tagID := parseUint(ctx.Param("tagId"))

	if err := h.tagService.Remove(imageID, tagID, userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"removed": true})
}

func (h *TagHandler) UpdateColor(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	tagID := parseUint(ctx.Param("id"))
	
	var req dto.UpdateTagColorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	tag, err := h.tagService.UpdateColor(userID, tagID, req.Color)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tag)
}

func (h *TagHandler) UpdateImageTag(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	imageID := parseUint(ctx.Param("id"))
	
	var req dto.UpdateImageTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.tagService.UpdateImageTag(userID, imageID, req.OldTagID, req.NewTagName); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"updated": true})
}

func (h *TagHandler) AddImageTag(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	imageID := parseUint(ctx.Param("id"))
	
	var req dto.AddImageTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.tagService.AddImageTagByName(userID, imageID, req.TagName); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"added": true})
}

// Delete 删除标签
// 删除标签时，同时删除所有图片与该标签的关联
// 路由: DELETE /api/v1/tags/:id
func (h *TagHandler) Delete(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	tagID := parseUint(ctx.Param("id"))

	if err := h.tagService.Delete(userID, tagID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}
