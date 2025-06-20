package services

import (
	"errors"
	"fmt"

	"gin-web-api/database"
	"gin-web-api/models"

	"gorm.io/gorm"
)

type PermissionService struct {
	db *gorm.DB
}

func NewPermissionService() *PermissionService {
	return &PermissionService{
		db: database.GetDB(),
	}
}

// CheckPermission 检查用户是否具有指定权限
func (s *PermissionService) CheckPermission(userID uint, permissionCode string) (bool, error) {
	var count int64
	
	// 通过用户角色查询权限
	err := s.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.code = ? AND user_roles.deleted_at IS NULL", userID, permissionCode).
		Count(&count).Error
	
	if err != nil {
		return false, fmt.Errorf("检查权限失败: %w", err)
	}
	
	return count > 0, nil
}

// CheckMultiplePermissions 检查用户是否具有多个权限中的任意一个
func (s *PermissionService) CheckMultiplePermissions(userID uint, permissionCodes []string) (bool, error) {
	for _, code := range permissionCodes {
		hasPermission, err := s.CheckPermission(userID, code)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}
	return false, nil
}

// GetUserPermissions 获取用户的所有权限
func (s *PermissionService) GetUserPermissions(userID uint) ([]models.Permission, error) {
	var permissions []models.Permission
	
	err := s.db.Table("permissions").
		Select("DISTINCT permissions.*").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.deleted_at IS NULL", userID).
		Find(&permissions).Error
	
	return permissions, err
}

// GetUserRoles 获取用户的所有角色
func (s *PermissionService) GetUserRoles(userID uint) ([]models.Role, error) {
	var roles []models.Role
	
	err := s.db.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.deleted_at IS NULL", userID).
		Find(&roles).Error
	
	return roles, err
}

// AssignRoleToUser 给用户分配角色
func (s *PermissionService) AssignRoleToUser(userID, roleID, grantedBy uint) error {
	// 检查角色是否存在
	var role models.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}
	
	// 检查是否已经分配过
	var count int64
	s.db.Model(&models.UserRole{}).Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count)
	if count > 0 {
		return errors.New("用户已经拥有此角色")
	}
	
	// 创建用户角色关联
	userRole := &models.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		GrantedBy: grantedBy,
	}
	
	return s.db.Create(userRole).Error
}

// RemoveRoleFromUser 移除用户角色
func (s *PermissionService) RemoveRoleFromUser(userID, roleID uint) error {
	return s.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{}).Error
}

// CreateRole 创建角色
func (s *PermissionService) CreateRole(req *CreateRoleRequest) (*models.Role, error) {
	role := &models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsActive:    true,
	}
	
	if err := s.db.Create(role).Error; err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}
	
	return role, nil
}

// UpdateRole 更新角色
func (s *PermissionService) UpdateRole(roleID uint, req *UpdateRoleRequest) (*models.Role, error) {
	var role models.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return nil, fmt.Errorf("角色不存在: %w", err)
	}
	
	// 更新字段
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}
	
	if err := s.db.Save(&role).Error; err != nil {
		return nil, fmt.Errorf("更新角色失败: %w", err)
	}
	
	return &role, nil
}

// DeleteRole 删除角色
func (s *PermissionService) DeleteRole(roleID uint) error {
	// 检查是否为系统角色
	var role models.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}
	
	if role.IsSystem {
		return errors.New("系统角色不能删除")
	}
	
	// 删除角色权限关联
	s.db.Where("role_id = ?", roleID).Delete(&models.RolePermission{})
	
	// 删除用户角色关联
	s.db.Where("role_id = ?", roleID).Delete(&models.UserRole{})
	
	// 删除角色
	return s.db.Delete(&role).Error
}

// AssignPermissionToRole 给角色分配权限
func (s *PermissionService) AssignPermissionToRole(roleID, permissionID, grantedBy uint) error {
	// 检查权限是否存在
	var permission models.Permission
	if err := s.db.First(&permission, permissionID).Error; err != nil {
		return fmt.Errorf("权限不存在: %w", err)
	}
	
	// 检查是否已经分配过
	var count int64
	s.db.Model(&models.RolePermission{}).Where("role_id = ? AND permission_id = ?", roleID, permissionID).Count(&count)
	if count > 0 {
		return errors.New("角色已经拥有此权限")
	}
	
	// 创建角色权限关联
	rolePermission := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		GrantedBy:    grantedBy,
	}
	
	return s.db.Create(rolePermission).Error
}

// RemovePermissionFromRole 移除角色权限
func (s *PermissionService) RemovePermissionFromRole(roleID, permissionID uint) error {
	return s.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&models.RolePermission{}).Error
}

// GetRolePermissions 获取角色的所有权限
func (s *PermissionService) GetRolePermissions(roleID uint) ([]models.Permission, error) {
	var permissions []models.Permission
	
	err := s.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	
	return permissions, err
}

// CreatePermission 创建权限
func (s *PermissionService) CreatePermission(req *CreatePermissionRequest) (*models.Permission, error) {
	permission := &models.Permission{
		Name:        req.Name,
		Code:        req.Code,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
		Category:    req.Category,
	}
	
	if err := s.db.Create(permission).Error; err != nil {
		return nil, fmt.Errorf("创建权限失败: %w", err)
	}
	
	return permission, nil
}

// GetAllRoles 获取所有角色
func (s *PermissionService) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Find(&roles).Error
	return roles, err
}

// GetAllPermissions 获取所有权限
func (s *PermissionService) GetAllPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := s.db.Find(&permissions).Error
	return permissions, err
}

// IsWorkflowApprover 检查用户是否是指定工作流实例的审批人
func (s *PermissionService) IsWorkflowApprover(userID, instanceID uint) (bool, error) {
	var count int64
	
	err := s.db.Model(&models.WorkflowTask{}).
		Where("instance_id = ? AND assignee_id = ? AND status = ?", 
			instanceID, userID, models.TaskStatusPending).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// IsWorkflowInitiator 检查用户是否是指定工作流实例的发起人
func (s *PermissionService) IsWorkflowInitiator(userID, instanceID uint) (bool, error) {
	var count int64
	
	err := s.db.Model(&models.WorkflowInstance{}).
		Where("id = ? AND initiator_id = ?", instanceID, userID).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// CheckWorkflowPermission 检查工作流相关权限
func (s *PermissionService) CheckWorkflowPermission(userID uint, action string, resourceID uint) (bool, error) {
	switch action {
	case "view_instance":
		// 可以查看实例：发起人、审批人、管理员
		isInitiator, _ := s.IsWorkflowInitiator(userID, resourceID)
		if isInitiator {
			return true, nil
		}
		
		isApprover, _ := s.IsWorkflowApprover(userID, resourceID)
		if isApprover {
			return true, nil
		}
		
		return s.CheckPermission(userID, models.PermissionInstanceRead)
		
	case "cancel_instance":
		// 可以取消实例：发起人、管理员
		isInitiator, _ := s.IsWorkflowInitiator(userID, resourceID)
		if isInitiator {
			return true, nil
		}
		
		return s.CheckPermission(userID, models.PermissionInstanceCancel)
		
	default:
		return false, errors.New("未知的权限检查类型")
	}
}

// InitializeDefaultData 初始化默认数据
func (s *PermissionService) InitializeDefaultData() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 创建默认权限
		defaultPermissions := []models.Permission{
			{Name: "创建工作流", Code: models.PermissionWorkflowCreate, Resource: "workflow", Action: "create", Category: "工作流", IsSystem: true},
			{Name: "查看工作流", Code: models.PermissionWorkflowRead, Resource: "workflow", Action: "read", Category: "工作流", IsSystem: true},
			{Name: "编辑工作流", Code: models.PermissionWorkflowUpdate, Resource: "workflow", Action: "update", Category: "工作流", IsSystem: true},
			{Name: "删除工作流", Code: models.PermissionWorkflowDelete, Resource: "workflow", Action: "delete", Category: "工作流", IsSystem: true},
			{Name: "部署工作流", Code: models.PermissionWorkflowDeploy, Resource: "workflow", Action: "deploy", Category: "工作流", IsSystem: true},
			{Name: "创建实例", Code: models.PermissionInstanceCreate, Resource: "instance", Action: "create", Category: "实例", IsSystem: true},
			{Name: "查看实例", Code: models.PermissionInstanceRead, Resource: "instance", Action: "read", Category: "实例", IsSystem: true},
			{Name: "取消实例", Code: models.PermissionInstanceCancel, Resource: "instance", Action: "cancel", Category: "实例", IsSystem: true},
			{Name: "审批任务", Code: models.PermissionTaskApprove, Resource: "task", Action: "approve", Category: "任务", IsSystem: true},
			{Name: "拒绝任务", Code: models.PermissionTaskReject, Resource: "task", Action: "reject", Category: "任务", IsSystem: true},
			{Name: "系统管理", Code: models.PermissionSystemAdmin, Resource: "system", Action: "admin", Category: "系统", IsSystem: true},
		}
		
		for _, permission := range defaultPermissions {
			var count int64
			tx.Model(&models.Permission{}).Where("code = ?", permission.Code).Count(&count)
			if count == 0 {
				if err := tx.Create(&permission).Error; err != nil {
					return err
				}
			}
		}
		
		// 创建默认角色
		defaultRoles := []models.Role{
			{Name: "系统管理员", Code: models.RoleAdmin, Description: "系统超级管理员", IsSystem: true, IsActive: true},
			{Name: "工作流管理员", Code: models.RoleWorkflowAdmin, Description: "工作流管理员", IsSystem: true, IsActive: true},
			{Name: "普通用户", Code: models.RoleUser, Description: "普通用户", IsSystem: true, IsActive: true},
			{Name: "审批人", Code: models.RoleApprover, Description: "审批人员", IsSystem: true, IsActive: true},
		}
		
		for _, role := range defaultRoles {
			var count int64
			tx.Model(&models.Role{}).Where("code = ?", role.Code).Count(&count)
			if count == 0 {
				if err := tx.Create(&role).Error; err != nil {
					return err
				}
			}
		}
		
		// 给管理员角色分配所有权限
		var adminRole models.Role
		var allPermissions []models.Permission
		tx.Where("code = ?", models.RoleAdmin).First(&adminRole)
		tx.Find(&allPermissions)
		
		for _, permission := range allPermissions {
			var count int64
			tx.Model(&models.RolePermission{}).Where("role_id = ? AND permission_id = ?", adminRole.ID, permission.ID).Count(&count)
			if count == 0 {
				rolePermission := models.RolePermission{
					RoleID:       adminRole.ID,
					PermissionID: permission.ID,
					GrantedBy:    1, // 系统自动分配
				}
				tx.Create(&rolePermission)
			}
		}
		
		return nil
	})
}

// 请求结构体
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Category    string `json:"category"`
} 