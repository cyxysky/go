package models

import (
	"time"

	"gorm.io/gorm"
)

// FormDefinition 表单定义
type FormDefinition struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ObjectKey   string         `json:"object" gorm:"not null"`                  // 对象标识
	Name        string         `json:"name" gorm:"not null"`                    // 表单名称
	FormKey     string         `json:"key" gorm:"uniqueIndex;not null"`         // 表单唯一标识
	Description string         `json:"description"`                             // 表单描述
	Version     int            `json:"version" gorm:"default:1"`                // 版本号
	IsActive    bool           `json:"is_active" gorm:"default:true"`           // 是否启用
	CreatedBy   uint           `json:"created_by"`                              // 创建者
	Creator     User           `json:"creator" gorm:"foreignKey:CreatedBy"`     // 创建者信息
	Cards       []FormCard     `json:"cards" gorm:"foreignKey:FormID;constraint:OnDelete:CASCADE"`          // 表单卡片
	Buttons     []FormButton   `json:"buttons" gorm:"foreignKey:FormID;constraint:OnDelete:CASCADE"`        // 表单按钮
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// FormCard 表单卡片/分组
type FormCard struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	FormID     uint           `json:"form_id"`                              // 表单ID
	Name       string         `json:"name" gorm:"not null"`                 // 卡片名称
	SortOrder  int            `json:"sort_order" gorm:"default:0"`          // 排序
	Attributes []FormAttribute `json:"attributes" gorm:"foreignKey:CardID;constraint:OnDelete:CASCADE"` // 字段属性
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// FormAttribute 表单字段属性
type FormAttribute struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	CardID        uint           `json:"card_id"`                                    // 卡片ID
	AttributeID   uint           `json:"attribute_id"`                               // 属性定义ID
	Attribute     FieldAttribute `json:"attribute" gorm:"foreignKey:AttributeID"`   // 属性定义
	Element       string         `json:"element" gorm:"not null"`                   // 表单元素类型
	Name          string         `json:"name" gorm:"not null"`                      // 字段名称
	Width         string         `json:"width" gorm:"default:inline"`               // 宽度设置
	Required      bool           `json:"required" gorm:"default:false"`             // 是否必填
	Disable       bool           `json:"disable" gorm:"default:false"`              // 是否禁用
	Show          bool           `json:"show" gorm:"default:true"`                  // 是否显示
	Placeholder   string         `json:"placeholder"`                               // 占位符
	LocationX     int            `json:"location_x" gorm:"default:1"`               // X坐标
	LocationY     int            `json:"location_y" gorm:"default:1"`               // Y坐标
	DefaultValue  string         `json:"default_value"`                             // 默认值
	Options       string         `json:"options"`                                   // 选项配置(JSON)
	Validation    string         `json:"validation"`                                // 验证规则(JSON)
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// FieldAttribute 字段属性定义
type FieldAttribute struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ObjectKey       string         `json:"object" gorm:"not null"`              // 对象标识
	Name            string         `json:"name" gorm:"not null"`                // 属性名称
	FieldKey        string         `json:"key" gorm:"not null"`                 // 字段标识
	DataType        string         `json:"type" gorm:"not null"`                // 数据类型
	Element         string         `json:"element" gorm:"not null"`             // 元素类型
	ParentObject    string         `json:"parent_object"`                       // 父对象
	JoinColumn      string         `json:"join_column"`                         // 关联字段
	JoinColumnZh    string         `json:"join_column_zh"`                      // 关联显示字段
	Transfer        bool           `json:"transfer" gorm:"default:false"`       // 是否传递
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// FormButton 表单按钮
type FormButton struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	FormID        uint           `json:"form_id"`                          // 表单ID
	Name          string         `json:"name" gorm:"not null"`             // 按钮名称
	ButtonType    string         `json:"type" gorm:"not null"`             // 按钮类型 primary/secondary
	ClickOperate  string         `json:"click_operate" gorm:"not null"`    // 点击操作
	ShowCondition string         `json:"show_condition"`                   // 显示条件(JSON)
	SortOrder     int            `json:"sort_order" gorm:"default:0"`      // 排序
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// FormData 表单数据实例
type FormData struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	FormID       uint           `json:"form_id"`                                // 表单定义ID
	Form         FormDefinition `json:"form" gorm:"foreignKey:FormID"`          // 表单定义
	InstanceID   uint           `json:"instance_id"`                            // 工作流实例ID
	Instance     WorkflowInstance `json:"instance" gorm:"foreignKey:InstanceID"` // 工作流实例
	BusinessKey  string         `json:"business_key"`                           // 业务标识
	FormValues   string         `json:"form_values"`                            // 表单值(JSON)
	Status       string         `json:"status" gorm:"default:draft"`            // 状态
	SubmittedBy  uint           `json:"submitted_by"`                           // 提交人ID
	Submitter    User           `json:"submitter" gorm:"foreignKey:SubmittedBy"` // 提交人信息
	SubmittedAt  *time.Time     `json:"submitted_at"`                           // 提交时间
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// 表单元素类型常量
const (
	ElementTypeInput      = "input"       // 文本输入
	ElementTypeNumber     = "number"      // 数字输入
	ElementTypeSelect     = "select"      // 下拉选择
	ElementTypeCheckbox   = "checkbox"    // 多选框
	ElementTypeRadio      = "radio"       // 单选框
	ElementTypeTextarea   = "text"        // 多行文本
	ElementTypeDate       = "date"        // 日期选择
	ElementTypeDatetime   = "datetime"    // 日期时间
	ElementTypeFile       = "file"        // 文件上传
	ElementTypeCascader   = "cascader"    // 级联选择
	ElementTypeTreeSelect = "treeSelect"  // 树形选择
)

// 数据类型常量
const (
	DataTypeString  = "STRING"
	DataTypeNumber  = "NUMERIC"
	DataTypeInteger = "NUMBER"
	DataTypeBoolean = "BOOLEAN"
	DataTypeDate    = "DATE"
	DataTypeJSON    = "JSONB"
)

// 按钮操作类型常量
const (
	ButtonOperateCancel  = "cancel"  // 取消
	ButtonOperateDraft   = "draft"   // 保存草稿
	ButtonOperateStart   = "start"   // 发起申请
	ButtonOperatePass    = "pass"    // 通过
	ButtonOperateReject  = "reject"  // 拒绝
	ButtonOperateRecall  = "recall"  // 撤回
	ButtonOperateSubmit  = "submit"  // 提交
)

// 表单状态常量
const (
	FormStatusDraft     = "draft"     // 草稿
	FormStatusSubmitted = "submitted" // 已提交
	FormStatusApproved  = "approved"  // 已通过
	FormStatusRejected  = "rejected"  // 已拒绝
) 