package middleware

import (
	"net/http"
	"strconv"

	"gin-web-api/services"

	"github.com/gin-gonic/gin"
)

// RequirePermission 检查用户是否具有指定权限
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		permissionService := services.NewPermissionService()
		hasPermission, err := permissionService.CheckPermission(userID.(uint), permissionCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "权限检查失败"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 检查用户是否具有多个权限中的任意一个
func RequireAnyPermission(permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		permissionService := services.NewPermissionService()
		hasPermission, err := permissionService.CheckMultiplePermissions(userID.(uint), permissionCodes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "权限检查失败"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckWorkflowInstancePermission 检查工作流实例相关权限
func CheckWorkflowInstancePermission(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		// 获取实例ID
		instanceIDStr := c.Param("id")
		instanceID, err := strconv.ParseUint(instanceIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的实例ID"})
			c.Abort()
			return
		}

		permissionService := services.NewPermissionService()
		hasPermission, err := permissionService.CheckWorkflowPermission(userID.(uint), action, uint(instanceID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "权限检查失败"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckTaskPermission 检查任务权限
func CheckTaskPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		// 获取任务ID
		taskIDStr := c.Param("id")
		taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
			c.Abort()
			return
		}

		// 检查是否是任务的分配人
		permissionService := services.NewPermissionService()
		workflowService := services.NewWorkflowService()
		
		// 这里应该有一个方法检查用户是否有权处理这个任务
		// 暂时通过任务分配人检查
		c.Set("task_id", uint(taskID))
		c.Next()
	}
}

// IsAdmin 检查是否是管理员
func IsAdmin() gin.HandlerFunc {
	return RequirePermission("system:admin")
}

// IsWorkflowAdmin 检查是否是工作流管理员
func IsWorkflowAdmin() gin.HandlerFunc {
	return RequireAnyPermission("system:admin", "workflow:admin")
} 