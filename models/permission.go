package models

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色
type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`        // 角色名称
	Code        string         `json:"code" gorm:"uniqueIndex;not null"`        // 角色编码
	Description string         `json:"description"`                             // 角色描述
	IsSystem    bool           `json:"is_system" gorm:"default:false"`          // 是否系统角色
	IsActive    bool           `json:"is_active" gorm:"default:true"`           // 是否启用
	Permissions []Permission   `json:"permissions" gorm:"many2many:role_permissions;"` // 角色权限
	Users       []User         `json:"users" gorm:"many2many:user_roles;"`      // 角色用户
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Permission 权限
type Permission struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`        // 权限名称
	Code        string         `json:"code" gorm:"uniqueIndex;not null"`        // 权限编码
	Resource    string         `json:"resource"`                                // 资源标识
	Action      string         `json:"action"`                                  // 操作类型
	Description string         `json:"description"`                             // 权限描述
	Category    string         `json:"category"`                                // 权限分类
	IsSystem    bool           `json:"is_system" gorm:"default:false"`          // 是否系统权限
	Roles       []Role         `json:"roles" gorm:"many2many:role_permissions;"` // 权限角色
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserRole 用户角色关联
type UserRole struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id"`                        // 用户ID
	User       User           `json:"user" gorm:"foreignKey:UserID"`  // 用户信息
	RoleID     uint           `json:"role_id"`                        // 角色ID
	Role       Role           `json:"role" gorm:"foreignKey:RoleID"`  // 角色信息  
	GrantedBy  uint           `json:"granted_by"`                     // 授权人ID
	Granter    User           `json:"granter" gorm:"foreignKey:GrantedBy"` // 授权人信息
	ExpiresAt  *time.Time     `json:"expires_at"`                     // 过期时间
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// RolePermission 角色权限关联
type RolePermission struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	RoleID       uint           `json:"role_id"`                                 // 角色ID
	Role         Role           `json:"role" gorm:"foreignKey:RoleID"`           // 角色信息
	PermissionID uint           `json:"permission_id"`                           // 权限ID
	Permission   Permission     `json:"permission" gorm:"foreignKey:PermissionID"` // 权限信息
	GrantedBy    uint           `json:"granted_by"`                              // 授权人ID
	Granter      User           `json:"granter" gorm:"foreignKey:GrantedBy"`     // 授权人信息
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// Department 部门
type Department struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`                    // 部门名称
	Code        string         `json:"code" gorm:"uniqueIndex"`                 // 部门编码
	ParentID    *uint          `json:"parent_id"`                               // 上级部门ID
	Parent      *Department    `json:"parent" gorm:"foreignKey:ParentID"`       // 上级部门
	Children    []Department   `json:"children" gorm:"foreignKey:ParentID"`     // 下级部门
	ManagerID   *uint          `json:"manager_id"`                              // 部门负责人ID
	Manager     *User          `json:"manager" gorm:"foreignKey:ManagerID"`     // 部门负责人
	Users       []User         `json:"users" gorm:"many2many:user_departments;"` // 部门用户
	Level       int            `json:"level" gorm:"default:1"`                  // 部门层级
	SortOrder   int            `json:"sort_order" gorm:"default:0"`             // 排序
	Description string         `json:"description"`                             // 部门描述
	IsActive    bool           `json:"is_active" gorm:"default:true"`           // 是否启用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserDepartment 用户部门关联
type UserDepartment struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id"`                                      // 用户ID
	User         User           `json:"user" gorm:"foreignKey:UserID"`                // 用户信息
	DepartmentID uint           `json:"department_id"`                                // 部门ID
	Department   Department     `json:"department" gorm:"foreignKey:DepartmentID"`    // 部门信息
	IsMain       bool           `json:"is_main" gorm:"default:false"`                 // 是否主部门
	Position     string         `json:"position"`                                     // 职位
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// 扩展用户模型，添加权限相关字段
type UserExtended struct {
	User
	Roles       []Role       `json:"roles" gorm:"many2many:user_roles;"`          // 用户角色
	Departments []Department `json:"departments" gorm:"many2many:user_departments;"` // 用户部门
}

// 预定义权限常量
const (
	// 工作流权限
	PermissionWorkflowCreate = "workflow:create"
	PermissionWorkflowRead   = "workflow:read"
	PermissionWorkflowUpdate = "workflow:update"
	PermissionWorkflowDelete = "workflow:delete"
	PermissionWorkflowDeploy = "workflow:deploy"
	
	// 实例权限
	PermissionInstanceCreate   = "instance:create"
	PermissionInstanceRead     = "instance:read"
	PermissionInstanceCancel   = "instance:cancel"
	PermissionInstanceSuspend  = "instance:suspend"
	PermissionInstanceTransfer = "instance:transfer"
	
	// 任务权限
	PermissionTaskApprove = "task:approve"
	PermissionTaskReject  = "task:reject"
	PermissionTaskClaim   = "task:claim"
	PermissionTaskDelegate = "task:delegate"
	
	// 系统权限
	PermissionSystemAdmin = "system:admin"
	PermissionUserManage  = "user:manage"
	PermissionRoleManage  = "role:manage"
)

// 预定义角色常量
const (
	RoleAdmin      = "admin"      // 系统管理员
	RoleWorkflowAdmin = "workflow_admin" // 工作流管理员
	RoleUser       = "user"       // 普通用户
	RoleApprover   = "approver"   // 审批人
) 