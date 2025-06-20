package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusDraft     WorkflowStatus = "draft"     // 草稿
	WorkflowStatusActive    WorkflowStatus = "active"    // 活跃
	WorkflowStatusInactive  WorkflowStatus = "inactive"  // 停用
)

// WorkflowDefinition 工作流定义
type WorkflowDefinition struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`                    // 工作流名称
	Description string         `json:"description"`                             // 描述
	Category    string         `json:"category"`                                // 分类
	Version     int            `json:"version" gorm:"default:1"`                // 版本号
	Status      WorkflowStatus `json:"status" gorm:"default:draft"`             // 状态
	IsDefault   bool           `json:"is_default" gorm:"default:false"`         // 是否默认版本
	FormID      *uint          `json:"form_id"`                                 // 关联表单ID
	Form        *FormDefinition `json:"form" gorm:"foreignKey:FormID"`          // 关联表单
	RootNodeID  *uint          `json:"root_node_id"`                            // 根节点ID
	RootNode    *WorkflowNode  `json:"root_node" gorm:"foreignKey:RootNodeID"`  // 根节点
	NodeData    string         `json:"node_data"`                               // 节点树结构(JSON)
	CreatedBy   uint           `json:"created_by"`                              // 创建者
	Creator     User           `json:"creator" gorm:"foreignKey:CreatedBy"`     // 创建者信息
	Nodes       []WorkflowNode `json:"nodes" gorm:"foreignKey:WorkflowID"`      // 节点列表
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// NodeType 节点类型
type NodeType string

const (
	NodeTypeRoot      NodeType = "ROOT"      // 根节点
	NodeTypeStart     NodeType = "start"     // 开始节点
	NodeTypeEnd       NodeType = "END"       // 结束节点
	NodeTypeApproval  NodeType = "approval"  // 审批节点
	NodeTypeCondition NodeType = "condition" // 条件节点
	NodeTypeParallel  NodeType = "parallel"  // 并行节点
	NodeTypeMerge     NodeType = "merge"     // 合并节点
)

// ApprovalMode 审批模式
type ApprovalMode string

const (
	ApprovalModeSequence ApprovalMode = "sequence" // 依次审批
	ApprovalModeParallel ApprovalMode = "parallel" // 并行审批
	ApprovalModeAny      ApprovalMode = "any"      // 任意一人审批
	ApprovalModeAll      ApprovalMode = "all"      // 全员审批
)

// WorkflowNode 工作流节点
type WorkflowNode struct {
	ID               uint              `json:"id" gorm:"primaryKey"`
	WorkflowID       uint              `json:"workflow_id"`                                       // 工作流ID
	NodeKey          string            `json:"key" gorm:"not null"`                              // 节点唯一标识
	Name             string            `json:"name" gorm:"not null"`                             // 节点名称
	Type             NodeType          `json:"type" gorm:"not null"`                             // 节点类型
	ApprovalMode     ApprovalMode      `json:"approval_mode" gorm:"default:sequence"`            // 审批模式
	ParentNodeID     *uint             `json:"parent_node_id"`                                   // 父节点ID
	ParentNode       *WorkflowNode     `json:"parent_node" gorm:"foreignKey:ParentNodeID"`       // 父节点
	ChildNodeID      *uint             `json:"child_node_id"`                                    // 子节点ID
	ChildNode        *WorkflowNode     `json:"child" gorm:"foreignKey:ChildNodeID"`              // 子节点
	Branches         []WorkflowBranch  `json:"branches" gorm:"foreignKey:NodeID"`                // 分支节点
	Position         string            `json:"position"`                                         // 节点位置信息(JSON)
	Assignees        string            `json:"assignees"`                                        // 审批人配置(JSON)
	Conditions       string            `json:"conditions"`                                       // 条件配置(JSON)
	Settings         string            `json:"settings"`                                         // 其他设置(JSON)
	NextNodes        string            `json:"next_nodes"`                                       // 下一个节点(JSON数组)
	FormConditions   string            `json:"form_conditions"`                                  // 表单条件(JSON)
	SortOrder        int               `json:"sort_order" gorm:"default:0"`                      // 排序
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	DeletedAt        gorm.DeletedAt    `json:"-" gorm:"index"`
}

// WorkflowBranch 工作流分支节点
type WorkflowBranch struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	NodeID       uint           `json:"node_id"`                                    // 父节点ID
	BranchKey    string         `json:"key" gorm:"not null"`                        // 分支唯一标识
	Name         string         `json:"name" gorm:"not null"`                       // 分支名称
	Type         NodeType       `json:"type" gorm:"not null"`                       // 分支类型
	ChildNodeID  *uint          `json:"child_node_id"`                              // 子节点ID
	ChildNode    *WorkflowNode  `json:"child" gorm:"foreignKey:ChildNodeID"`        // 子节点
	Conditions   string         `json:"conditions"`                                 // 分支条件(JSON)
	SortOrder    int            `json:"sort_order" gorm:"default:0"`                // 排序
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// InstanceStatus 实例状态
type InstanceStatus string

const (
	InstanceStatusRunning   InstanceStatus = "running"   // 运行中
	InstanceStatusApproved  InstanceStatus = "approved"  // 已通过
	InstanceStatusRejected  InstanceStatus = "rejected"  // 已拒绝
	InstanceStatusCancelled InstanceStatus = "cancelled" // 已取消
	InstanceStatusSuspended InstanceStatus = "suspended" // 已挂起
	InstanceStatusDraft     InstanceStatus = "draft"     // 草稿
)

// WorkflowInstance 工作流实例
type WorkflowInstance struct {
	ID                 uint             `json:"id" gorm:"primaryKey"`
	WorkflowID         uint             `json:"workflow_id"`                                     // 工作流定义ID
	Workflow           WorkflowDefinition `json:"workflow" gorm:"foreignKey:WorkflowID"`        // 工作流定义
	Title              string           `json:"title" gorm:"not null"`                          // 实例标题
	BusinessKey        string           `json:"business_key"`                                   // 业务标识
	BusinessType       string           `json:"business_type"`                                  // 业务类型
	BusinessData       string           `json:"business_data"`                                  // 业务数据(JSON)
	FormDataID         *uint            `json:"form_data_id"`                                   // 表单数据ID
	FormData           *FormData        `json:"form_data" gorm:"foreignKey:FormDataID"`         // 表单数据
	Status             InstanceStatus   `json:"status" gorm:"default:running"`                  // 实例状态
	CurrentNodes       string           `json:"current_nodes"`                                  // 当前节点(JSON数组)
	ExecutionPath      string           `json:"execution_path"`                                 // 执行路径(JSON)
	Variables          string           `json:"variables"`                                      // 流程变量(JSON)
	StartTime          time.Time        `json:"start_time"`                                     // 开始时间
	EndTime            *time.Time       `json:"end_time"`                                       // 结束时间
	InitiatorID        uint             `json:"initiator_id"`                                   // 发起人ID
	Initiator          User             `json:"initiator" gorm:"foreignKey:InitiatorID"`       // 发起人信息
	Tasks              []WorkflowTask   `json:"tasks" gorm:"foreignKey:InstanceID"`             // 任务列表
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	DeletedAt          gorm.DeletedAt   `json:"-" gorm:"index"`
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待处理
	TaskStatusApproved  TaskStatus = "approved"  // 已通过
	TaskStatusRejected  TaskStatus = "rejected"  // 已拒绝
	TaskStatusSkipped   TaskStatus = "skipped"   // 已跳过
	TaskStatusCancelled TaskStatus = "cancelled" // 已取消
	TaskStatusClaimed   TaskStatus = "claimed"   // 已认领
)

// WorkflowTask 工作流任务
type WorkflowTask struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	InstanceID   uint           `json:"instance_id"`                                   // 实例ID
	Instance     WorkflowInstance `json:"instance" gorm:"foreignKey:InstanceID"`      // 实例信息
	NodeKey      string         `json:"node_key" gorm:"not null"`                     // 节点标识
	NodeName     string         `json:"node_name" gorm:"not null"`                    // 节点名称
	AssigneeID   uint           `json:"assignee_id"`                                  // 处理人ID
	Assignee     User           `json:"assignee" gorm:"foreignKey:AssigneeID"`        // 处理人信息
	Status       TaskStatus     `json:"status" gorm:"default:pending"`                // 任务状态
	Comment      string         `json:"comment"`                                      // 处理意见
	FormValues   string         `json:"form_values"`                                  // 表单提交值(JSON)
	ProcessTime  *time.Time     `json:"process_time"`                                 // 处理时间
	DueTime      *time.Time     `json:"due_time"`                                     // 截止时间
	Priority     int            `json:"priority" gorm:"default:0"`                    // 优先级
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// WorkflowHistory 工作流历史记录
type WorkflowHistory struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	InstanceID  uint           `json:"instance_id"`                              // 实例ID
	Instance    WorkflowInstance `json:"instance" gorm:"foreignKey:InstanceID"` // 实例信息
	NodeKey     string         `json:"node_key"`                                 // 节点标识
	NodeName    string         `json:"node_name"`                                // 节点名称
	Action      string         `json:"action"`                                   // 操作类型
	OperatorID  uint           `json:"operator_id"`                              // 操作人ID
	Operator    User           `json:"operator" gorm:"foreignKey:OperatorID"`    // 操作人信息
	Comment     string         `json:"comment"`                                  // 操作备注
	FormValues  string         `json:"form_values"`                              // 表单值(JSON)
	Variables   string         `json:"variables"`                                // 变量信息(JSON)
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// NodeTreeData 节点树结构
type NodeTreeData struct {
	Key      string          `json:"key"`
	Name     string          `json:"name"`
	Type     NodeType        `json:"type"`
	Child    *NodeTreeData   `json:"child,omitempty"`
	Branches []NodeTreeData  `json:"branches,omitempty"`
}

// ParseNodeTree 解析节点树结构
func (w *WorkflowDefinition) ParseNodeTree() (*NodeTreeData, error) {
	if w.NodeData == "" {
		return nil, nil
	}
	
	var nodeTree NodeTreeData
	err := json.Unmarshal([]byte(w.NodeData), &nodeTree)
	return &nodeTree, err
}

// SetNodeTree 设置节点树结构
func (w *WorkflowDefinition) SetNodeTree(nodeTree *NodeTreeData) error {
	data, err := json.Marshal(nodeTree)
	if err != nil {
		return err
	}
	w.NodeData = string(data)
	return nil
} 