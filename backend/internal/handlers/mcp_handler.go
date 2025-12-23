// Package handlers 提供HTTP请求处理器
// mcp_handler.go 实现了MCP（Model Context Protocol）相关的HTTP处理器，用于对话式图片检索
package handlers

import (
	"net/http"

	"image-manager/internal/services"

	"github.com/gin-gonic/gin"
)

// MCPHandler MCP处理器结构体
// 处理对话式图片检索的HTTP请求
type MCPHandler struct {
	imageService *services.ImageService
	aiService    *services.AIService
	tagService   *services.TagService
}

// NewMCPHandler 创建MCP处理器实例
// 参数:
//   - imageService: 图片服务实例
//   - aiService: AI服务实例
//   - tagService: 标签服务实例
// 返回: MCPHandler指针
func NewMCPHandler(imageService *services.ImageService, aiService *services.AIService, tagService *services.TagService) *MCPHandler {
	return &MCPHandler{
		imageService: imageService,
		aiService:    aiService,
		tagService:   tagService,
	}
}

// SearchRequest 对话式搜索请求结构
type SearchRequest struct {
	Query string `json:"query" binding:"required"`  // 自然语言查询字符串
	Page  int    `json:"page"`                      // 页码，默认为1
	PageSize int `json:"pageSize"`                  // 每页数量，默认为20
}

// Search 对话式图片搜索
// 接收自然语言查询，使用AI转换为搜索条件，然后搜索图片
// 路由: POST /api/v1/mcp/search
// 请求体: {"query": "找一些风景照片", "page": 1, "pageSize": 20}
// 返回: 搜索到的图片列表和总数
func (h *MCPHandler) Search(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req SearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误: " + err.Error()})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 先获取用户已有的标签库，让AI优先从中选择标签
	existingTags, err := h.tagService.List(userID)
	existingTagNames := []string{}
	if err == nil {
		for _, tag := range existingTags {
			existingTagNames = append(existingTagNames, tag.Name)
		}
	}

	// 使用AI将自然语言查询转换为搜索过滤器，传入已有标签库
	filters, err := h.aiService.ConvertQueryToFilters(req.Query, existingTagNames)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "AI查询转换失败: " + err.Error(),
		})
		return
	}

	// AI搜索时，可选部分始终选择OR逻辑
	// keyword_mode: 关键词和其他条件使用OR关系
	// tag_mode: 标签之间使用OR关系
	filters["keyword_mode"] = "or"
	filters["tag_mode"] = "or"

	// 调用图片服务的List方法进行搜索
	images, total, err := h.imageService.List(userID, filters, req.Page, req.PageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "搜索失败: " + err.Error(),
		})
		return
	}

	// 返回搜索结果
	ctx.JSON(http.StatusOK, gin.H{
		"query":   req.Query,           // 原始查询
		"filters": filters,              // 转换后的过滤器
		"total":   total,                // 总数量
		"page":    req.Page,             // 当前页码
		"pageSize": req.PageSize,        // 每页数量
		"items":   images,               // 图片列表
	})
}

