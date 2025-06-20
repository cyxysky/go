package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gin-web-api/database"
	"gin-web-api/models"

	"gorm.io/gorm"
)

type FormService struct {
	db *gorm.DB
}

func NewFormService() *FormService {
	return &FormService{
		db: database.GetDB(),
	}
}

// CreateFormDefinition 创建表单定义
func (s *FormService) CreateFormDefinition(req *CreateFormRequest, creatorID uint) (*models.FormDefinition, error) {
	return s.db.Transaction(func(tx *gorm.DB) (*models.FormDefinition, error) {
		// 创建表单定义
		form := &models.FormDefinition{
			ObjectKey:   req.Object,
			Name:        req.Name,
			FormKey:     req.Key,
			Description: req.Description,
			Version:     1,
			IsActive:    true,
			CreatedBy:   creatorID,
		}

		if err := tx.Create(form).Error; err != nil {
			return nil, fmt.Errorf("创建表单定义失败: %w", err)
		}

		// 创建表单卡片和字段
		for cardIndex, cardReq := range req.Cards {
			card := &models.FormCard{
				FormID:    form.ID,
				Name:      cardReq.Name,
				SortOrder: cardIndex,
			}

			if err := tx.Create(card).Error; err != nil {
				return nil, fmt.Errorf("创建表单卡片失败: %w", err)
			}

			// 创建字段属性
			for attrIndex, attrReq := range cardReq.Attributes {
				// 先创建或获取字段属性定义
				var fieldAttr models.FieldAttribute
				err := tx.Where("object_key = ? AND field_key = ?", 
					attrReq.Attribute.Object, attrReq.Attribute.Key).First(&fieldAttr).Error
				
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 创建新的字段属性定义
					fieldAttr = models.FieldAttribute{
						ObjectKey:       attrReq.Attribute.Object,
						Name:           attrReq.Attribute.Name,
						FieldKey:       attrReq.Attribute.Key,
						DataType:       attrReq.Attribute.Type,
						Element:        attrReq.Attribute.Element,
						ParentObject:   attrReq.Attribute.ParentObject,
						JoinColumn:     attrReq.Attribute.JoinColumn,
						JoinColumnZh:   attrReq.Attribute.JoinColumnZh,
						Transfer:       attrReq.Attribute.Transfer,
					}
					
					if err := tx.Create(&fieldAttr).Error; err != nil {
						return nil, fmt.Errorf("创建字段属性定义失败: %w", err)
					}
				}

				// 创建表单属性
				formAttr := &models.FormAttribute{
					CardID:       card.ID,
					AttributeID:  fieldAttr.ID,
					Element:      attrReq.Element,
					Name:         attrReq.Name,
					Width:        attrReq.Width,
					Required:     attrReq.Required,
					Disable:      attrReq.Disable,
					Show:         attrReq.Show,
					Placeholder:  attrReq.Placeholder,
					LocationX:    attrReq.Location.X,
					LocationY:    attrReq.Location.Y,
				}

				// 设置默认值和选项
				if attrReq.DefaultValue != "" {
					formAttr.DefaultValue = attrReq.DefaultValue
				}
				if attrReq.Options != nil {
					optionsJson, _ := json.Marshal(attrReq.Options)
					formAttr.Options = string(optionsJson)
				}
				if attrReq.Validation != nil {
					validationJson, _ := json.Marshal(attrReq.Validation)
					formAttr.Validation = string(validationJson)
				}

				formAttr.LocationX = attrIndex + 1
				if err := tx.Create(formAttr).Error; err != nil {
					return nil, fmt.Errorf("创建表单属性失败: %w", err)
				}
			}
		}

		// 创建表单按钮
		for btnIndex, btnReq := range req.Buttons {
			button := &models.FormButton{
				FormID:       form.ID,
				Name:         btnReq.Name,
				ButtonType:   btnReq.Type,
				ClickOperate: btnReq.Click.Operate,
				SortOrder:    btnIndex,
			}

			if btnReq.ShowCondition != nil {
				conditionJson, _ := json.Marshal(btnReq.ShowCondition)
				button.ShowCondition = string(conditionJson)
			}

			if err := tx.Create(button).Error; err != nil {
				return nil, fmt.Errorf("创建表单按钮失败: %w", err)
			}
		}

		// 重新加载完整数据
		if err := tx.Preload("Cards.Attributes.Attribute").
			Preload("Buttons").First(form, form.ID).Error; err != nil {
			return nil, fmt.Errorf("加载表单数据失败: %w", err)
		}

		return form, nil
	})
}

// GetFormDefinition 获取表单定义
func (s *FormService) GetFormDefinition(formID uint) (*models.FormDefinition, error) {
	var form models.FormDefinition
	err := s.db.Preload("Cards.Attributes.Attribute").
		Preload("Buttons").
		Preload("Creator").
		First(&form, formID).Error
	
	if err != nil {
		return nil, fmt.Errorf("获取表单定义失败: %w", err)
	}
	
	return &form, nil
}

// GetFormDefinitionByKey 根据key获取表单定义
func (s *FormService) GetFormDefinitionByKey(formKey string) (*models.FormDefinition, error) {
	var form models.FormDefinition
	err := s.db.Preload("Cards.Attributes.Attribute").
		Preload("Buttons").
		Preload("Creator").
		Where("form_key = ? AND is_active = ?", formKey, true).
		First(&form).Error
	
	if err != nil {
		return nil, fmt.Errorf("获取表单定义失败: %w", err)
	}
	
	return &form, nil
}

// ListFormDefinitions 获取表单定义列表
func (s *FormService) ListFormDefinitions(page, pageSize int) ([]models.FormDefinition, int64, error) {
	var forms []models.FormDefinition
	var total int64

	// 计算总数
	s.db.Model(&models.FormDefinition{}).Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Preload("Creator").
		Limit(pageSize).Offset(offset).
		Order("created_at DESC").
		Find(&forms).Error

	return forms, total, err
}

// CreateFormData 创建表单数据
func (s *FormService) CreateFormData(req *CreateFormDataRequest, userID uint) (*models.FormData, error) {
	// 验证表单定义是否存在
	var form models.FormDefinition
	if err := s.db.First(&form, req.FormID).Error; err != nil {
		return nil, fmt.Errorf("表单定义不存在: %w", err)
	}

	// 验证表单数据
	if err := s.validateFormData(req.FormID, req.FormValues); err != nil {
		return nil, fmt.Errorf("表单数据验证失败: %w", err)
	}

	formData := &models.FormData{
		FormID:      req.FormID,
		InstanceID:  req.InstanceID,
		BusinessKey: req.BusinessKey,
		FormValues:  req.FormValues,
		Status:      models.FormStatusDraft,
		SubmittedBy: userID,
	}

	if err := s.db.Create(formData).Error; err != nil {
		return nil, fmt.Errorf("创建表单数据失败: %w", err)
	}

	return formData, nil
}

// UpdateFormData 更新表单数据
func (s *FormService) UpdateFormData(formDataID uint, req *UpdateFormDataRequest, userID uint) (*models.FormData, error) {
	var formData models.FormData
	if err := s.db.First(&formData, formDataID).Error; err != nil {
		return nil, fmt.Errorf("表单数据不存在: %w", err)
	}

	// 检查权限
	if formData.SubmittedBy != userID {
		return nil, errors.New("无权限修改此表单数据")
	}

	// 只有草稿状态才能修改
	if formData.Status != models.FormStatusDraft {
		return nil, errors.New("只有草稿状态的表单才能修改")
	}

	// 验证表单数据
	if req.FormValues != "" {
		if err := s.validateFormData(formData.FormID, req.FormValues); err != nil {
			return nil, fmt.Errorf("表单数据验证失败: %w", err)
		}
		formData.FormValues = req.FormValues
	}

	if req.Status != "" {
		formData.Status = req.Status
		if req.Status == models.FormStatusSubmitted {
			now := time.Now()
			formData.SubmittedAt = &now
		}
	}

	if err := s.db.Save(&formData).Error; err != nil {
		return nil, fmt.Errorf("更新表单数据失败: %w", err)
	}

	return &formData, nil
}

// GetFormData 获取表单数据
func (s *FormService) GetFormData(formDataID uint) (*models.FormData, error) {
	var formData models.FormData
	err := s.db.Preload("Form.Cards.Attributes.Attribute").
		Preload("Form.Buttons").
		Preload("Instance").
		Preload("Submitter").
		First(&formData, formDataID).Error
	
	if err != nil {
		return nil, fmt.Errorf("获取表单数据失败: %w", err)
	}
	
	return &formData, nil
}

// RenderFormWithData 渲染带数据的表单
func (s *FormService) RenderFormWithData(formID uint, formValues string) (*FormRenderData, error) {
	// 获取表单定义
	form, err := s.GetFormDefinition(formID)
	if err != nil {
		return nil, err
	}

	// 解析表单值
	var values map[string]interface{}
	if formValues != "" {
		if err := json.Unmarshal([]byte(formValues), &values); err != nil {
			return nil, fmt.Errorf("解析表单值失败: %w", err)
		}
	}

	// 构建渲染数据
	renderData := &FormRenderData{
		Form:   *form,
		Values: values,
	}

	return renderData, nil
}

// validateFormData 验证表单数据
func (s *FormService) validateFormData(formID uint, formValues string) error {
	// 获取表单定义
	form, err := s.GetFormDefinition(formID)
	if err != nil {
		return err
	}

	// 解析表单值
	var values map[string]interface{}
	if err := json.Unmarshal([]byte(formValues), &values); err != nil {
		return fmt.Errorf("表单数据格式错误: %w", err)
	}

	// 验证必填字段
	for _, card := range form.Cards {
		for _, attr := range card.Attributes {
			if attr.Required {
				fieldKey := attr.Attribute.FieldKey
				if _, exists := values[fieldKey]; !exists {
					return fmt.Errorf("必填字段 %s 不能为空", attr.Name)
				}
			}

			// 可以在这里添加更多的验证逻辑
			// 比如数据类型验证、格式验证等
		}
	}

	return nil
}

// 请求结构体
type CreateFormRequest struct {
	ID      int                `json:"id"`
	Object  string             `json:"object" binding:"required"`
	Name    string             `json:"name" binding:"required"`
	Key     string             `json:"key" binding:"required"`
	Cards   []CreateCardRequest `json:"cards" binding:"required"`
	Buttons []CreateButtonRequest `json:"buttons"`
}

type CreateCardRequest struct {
	Name       string                    `json:"name" binding:"required"`
	Attributes []CreateAttributeRequest  `json:"attributes" binding:"required"`
}

type CreateAttributeRequest struct {
	AttributeID  int                   `json:"attributeId"`
	Attribute    CreateFieldAttrRequest `json:"attribute" binding:"required"`
	Element      string                `json:"element" binding:"required"`
	Name         string                `json:"name" binding:"required"`
	Width        string                `json:"width"`
	Required     bool                  `json:"required"`
	Disable      bool                  `json:"disable"`
	Show         bool                  `json:"show"`
	Placeholder  string                `json:"placeholder"`
	Location     LocationRequest       `json:"location"`
	DefaultValue string                `json:"default_value"`
	Options      interface{}           `json:"options"`
	Validation   interface{}           `json:"validation"`
}

type CreateFieldAttrRequest struct {
	ID           int    `json:"id"`
	Object       string `json:"object" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Key          string `json:"key" binding:"required"`
	Type         string `json:"type" binding:"required"`
	Element      string `json:"element" binding:"required"`
	ParentObject string `json:"parentObject"`
	JoinColumn   string `json:"joinColumn"`
	JoinColumnZh string `json:"joinColumnZh"`
	Transfer     bool   `json:"transfer"`
}

type LocationRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type CreateButtonRequest struct {
	Name          string               `json:"name" binding:"required"`
	Type          string               `json:"type" binding:"required"`
	Click         ClickRequest         `json:"click" binding:"required"`
	ShowCondition *ShowConditionRequest `json:"showCondition"`
}

type ClickRequest struct {
	Operate string `json:"operate" binding:"required"`
}

type ShowConditionRequest struct {
	UUID   string                   `json:"uuid"`
	Object string                   `json:"object"`
	Groups []ConditionGroupRequest  `json:"groups"`
}

type ConditionGroupRequest struct {
	Conditions []ConditionRequest `json:"conditions"`
}

type ConditionRequest struct {
	Condition string      `json:"condition"`
	Keyword   string      `json:"keyword"`
	Value     interface{} `json:"value"`
}

type CreateFormDataRequest struct {
	FormID      uint   `json:"form_id" binding:"required"`
	InstanceID  uint   `json:"instance_id"`
	BusinessKey string `json:"business_key"`
	FormValues  string `json:"form_values" binding:"required"`
}

type UpdateFormDataRequest struct {
	FormValues string `json:"form_values"`
	Status     string `json:"status"`
}

type FormRenderData struct {
	Form   models.FormDefinition  `json:"form"`
	Values map[string]interface{} `json:"values"`
}

// UpdateFormDefinition 更新表单定义
func (s *FormService) UpdateFormDefinition(formID uint, req *CreateFormRequest, userID uint) (*models.FormDefinition, error) {
	return s.db.Transaction(func(tx *gorm.DB) (*models.FormDefinition, error) {
		// 获取现有表单定义
		var form models.FormDefinition
		if err := tx.Preload("Cards.Attributes").Preload("Buttons").First(&form, formID).Error; err != nil {
			return nil, fmt.Errorf("表单定义不存在: %w", err)
		}

		// 检查权限
		if form.CreatedBy != userID {
			return nil, errors.New("无权限修改此表单")
		}

		// 更新基本信息
		form.Name = req.Name
		form.Description = req.Description
		form.Version++ // 增加版本号

		// 删除现有的卡片和属性
		if err := tx.Where("form_id = ?", formID).Delete(&models.FormCard{}).Error; err != nil {
			return nil, fmt.Errorf("删除现有卡片失败: %w", err)
		}
		if err := tx.Where("form_id = ?", formID).Delete(&models.FormButton{}).Error; err != nil {
			return nil, fmt.Errorf("删除现有按钮失败: %w", err)
		}

		// 重新创建卡片和字段
		for cardIndex, cardReq := range req.Cards {
			card := &models.FormCard{
				FormID:    form.ID,
				Name:      cardReq.Name,
				SortOrder: cardIndex,
			}

			if err := tx.Create(card).Error; err != nil {
				return nil, fmt.Errorf("创建表单卡片失败: %w", err)
			}

			// 创建字段属性
			for attrIndex, attrReq := range cardReq.Attributes {
				// 先创建或获取字段属性定义
				var fieldAttr models.FieldAttribute
				err := tx.Where("object_key = ? AND field_key = ?", 
					attrReq.Attribute.Object, attrReq.Attribute.Key).First(&fieldAttr).Error
				
				if errors.Is(err, gorm.ErrRecordNotFound) {
					fieldAttr = models.FieldAttribute{
						ObjectKey:       attrReq.Attribute.Object,
						Name:           attrReq.Attribute.Name,
						FieldKey:       attrReq.Attribute.Key,
						DataType:       attrReq.Attribute.Type,
						Element:        attrReq.Attribute.Element,
						ParentObject:   attrReq.Attribute.ParentObject,
						JoinColumn:     attrReq.Attribute.JoinColumn,
						JoinColumnZh:   attrReq.Attribute.JoinColumnZh,
						Transfer:       attrReq.Attribute.Transfer,
					}
					
					if err := tx.Create(&fieldAttr).Error; err != nil {
						return nil, fmt.Errorf("创建字段属性定义失败: %w", err)
					}
				}

				// 创建表单属性
				formAttr := &models.FormAttribute{
					CardID:       card.ID,
					AttributeID:  fieldAttr.ID,
					Element:      attrReq.Element,
					Name:         attrReq.Name,
					Width:        attrReq.Width,
					Required:     attrReq.Required,
					Disable:      attrReq.Disable,
					Show:         attrReq.Show,
					Placeholder:  attrReq.Placeholder,
					LocationX:    attrIndex + 1,
					LocationY:    attrReq.Location.Y,
				}

				// 设置默认值和选项
				if attrReq.DefaultValue != "" {
					formAttr.DefaultValue = attrReq.DefaultValue
				}
				if attrReq.Options != nil {
					optionsJson, _ := json.Marshal(attrReq.Options)
					formAttr.Options = string(optionsJson)
				}
				if attrReq.Validation != nil {
					validationJson, _ := json.Marshal(attrReq.Validation)
					formAttr.Validation = string(validationJson)
				}

				if err := tx.Create(formAttr).Error; err != nil {
					return nil, fmt.Errorf("创建表单属性失败: %w", err)
				}
			}
		}

		// 重新创建表单按钮
		for btnIndex, btnReq := range req.Buttons {
			button := &models.FormButton{
				FormID:       form.ID,
				Name:         btnReq.Name,
				ButtonType:   btnReq.Type,
				ClickOperate: btnReq.Click.Operate,
				SortOrder:    btnIndex,
			}

			if btnReq.ShowCondition != nil {
				conditionJson, _ := json.Marshal(btnReq.ShowCondition)
				button.ShowCondition = string(conditionJson)
			}

			if err := tx.Create(button).Error; err != nil {
				return nil, fmt.Errorf("创建表单按钮失败: %w", err)
			}
		}

		// 保存表单
		if err := tx.Save(&form).Error; err != nil {
			return nil, fmt.Errorf("保存表单失败: %w", err)
		}

		// 重新加载完整数据
		if err := tx.Preload("Cards.Attributes.Attribute").
			Preload("Buttons").First(&form, form.ID).Error; err != nil {
			return nil, fmt.Errorf("加载表单数据失败: %w", err)
		}

		return &form, nil
	})
}

// DeleteFormDefinition 删除表单定义
func (s *FormService) DeleteFormDefinition(formID uint, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取表单定义
		var form models.FormDefinition
		if err := tx.First(&form, formID).Error; err != nil {
			return fmt.Errorf("表单定义不存在: %w", err)
		}

		// 检查权限
		if form.CreatedBy != userID {
			return errors.New("无权限删除此表单")
		}

		// 检查是否有工作流在使用
		var workflowCount int64
		if err := tx.Model(&models.WorkflowDefinition{}).Where("form_id = ?", formID).Count(&workflowCount).Error; err != nil {
			return fmt.Errorf("检查工作流使用情况失败: %w", err)
		}

		if workflowCount > 0 {
			return errors.New("表单正在被工作流使用，不能删除")
		}

		// 软删除表单定义（级联删除相关数据）
		if err := tx.Delete(&form).Error; err != nil {
			return fmt.Errorf("删除表单定义失败: %w", err)
		}

		return nil
	})
}

// ExportFormDefinition 导出表单定义
func (s *FormService) ExportFormDefinition(formID uint) (*CreateFormRequest, error) {
	form, err := s.GetFormDefinition(formID)
	if err != nil {
		return nil, err
	}

	// 转换为导出格式
	exportReq := &CreateFormRequest{
		ID:      int(form.ID),
		Object:  form.ObjectKey,
		Name:    form.Name,
		Key:     form.FormKey,
		Cards:   make([]CreateCardRequest, 0),
		Buttons: make([]CreateButtonRequest, 0),
	}

	// 转换卡片
	for _, card := range form.Cards {
		cardReq := CreateCardRequest{
			Name:       card.Name,
			Attributes: make([]CreateAttributeRequest, 0),
		}

		// 转换属性
		for _, attr := range card.Attributes {
			attrReq := CreateAttributeRequest{
				AttributeID: int(attr.AttributeID),
				Attribute: CreateFieldAttrRequest{
					ID:           int(attr.Attribute.ID),
					Object:       attr.Attribute.ObjectKey,
					Name:         attr.Attribute.Name,
					Key:          attr.Attribute.FieldKey,
					Type:         attr.Attribute.DataType,
					Element:      attr.Attribute.Element,
					ParentObject: attr.Attribute.ParentObject,
					JoinColumn:   attr.Attribute.JoinColumn,
					JoinColumnZh: attr.Attribute.JoinColumnZh,
					Transfer:     attr.Attribute.Transfer,
				},
				Element:      attr.Element,
				Name:         attr.Name,
				Width:        attr.Width,
				Required:     attr.Required,
				Disable:      attr.Disable,
				Show:         attr.Show,
				Placeholder:  attr.Placeholder,
				Location:     LocationRequest{X: attr.LocationX, Y: attr.LocationY},
				DefaultValue: attr.DefaultValue,
			}

			// 解析选项和验证规则
			if attr.Options != "" {
				json.Unmarshal([]byte(attr.Options), &attrReq.Options)
			}
			if attr.Validation != "" {
				json.Unmarshal([]byte(attr.Validation), &attrReq.Validation)
			}

			cardReq.Attributes = append(cardReq.Attributes, attrReq)
		}

		exportReq.Cards = append(exportReq.Cards, cardReq)
	}

	// 转换按钮
	for _, btn := range form.Buttons {
		btnReq := CreateButtonRequest{
			Name: btn.Name,
			Type: btn.ButtonType,
			Click: ClickRequest{
				Operate: btn.ClickOperate,
			},
		}

		// 解析显示条件
		if btn.ShowCondition != "" {
			var showCondition ShowConditionRequest
			if json.Unmarshal([]byte(btn.ShowCondition), &showCondition) == nil {
				btnReq.ShowCondition = &showCondition
			}
		}

		exportReq.Buttons = append(exportReq.Buttons, btnReq)
	}

	return exportReq, nil
}

// CloneFormDefinition 克隆表单定义
func (s *FormService) CloneFormDefinition(formID uint, newName string, userID uint) (*models.FormDefinition, error) {
	// 导出现有表单定义
	exportReq, err := s.ExportFormDefinition(formID)
	if err != nil {
		return nil, fmt.Errorf("导出表单定义失败: %w", err)
	}

	// 修改名称和key
	exportReq.Name = newName
	exportReq.Key = fmt.Sprintf("%s_clone_%d", exportReq.Key, time.Now().Unix())

	// 创建新的表单定义
	return s.CreateFormDefinition(exportReq, userID)
}

// ActivateFormDefinition 激活表单定义
func (s *FormService) ActivateFormDefinition(formID uint, userID uint) error {
	var form models.FormDefinition
	if err := s.db.First(&form, formID).Error; err != nil {
		return fmt.Errorf("表单定义不存在: %w", err)
	}

	// 检查权限
	if form.CreatedBy != userID {
		return errors.New("无权限修改此表单")
	}

	form.IsActive = true
	return s.db.Save(&form).Error
}

// DeactivateFormDefinition 停用表单定义
func (s *FormService) DeactivateFormDefinition(formID uint, userID uint) error {
	var form models.FormDefinition
	if err := s.db.First(&form, formID).Error; err != nil {
		return fmt.Errorf("表单定义不存在: %w", err)
	}

	// 检查权限
	if form.CreatedBy != userID {
		return errors.New("无权限修改此表单")
	}

	form.IsActive = false
	return s.db.Save(&form).Error
} 