package handlers

import (
	"net/http"
	"strconv"

	"gin-web-api/models"
	"gin-web-api/services"

	"github.com/gin-gonic/gin"
)

type WorkflowHandler struct {
	workflowService   *services.WorkflowService
	permissionService *services.PermissionService
	formService       *services.FormService
}

func NewWorkflowHandler() *WorkflowHandler {
	return &WorkflowHandler{
		workflowService:   services.NewWorkflowService(),
		permissionService: services.NewPermissionService(),
		formService:       services.NewFormService(),
	}
}

// CreateWorkflow 创建工作流定义
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req services.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	workflow, err := h.workflowService.CreateWorkflow(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": workflow})
}

// CreateWorkflowWithNodeTree 根据节点树创建工作流
func (h *WorkflowHandler) CreateWorkflowWithNodeTree(c *gin.Context) {
	var req services.CreateWorkflowWithTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	workflow, err := h.workflowService.CreateWorkflowWithNodeTree(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "工作流创建成功",
		"data":    workflow,
	})
}

// ImportWorkflowFromJSON 从JSON导入工作流（支持node.txt格式）
func (h *WorkflowHandler) ImportWorkflowFromJSON(c *gin.Context) {
	var req struct {
		Workflow services.CreateWorkflowWithTreeRequest `json:"workflow" binding:"required"`
		Form     *services.CreateFormRequest            `json:"form"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	
	// 如果有表单定义，先创建表单
	if req.Form != nil {
		form, err := h.formService.CreateFormDefinition(req.Form, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建表单失败: " + err.Error()})
			return
		}
		req.Workflow.FormID = form.ID
	}

	// 创建工作流
	workflow, err := h.workflowService.CreateWorkflowWithNodeTree(&req.Workflow, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建工作流失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "导入成功",
		"data": gin.H{
			"workflow": workflow,
			"form":     req.Form,
		},
	})
}

// GetWorkflows 获取工作流列表
func (h *WorkflowHandler) GetWorkflows(c *gin.Context) {
	var workflows []models.WorkflowDefinition
	
	// 这里可以添加分页和筛选逻辑
	db := h.workflowService.GetDB()
	if err := db.Preload("Creator").Preload("Form").Find(&workflows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取工作流列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workflows})
}

// GetWorkflow 获取工作流详情
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工作流ID"})
		return
	}

	var workflow models.WorkflowDefinition
	db := h.workflowService.GetDB()
	if err := db.Preload("Creator").
		Preload("Form.Cards.Attributes.Attribute").
		Preload("Form.Buttons").
		Preload("Nodes").
		Preload("RootNode").
		First(&workflow, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工作流不存在"})
		return
	}

	// 解析节点树结构
	nodeTree, _ := workflow.ParseNodeTree()

	response := gin.H{
		"workflow":  workflow,
		"node_tree": nodeTree,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetWorkflowNodeTree 获取工作流节点树
func (h *WorkflowHandler) GetWorkflowNodeTree(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工作流ID"})
		return
	}

	var workflow models.WorkflowDefinition
	db := h.workflowService.GetDB()
	if err := db.First(&workflow, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工作流不存在"})
		return
	}

	// 解析节点树结构
	nodeTree, err := workflow.ParseNodeTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析节点树失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": nodeTree})
}

// UpdateWorkflowStatus 更新工作流状态
func (h *WorkflowHandler) UpdateWorkflowStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工作流ID"})
		return
	}

	var req struct {
		Status models.WorkflowStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var workflow models.WorkflowDefinition
	db := h.workflowService.GetDB()
	if err := db.First(&workflow, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工作流不存在"})
		return
	}

	workflow.Status = req.Status
	if err := db.Save(&workflow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新工作流状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workflow})
}

// StartWorkflow 启动工作流实例
func (h *WorkflowHandler) StartWorkflow(c *gin.Context) {
	var req services.StartWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	instance, err := h.workflowService.StartWorkflow(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": instance})
}

// StartWorkflowWithForm 启动带表单的工作流实例
func (h *WorkflowHandler) StartWorkflowWithForm(c *gin.Context) {
	var req services.StartWorkflowWithFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	instance, err := h.workflowService.StartWorkflowWithForm(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "工作流实例启动成功",
		"data":    instance,
	})
}

// GetInstances 获取工作流实例列表
func (h *WorkflowHandler) GetInstances(c *gin.Context) {
	var instances []models.WorkflowInstance
	
	userID := c.GetUint("user_id")
	db := h.workflowService.GetDB()
	
	// 根据权限过滤实例
	hasReadPermission, _ := h.permissionService.CheckPermission(userID, models.PermissionInstanceRead)
	if hasReadPermission {
		// 管理员可以看到所有实例
		err := db.Preload("Workflow").
			Preload("Initiator").
			Preload("FormData.Form").
			Find(&instances).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取实例列表失败"})
			return
		}
	} else {
		// 普通用户只能看到自己发起的实例和需要自己审批的实例
		err := db.Preload("Workflow").
			Preload("Initiator").
			Preload("FormData.Form").
			Where("initiator_id = ? OR id IN (SELECT DISTINCT instance_id FROM workflow_tasks WHERE assignee_id = ?)", 
				userID, userID).
			Find(&instances).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取实例列表失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": instances})
}

// GetInstance 获取工作流实例详情
func (h *WorkflowHandler) GetInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的实例ID"})
		return
	}

	var instance models.WorkflowInstance
	db := h.workflowService.GetDB()
	if err := db.Preload("Workflow").
		Preload("Initiator").
		Preload("FormData.Form.Cards.Attributes.Attribute").
		Preload("FormData.Form.Buttons").
		Preload("Tasks.Assignee").
		First(&instance, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": instance})
}

// GetInstanceFormData 获取实例的表单数据
func (h *WorkflowHandler) GetInstanceFormData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的实例ID"})
		return
	}

	var instance models.WorkflowInstance
	db := h.workflowService.GetDB()
	if err := db.Preload("FormData.Form.Cards.Attributes.Attribute").
		Preload("FormData.Form.Buttons").
		First(&instance, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}

	if instance.FormData == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "该实例没有关联的表单数据"})
		return
	}

	// 渲染表单数据
	renderData, err := h.formService.RenderFormWithData(instance.FormData.FormID, instance.FormData.FormValues)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "渲染表单数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": renderData})
}

// CancelInstance 取消工作流实例
func (h *WorkflowHandler) CancelInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的实例ID"})
		return
	}

	var instance models.WorkflowInstance
	db := h.workflowService.GetDB()
	if err := db.First(&instance, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}

	if instance.Status != models.InstanceStatusRunning {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能取消运行中的实例"})
		return
	}

	instance.Status = models.InstanceStatusCancelled
	if err := db.Save(&instance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消实例失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "实例已取消"})
}

// GetMyTasks 获取我的待办任务
func (h *WorkflowHandler) GetMyTasks(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var tasks []models.WorkflowTask
	db := h.workflowService.GetDB()
	if err := db.Preload("Instance.Workflow").
		Preload("Instance.Initiator").
		Preload("Instance.FormData.Form").
		Where("assignee_id = ? AND status = ?", userID, models.TaskStatusPending).
		Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待办任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tasks})
}

// ApproveTask 审批通过任务
func (h *WorkflowHandler) ApproveTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	var req struct {
		Comment    string `json:"comment"`
		FormValues string `json:"form_values"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.workflowService.ApproveTaskWithForm(uint(id), userID, req.Comment, req.FormValues); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已审批通过"})
}

// RejectTask 拒绝任务
func (h *WorkflowHandler) RejectTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	var req struct {
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.workflowService.RejectTask(uint(id), userID, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已拒绝"})
}

// GetInstanceHistory 获取实例历史记录
func (h *WorkflowHandler) GetInstanceHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的实例ID"})
		return
	}

	var history []models.WorkflowHistory
	db := h.workflowService.GetDB()
	if err := db.Preload("Operator").Where("instance_id = ?", uint(id)).
		Order("created_at ASC").Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": history})
}

// GetWorkflowStatistics 获取工作流统计信息
func (h *WorkflowHandler) GetWorkflowStatistics(c *gin.Context) {
	db := h.workflowService.GetDB()
	
	// 统计工作流数量
	var workflowCount int64
	db.Model(&models.WorkflowDefinition{}).Count(&workflowCount)
	
	// 统计实例数量
	var instanceCount int64
	db.Model(&models.WorkflowInstance{}).Count(&instanceCount)
	
	// 统计表单数量
	var formCount int64
	db.Model(&models.FormDefinition{}).Count(&formCount)
	
	// 统计待办任务数量
	userID := c.GetUint("user_id")
	var pendingTaskCount int64
	db.Model(&models.WorkflowTask{}).Where("assignee_id = ? AND status = ?", 
		userID, models.TaskStatusPending).Count(&pendingTaskCount)
	
	// 统计各状态的实例数量
	var statusStats []struct {
		Status models.InstanceStatus `json:"status"`
		Count  int64                 `json:"count"`
	}
	db.Model(&models.WorkflowInstance{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusStats)

	stats := gin.H{
		"workflow_count":      workflowCount,
		"instance_count":      instanceCount,
		"form_count":          formCount,
		"pending_task_count":  pendingTaskCount,
		"status_statistics":   statusStats,
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
} 