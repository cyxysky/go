package handlers

import (
	"net/http"
	"strconv"

	"gin-web-api/models"
	"gin-web-api/services"

	"github.com/gin-gonic/gin"
)

type FormHandler struct {
	formService *services.FormService
}

func NewFormHandler() *FormHandler {
	return &FormHandler{
		formService: services.NewFormService(),
	}
}

// CreateFormDefinition 创建表单定义
func (h *FormHandler) CreateFormDefinition(c *gin.Context) {
	var req services.CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	form, err := h.formService.CreateFormDefinition(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": form})
}

// GetFormDefinitions 获取表单定义列表
func (h *FormHandler) GetFormDefinitions(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	forms, total, err := h.formService.ListFormDefinitions(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取表单列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"forms": forms,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		},
	})
}

// GetFormDefinition 获取表单定义详情
func (h *FormHandler) GetFormDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单ID"})
		return
	}

	form, err := h.formService.GetFormDefinition(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "表单不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": form})
}

// GetFormDefinitionByKey 根据key获取表单定义
func (h *FormHandler) GetFormDefinitionByKey(c *gin.Context) {
	formKey := c.Param("key")
	if formKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "表单key不能为空"})
		return
	}

	form, err := h.formService.GetFormDefinitionByKey(formKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "表单不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": form})
}

// CreateFormData 创建表单数据
func (h *FormHandler) CreateFormData(c *gin.Context) {
	var req services.CreateFormDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	formData, err := h.formService.CreateFormData(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": formData})
}

// UpdateFormData 更新表单数据
func (h *FormHandler) UpdateFormData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单数据ID"})
		return
	}

	var req services.UpdateFormDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	formData, err := h.formService.UpdateFormData(uint(id), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": formData})
}

// GetFormData 获取表单数据
func (h *FormHandler) GetFormData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单数据ID"})
		return
	}

	formData, err := h.formService.GetFormData(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "表单数据不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": formData})
}

// RenderForm 渲染表单
func (h *FormHandler) RenderForm(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单ID"})
		return
	}

	// 获取表单值参数
	formValues := c.Query("form_values")

	renderData, err := h.formService.RenderFormWithData(uint(id), formValues)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": renderData})
}

// CreateFormFromJSON 从JSON创建表单（用于支持node.txt格式）
func (h *FormHandler) CreateFormFromJSON(c *gin.Context) {
	var req services.CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	form, err := h.formService.CreateFormDefinition(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "表单创建成功",
		"data":    form,
	})
}

// PreviewForm 预览表单
func (h *FormHandler) PreviewForm(c *gin.Context) {
	var req services.CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 不实际创建表单，只返回预览数据
	previewData := gin.H{
		"form_definition": req,
		"preview_url":     "/api/v1/forms/preview/" + req.Key,
		"card_count":      len(req.Cards),
		"button_count":    len(req.Buttons),
	}

	// 统计字段数量
	fieldCount := 0
	for _, card := range req.Cards {
		fieldCount += len(card.Attributes)
	}
	previewData["field_count"] = fieldCount

	c.JSON(http.StatusOK, gin.H{"data": previewData})
}

// ValidateFormData 验证表单数据
func (h *FormHandler) ValidateFormData(c *gin.Context) {
	var req struct {
		FormID     uint   `json:"form_id" binding:"required"`
		FormValues string `json:"form_values" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 通过创建临时表单数据来验证
	userID := c.GetUint("user_id")
	tempReq := &services.CreateFormDataRequest{
		FormID:      req.FormID,
		BusinessKey: "temp-validation",
		FormValues:  req.FormValues,
	}

	_, err := h.formService.CreateFormData(tempReq, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	// 删除临时数据（实际实现中可能需要更好的验证方式）
	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "表单数据验证通过",
	})
}

// UpdateFormDefinition 更新表单定义
func (h *FormHandler) UpdateFormDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单ID"})
		return
	}

	var req services.CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	form, err := h.formService.UpdateFormDefinition(uint(id), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "表单更新成功",
		"data":    form,
	})
}

// DeleteFormDefinition 删除表单定义
func (h *FormHandler) DeleteFormDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单ID"})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.formService.DeleteFormDefinition(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "表单删除成功"})
}

// ExportFormDefinition 导出表单定义
func (h *FormHandler) ExportFormDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单ID"})
		return
	}

	exportData, err := h.formService.ExportFormDefinition(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "表单导出成功",
		"data":    exportData,
	})
}

// CloneFormDefinition 克隆表单定义  
func (h *FormHandler) CloneFormDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单ID"})
		return
	}

	var req struct {
		NewName string `json:"new_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	form, err := h.formService.CloneFormDefinition(uint(id), req.NewName, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "表单克隆成功",
		"data":    form,
	})
}

// ActivateFormDefinition 激活表单定义
func (h *FormHandler) ActivateFormDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单ID"})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.formService.ActivateFormDefinition(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "表单已激活"})
}

// DeactivateFormDefinition 停用表单定义
func (h *FormHandler) DeactivateFormDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的表单ID"})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.formService.DeactivateFormDefinition(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "表单已停用"})
} 