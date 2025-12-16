package dto

import "mime/multipart"

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=6,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UploadImageRequest struct {
	File multipart.FileHeader `form:"file" binding:"required"`
	Tags []string             `form:"tags[]" binding:"omitempty"`
}

type AssignTagsRequest struct {
	TagIDs []uint `json:"tagIds" binding:"required"`
}

type CreateTagRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=50"`
	Color string `json:"color" binding:"required"` // 手动创建标签时必须指定颜色
}

type UpdateTagColorRequest struct {
	Color string `json:"color" binding:"required"`
}

type UpdateImageTagRequest struct {
	OldTagID   uint   `json:"oldTagId" binding:"required"`
	NewTagName string `json:"newTagName" binding:"required,min=1,max=50"`
}

type AddImageTagRequest struct {
	TagName string `json:"tagName" binding:"required,min=1,max=50"`
}

type CropRequest struct {
	X      int `json:"x" binding:"required"`
	Y      int `json:"y" binding:"required"`
	Width  int `json:"width" binding:"required,gt=0"`
	Height int `json:"height" binding:"required,gt=0"`
}

type AdjustRequest struct {
	Brightness int `json:"brightness" binding:"gte=-100,lte=100"`
	Contrast   int `json:"contrast" binding:"gte=-100,lte=100"`
	Saturation int `json:"saturation" binding:"gte=-100,lte=100"`
	Hue        int `json:"hue" binding:"gte=-180,lte=180"`
}
