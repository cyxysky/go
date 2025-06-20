package routes

import (
	"gin-web-api/config"
	"gin-web-api/handlers"
	"gin-web-api/middleware"
	"gin-web-api/models"

	"github.com/gin-gonic/gin"
)

func SetupWorkflowRoutes(r *gin.Engine, cfg *config.Config) {
	workflowHandler := handlers.NewWorkflowHandler()
	permissionHandler := handlers.NewPermissionHandler()
	formHandler := handlers.NewFormHandler()

	// API v1 路由组
	api := r.Group("/api/v1")
	api.Use(middleware.JWTMiddleware(cfg))

	// 表单管理路由
	formGroup := api.Group("/forms")
	{
		// 获取表单列表
		formGroup.GET("", 
			middleware.RequirePermission(models.PermissionWorkflowRead), 
			formHandler.GetFormDefinitions)
		
		// 创建表单定义
		formGroup.POST("", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			formHandler.CreateFormDefinition)
		
		// 从JSON创建表单（支持node.txt格式）
		formGroup.POST("/import", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			formHandler.CreateFormFromJSON)
		
		// 预览表单
		formGroup.POST("/preview", 
			middleware.RequirePermission(models.PermissionWorkflowRead), 
			formHandler.PreviewForm)
		
		// 获取表单定义详情
		formGroup.GET("/:id", 
			middleware.RequirePermission(models.PermissionWorkflowRead), 
			formHandler.GetFormDefinition)
		
		// 更新表单定义
		formGroup.PUT("/:id", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			formHandler.UpdateFormDefinition)
		
		// 删除表单定义
		formGroup.DELETE("/:id", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			formHandler.DeleteFormDefinition)
		
		// 导出表单定义
		formGroup.GET("/:id/export", 
			middleware.RequirePermission(models.PermissionWorkflowRead), 
			formHandler.ExportFormDefinition)
		
		// 克隆表单定义
		formGroup.POST("/:id/clone", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			formHandler.CloneFormDefinition)
		
		// 激活表单定义
		formGroup.PUT("/:id/activate", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			formHandler.ActivateFormDefinition)
		
		// 停用表单定义
		formGroup.PUT("/:id/deactivate", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			formHandler.DeactivateFormDefinition)
		
		// 根据key获取表单定义
		formGroup.GET("/key/:key", 
			middleware.RequirePermission(models.PermissionWorkflowRead), 
			formHandler.GetFormDefinitionByKey)
		
		// 渲染表单
		formGroup.GET("/:id/render", 
			middleware.RequirePermission(models.PermissionWorkflowRead), 
			formHandler.RenderForm)
		
		// 验证表单数据
		formGroup.POST("/validate", 
			middleware.RequirePermission(models.PermissionInstanceCreate), 
			formHandler.ValidateFormData)
	}

	// 表单数据路由
	formDataGroup := api.Group("/form-data")
	{
		// 创建表单数据
		formDataGroup.POST("", 
			middleware.RequirePermission(models.PermissionInstanceCreate), 
			formHandler.CreateFormData)
		
		// 获取表单数据
		formDataGroup.GET("/:id", 
			middleware.RequirePermission(models.PermissionInstanceRead), 
			formHandler.GetFormData)
		
		// 更新表单数据
		formDataGroup.PUT("/:id", 
			middleware.RequirePermission(models.PermissionInstanceCreate), 
			formHandler.UpdateFormData)
	}

	// 工作流管理路由
	workflowGroup := api.Group("/workflows")
	{
		// 获取工作流列表 - 需要读取权限
		workflowGroup.GET("", 
			middleware.RequirePermission(models.PermissionWorkflowRead), 
			workflowHandler.GetWorkflows)
		
		// 创建工作流 - 需要创建权限
		workflowGroup.POST("", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			workflowHandler.CreateWorkflow)
		
		// 根据节点树创建工作流
		workflowGroup.POST("/with-tree", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			workflowHandler.CreateWorkflowWithNodeTree)
		
		// 从JSON导入工作流（支持node.txt格式）
		workflowGroup.POST("/import", 
			middleware.RequirePermission(models.PermissionWorkflowCreate), 
			workflowHandler.ImportWorkflowFromJSON)
		
		// 获取工作流详情 - 需要读取权限
		workflowGroup.GET("/:id", 
			middleware.RequirePermission(models.PermissionWorkflowRead), 
			workflowHandler.GetWorkflow)
		
		// 获取工作流节点树
		workflowGroup.GET("/:id/node-tree", 
			middleware.RequirePermission(models.PermissionWorkflowRead), 
			workflowHandler.GetWorkflowNodeTree)
		
		// 更新工作流状态 - 需要部署权限
		workflowGroup.PUT("/:id/status", 
			middleware.RequirePermission(models.PermissionWorkflowDeploy), 
			workflowHandler.UpdateWorkflowStatus)
	}

	// 工作流实例路由
	instanceGroup := api.Group("/instances")
	{
		// 启动工作流实例 - 需要创建实例权限
		instanceGroup.POST("", 
			middleware.RequirePermission(models.PermissionInstanceCreate), 
			workflowHandler.StartWorkflow)
		
		// 启动带表单的工作流实例
		instanceGroup.POST("/with-form", 
			middleware.RequirePermission(models.PermissionInstanceCreate), 
			workflowHandler.StartWorkflowWithForm)
		
		// 获取实例列表
		instanceGroup.GET("", workflowHandler.GetInstances)
		
		// 获取实例详情 - 需要查看权限检查
		instanceGroup.GET("/:id", 
			middleware.CheckWorkflowInstancePermission("view_instance"), 
			workflowHandler.GetInstance)
		
		// 获取实例的表单数据
		instanceGroup.GET("/:id/form-data", 
			middleware.CheckWorkflowInstancePermission("view_instance"), 
			workflowHandler.GetInstanceFormData)
		
		// 取消实例 - 需要取消权限检查
		instanceGroup.PUT("/:id/cancel", 
			middleware.CheckWorkflowInstancePermission("cancel_instance"), 
			workflowHandler.CancelInstance)
		
		// 获取实例历史记录
		instanceGroup.GET("/:id/history", 
			middleware.CheckWorkflowInstancePermission("view_instance"), 
			workflowHandler.GetInstanceHistory)
	}

	// 任务路由
	taskGroup := api.Group("/tasks")
	{
		// 获取我的待办任务
		taskGroup.GET("/my", workflowHandler.GetMyTasks)
		
		// 审批任务 - 需要审批权限
		taskGroup.POST("/:id/approve", 
			middleware.RequirePermission(models.PermissionTaskApprove), 
			middleware.CheckTaskPermission(), 
			workflowHandler.ApproveTask)
		
		// 拒绝任务 - 需要拒绝权限
		taskGroup.POST("/:id/reject", 
			middleware.RequirePermission(models.PermissionTaskReject), 
			middleware.CheckTaskPermission(), 
			workflowHandler.RejectTask)
	}

	// 统计信息路由
	api.GET("/workflow/statistics", workflowHandler.GetWorkflowStatistics)

	// 权限管理路由 - 只有管理员可以访问
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.IsAdmin())
	{
		// 角色管理
		roleGroup := adminGroup.Group("/roles")
		{
			roleGroup.GET("", permissionHandler.GetRoles)
			roleGroup.POST("", permissionHandler.CreateRole)
			roleGroup.PUT("/:id", permissionHandler.UpdateRole)
			roleGroup.DELETE("/:id", permissionHandler.DeleteRole)
			roleGroup.GET("/:id/permissions", permissionHandler.GetRolePermissions)
			roleGroup.POST("/:id/permissions", permissionHandler.AssignPermissionToRole)
			roleGroup.DELETE("/:id/permissions/:permission_id", permissionHandler.RemovePermissionFromRole)
		}

		// 权限管理
		permissionGroup := adminGroup.Group("/permissions")
		{
			permissionGroup.GET("", permissionHandler.GetPermissions)
			permissionGroup.POST("", permissionHandler.CreatePermission)
		}

		// 用户权限管理
		userGroup := adminGroup.Group("/users")
		{
			userGroup.GET("/:id/roles", permissionHandler.GetUserRoles)
			userGroup.GET("/:id/permissions", permissionHandler.GetUserPermissions)
			userGroup.POST("/:id/roles", permissionHandler.AssignRoleToUser)
			userGroup.DELETE("/:id/roles/:role_id", permissionHandler.RemoveRoleFromUser)
		}

		// 系统初始化
		adminGroup.POST("/initialize", permissionHandler.InitializeData)
	}

	// 用户个人信息路由
	api.GET("/profile", permissionHandler.GetMyProfile)
	api.POST("/check-permission", permissionHandler.CheckPermission)
} 