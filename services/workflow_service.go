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

type WorkflowService struct {
	db          *gorm.DB
	formService *FormService
}

func NewWorkflowService() *WorkflowService {
	return &WorkflowService{
		db:          database.GetDB(),
		formService: NewFormService(),
	}
}

// CreateWorkflowWithNodeTree 根据节点树创建工作流定义
func (s *WorkflowService) CreateWorkflowWithNodeTree(req *CreateWorkflowWithTreeRequest, creatorID uint) (*models.WorkflowDefinition, error) {
	return s.db.Transaction(func(tx *gorm.DB) (*models.WorkflowDefinition, error) {
		// 创建工作流定义
		workflow := &models.WorkflowDefinition{
			Name:        req.Name,
			Description: req.Description,
			Category:    req.Category,
			Version:     1,
			Status:      models.WorkflowStatusDraft,
			CreatedBy:   creatorID,
		}

		// 关联表单
		if req.FormID != 0 {
			workflow.FormID = &req.FormID
		}

		if err := tx.Create(workflow).Error; err != nil {
			return nil, fmt.Errorf("创建工作流失败: %w", err)
		}

		// 设置节点树
		if req.NodeTree != nil {
			if err := workflow.SetNodeTree(req.NodeTree); err != nil {
				return nil, fmt.Errorf("设置节点树失败: %w", err)
			}
		}

		// 从节点树创建节点记录
		if req.NodeTree != nil {
			rootNode, err := s.createNodesFromTree(tx, workflow.ID, req.NodeTree, nil)
			if err != nil {
				return nil, fmt.Errorf("创建节点失败: %w", err)
			}
			workflow.RootNodeID = &rootNode.ID
		}

		// 保存工作流
		if err := tx.Save(workflow).Error; err != nil {
			return nil, fmt.Errorf("保存工作流失败: %w", err)
		}

		return workflow, nil
	})
}

// createNodesFromTree 从节点树创建节点记录
func (s *WorkflowService) createNodesFromTree(tx *gorm.DB, workflowID uint, nodeTree *models.NodeTreeData, parentNode *models.WorkflowNode) (*models.WorkflowNode, error) {
	// 创建当前节点
	node := &models.WorkflowNode{
		WorkflowID: workflowID,
		NodeKey:    nodeTree.Key,
		Name:       nodeTree.Name,
		Type:       nodeTree.Type,
	}

	if parentNode != nil {
		node.ParentNodeID = &parentNode.ID
	}

	if err := tx.Create(node).Error; err != nil {
		return nil, err
	}

	// 创建子节点
	if nodeTree.Child != nil {
		childNode, err := s.createNodesFromTree(tx, workflowID, nodeTree.Child, node)
		if err != nil {
			return nil, err
		}
		node.ChildNodeID = &childNode.ID
		tx.Save(node)
	}

	// 创建分支节点
	for i, branch := range nodeTree.Branches {
		branchRecord := &models.WorkflowBranch{
			NodeID:    node.ID,
			BranchKey: branch.Key,
			Name:      branch.Name,
			Type:      branch.Type,
			SortOrder: i,
		}

		if err := tx.Create(branchRecord).Error; err != nil {
			return nil, err
		}

		// 递归创建分支子节点
		if branch.Child != nil {
			childNode, err := s.createNodesFromTree(tx, workflowID, branch.Child, node)
			if err != nil {
				return nil, err
			}
			branchRecord.ChildNodeID = &childNode.ID
			tx.Save(branchRecord)
		}
	}

	return node, nil
}

// CreateWorkflow 创建工作流定义
func (s *WorkflowService) CreateWorkflow(req *CreateWorkflowRequest, creatorID uint) (*models.WorkflowDefinition, error) {
	workflow := &models.WorkflowDefinition{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Version:     1,
		Status:      models.WorkflowStatusDraft,
		CreatedBy:   creatorID,
	}

	if err := s.db.Create(workflow).Error; err != nil {
		return nil, fmt.Errorf("创建工作流失败: %w", err)
	}

	// 创建节点
	for _, nodeReq := range req.Nodes {
		node := &models.WorkflowNode{
			WorkflowID:     workflow.ID,
			NodeKey:        nodeReq.NodeKey,
			Name:           nodeReq.Name,
			Type:           nodeReq.Type,
			ApprovalMode:   nodeReq.ApprovalMode,
			Position:       nodeReq.Position,
			Assignees:      nodeReq.Assignees,
			Conditions:     nodeReq.Conditions,
			Settings:       nodeReq.Settings,
			NextNodes:      nodeReq.NextNodes,
			FormConditions: nodeReq.FormConditions,
			SortOrder:      nodeReq.SortOrder,
		}
		if err := s.db.Create(node).Error; err != nil {
			return nil, fmt.Errorf("创建工作流节点失败: %w", err)
		}
	}

	return workflow, nil
}

// StartWorkflowWithForm 启动带表单的工作流实例
func (s *WorkflowService) StartWorkflowWithForm(req *StartWorkflowWithFormRequest, initiatorID uint) (*models.WorkflowInstance, error) {
	return s.db.Transaction(func(tx *gorm.DB) (*models.WorkflowInstance, error) {
		// 获取工作流定义
		var workflow models.WorkflowDefinition
		if err := tx.Preload("Nodes").Preload("Form").First(&workflow, req.WorkflowID).Error; err != nil {
			return nil, fmt.Errorf("工作流定义不存在: %w", err)
		}

		if workflow.Status != models.WorkflowStatusActive {
			return nil, errors.New("工作流未激活，无法启动")
		}

		// 如果有表单数据，先创建表单数据
		var formData *models.FormData
		if req.FormValues != "" && workflow.FormID != nil {
			formDataReq := &CreateFormDataRequest{
				FormID:      *workflow.FormID,
				BusinessKey: req.BusinessKey,
				FormValues:  req.FormValues,
			}
			
			var err error
			formData, err = s.formService.CreateFormData(formDataReq, initiatorID)
			if err != nil {
				return nil, fmt.Errorf("创建表单数据失败: %w", err)
			}
		}

		// 创建工作流实例
		instance := &models.WorkflowInstance{
			WorkflowID:   req.WorkflowID,
			Title:        req.Title,
			BusinessKey:  req.BusinessKey,
			BusinessType: req.BusinessType,
			BusinessData: req.BusinessData,
			Status:       models.InstanceStatusRunning,
			StartTime:    time.Now(),
			InitiatorID:  initiatorID,
		}

		if formData != nil {
			instance.FormDataID = &formData.ID
		}

		// 初始化流程变量
		variables := make(map[string]interface{})
		if req.Variables != "" {
			json.Unmarshal([]byte(req.Variables), &variables)
		}
		
		variablesJson, _ := json.Marshal(variables)
		instance.Variables = string(variablesJson)

		if err := tx.Create(instance).Error; err != nil {
			return nil, fmt.Errorf("创建工作流实例失败: %w", err)
		}

		// 执行流程引擎，开始第一个节点
		if err := s.executeWorkflowWithTree(tx, instance, workflow); err != nil {
			return nil, fmt.Errorf("启动工作流失败: %w", err)
		}

		// 记录历史
		s.recordHistoryInTx(tx, instance.ID, "", "开始", initiatorID, "工作流已启动", req.FormValues, "")

		return instance, nil
	})
}

// executeWorkflowWithTree 执行基于节点树的工作流
func (s *WorkflowService) executeWorkflowWithTree(tx *gorm.DB, instance *models.WorkflowInstance, workflow models.WorkflowDefinition) error {
	// 解析节点树
	nodeTree, err := workflow.ParseNodeTree()
	if err != nil || nodeTree == nil {
		// 回退到传统节点执行
		return s.executeWorkflow(instance, workflow.Nodes, "")
	}

	// 从根节点开始执行
	return s.executeNodeTree(tx, instance, nodeTree, nil)
}

// executeNodeTree 执行节点树
func (s *WorkflowService) executeNodeTree(tx *gorm.DB, instance *models.WorkflowInstance, nodeTree *models.NodeTreeData, variables map[string]interface{}) error {
	switch nodeTree.Type {
	case models.NodeTypeRoot, models.NodeTypeStart:
		// 根节点或开始节点，执行子节点
		if nodeTree.Child != nil {
			return s.executeNodeTree(tx, instance, nodeTree.Child, variables)
		}
		
	case models.NodeTypeEnd:
		// 结束节点，完成工作流
		return s.completeWorkflowInTx(tx, instance, models.InstanceStatusApproved)
		
	case models.NodeTypeApproval:
		// 审批节点，创建任务
		return s.createApprovalTasksFromNode(tx, instance, nodeTree)
		
	case models.NodeTypeCondition:
		// 条件节点，评估分支
		return s.evaluateConditionBranches(tx, instance, nodeTree, variables)
	}

	return nil
}

// evaluateConditionBranches 评估条件分支
func (s *WorkflowService) evaluateConditionBranches(tx *gorm.DB, instance *models.WorkflowInstance, nodeTree *models.NodeTreeData, variables map[string]interface{}) error {
	// 获取表单数据进行条件评估
	var formValues map[string]interface{}
	if instance.FormDataID != nil {
		var formData models.FormData
		if err := tx.First(&formData, *instance.FormDataID).Error; err == nil {
			json.Unmarshal([]byte(formData.FormValues), &formValues)
		}
	}

	// 评估每个分支条件
	for _, branch := range nodeTree.Branches {
		if s.evaluateBranchCondition(&branch, formValues, variables) {
			// 条件满足，执行该分支
			if branch.Child != nil {
				return s.executeNodeTree(tx, instance, branch.Child, variables)
			}
		}
	}

	// 如果没有分支满足条件，执行默认子节点
	if nodeTree.Child != nil {
		return s.executeNodeTree(tx, instance, nodeTree.Child, variables)
	}

	return nil
}

// evaluateBranchCondition 评估分支条件
func (s *WorkflowService) evaluateBranchCondition(branch *models.NodeTreeData, formValues, variables map[string]interface{}) bool {
	// 解析条件配置
	if branch.Type != models.NodeTypeCondition {
		return true // 非条件分支默认为true
	}

	// 这里可以实现复杂的条件评估逻辑
	// 暂时返回true，表示所有条件都满足
	return true
}

// createApprovalTasksFromNode 从节点创建审批任务
func (s *WorkflowService) createApprovalTasksFromNode(tx *gorm.DB, instance *models.WorkflowInstance, nodeTree *models.NodeTreeData) error {
	// 获取节点详细信息
	var node models.WorkflowNode
	if err := tx.Where("workflow_id = ? AND node_key = ?", instance.WorkflowID, nodeTree.Key).First(&node).Error; err != nil {
		return fmt.Errorf("找不到节点配置: %w", err)
	}

	// 解析审批人配置
	var assigneeConfig AssigneeConfig
	if node.Assignees != "" {
		if err := json.Unmarshal([]byte(node.Assignees), &assigneeConfig); err != nil {
			return fmt.Errorf("解析审批人配置失败: %w", err)
		}
	}

	// 获取审批人列表
	assignees, err := s.resolveAssignees(assigneeConfig, instance)
	if err != nil {
		return fmt.Errorf("获取审批人失败: %w", err)
	}

	if len(assignees) == 0 {
		return errors.New("未找到有效的审批人")
	}

	// 根据审批模式创建任务
	switch node.ApprovalMode {
	case models.ApprovalModeSequence:
		// 依次审批，只创建第一个人的任务
		return s.createSingleTaskInTx(tx, instance, node, assignees[0])
		
	case models.ApprovalModeParallel, models.ApprovalModeAny, models.ApprovalModeAll:
		// 并行审批，为所有人创建任务
		for _, assigneeID := range assignees {
			if err := s.createSingleTaskInTx(tx, instance, node, assigneeID); err != nil {
				return err
			}
		}
	}

	// 更新实例当前节点
	currentNodes := []string{node.NodeKey}
	currentNodesJson, _ := json.Marshal(currentNodes)
	instance.CurrentNodes = string(currentNodesJson)
	
	return tx.Save(instance).Error
}

// StartWorkflow 启动工作流实例
func (s *WorkflowService) StartWorkflow(req *StartWorkflowRequest, initiatorID uint) (*models.WorkflowInstance, error) {
	// 获取工作流定义
	var workflow models.WorkflowDefinition
	if err := s.db.Preload("Nodes").First(&workflow, req.WorkflowID).Error; err != nil {
		return nil, fmt.Errorf("工作流定义不存在: %w", err)
	}

	if workflow.Status != models.WorkflowStatusActive {
		return nil, errors.New("工作流未激活，无法启动")
	}

	// 创建工作流实例
	instance := &models.WorkflowInstance{
		WorkflowID:   req.WorkflowID,
		Title:        req.Title,
		BusinessKey:  req.BusinessKey,
		BusinessType: req.BusinessType,
		BusinessData: req.BusinessData,
		Status:       models.InstanceStatusRunning,
		StartTime:    time.Now(),
		InitiatorID:  initiatorID,
	}

	if err := s.db.Create(instance).Error; err != nil {
		return nil, fmt.Errorf("创建工作流实例失败: %w", err)
	}

	// 执行流程引擎，开始第一个节点
	if err := s.executeWorkflow(instance, workflow.Nodes, ""); err != nil {
		return nil, fmt.Errorf("启动工作流失败: %w", err)
	}

	// 记录历史
	s.recordHistory(instance.ID, "", "开始", initiatorID, "工作流已启动", "", "")

	return instance, nil
}

// executeWorkflow 执行工作流引擎
func (s *WorkflowService) executeWorkflow(instance *models.WorkflowInstance, nodes []models.WorkflowNode, fromNodeKey string) error {
	var nextNodes []models.WorkflowNode
	
	if fromNodeKey == "" {
		// 查找开始节点
		for _, node := range nodes {
			if node.Type == models.NodeTypeStart {
				nextNodes = append(nextNodes, node)
				break
			}
		}
	} else {
		// 查找当前节点的下一个节点
		for _, node := range nodes {
			if node.NodeKey == fromNodeKey {
				var nextNodeKeys []string
				if node.NextNodes != "" {
					json.Unmarshal([]byte(node.NextNodes), &nextNodeKeys)
				}
				
				for _, nextKey := range nextNodeKeys {
					for _, nextNode := range nodes {
						if nextNode.NodeKey == nextKey {
							// 检查条件是否满足
							if s.checkNodeCondition(nextNode, instance) {
								nextNodes = append(nextNodes, nextNode)
							}
						}
					}
				}
				break
			}
		}
	}

	// 处理下一个节点
	for _, nextNode := range nextNodes {
		if err := s.processNode(instance, nextNode); err != nil {
			return err
		}
	}

	return nil
}

// processNode 处理单个节点
func (s *WorkflowService) processNode(instance *models.WorkflowInstance, node models.WorkflowNode) error {
	switch node.Type {
	case models.NodeTypeStart:
		// 开始节点，直接执行下一个节点
		return s.executeWorkflow(instance, []models.WorkflowNode{node}, node.NodeKey)
		
	case models.NodeTypeEnd:
		// 结束节点，完成工作流
		return s.completeWorkflow(instance, models.InstanceStatusApproved)
		
	case models.NodeTypeApproval:
		// 审批节点，创建任务
		return s.createApprovalTasks(instance, node)
		
	case models.NodeTypeCondition:
		// 条件节点，直接执行下一个节点
		return s.executeWorkflow(instance, []models.WorkflowNode{node}, node.NodeKey)
		
	default:
		return fmt.Errorf("不支持的节点类型: %s", node.Type)
	}
}

// createApprovalTasks 创建审批任务
func (s *WorkflowService) createApprovalTasks(instance *models.WorkflowInstance, node models.WorkflowNode) error {
	// 解析审批人配置
	var assigneeConfig AssigneeConfig
	if node.Assignees != "" {
		if err := json.Unmarshal([]byte(node.Assignees), &assigneeConfig); err != nil {
			return fmt.Errorf("解析审批人配置失败: %w", err)
		}
	}

	// 获取审批人列表
	assignees, err := s.resolveAssignees(assigneeConfig, instance)
	if err != nil {
		return fmt.Errorf("获取审批人失败: %w", err)
	}

	if len(assignees) == 0 {
		return errors.New("未找到有效的审批人")
	}

	// 根据审批模式创建任务
	switch node.ApprovalMode {
	case models.ApprovalModeSequence:
		// 依次审批，只创建第一个人的任务
		return s.createSingleTask(instance, node, assignees[0])
		
	case models.ApprovalModeParallel, models.ApprovalModeAny, models.ApprovalModeAll:
		// 并行审批，为所有人创建任务
		for _, assigneeID := range assignees {
			if err := s.createSingleTask(instance, node, assigneeID); err != nil {
				return err
			}
		}
		
	default:
		return fmt.Errorf("不支持的审批模式: %s", node.ApprovalMode)
	}

	// 更新实例当前节点
	currentNodes := []string{node.NodeKey}
	currentNodesJson, _ := json.Marshal(currentNodes)
	instance.CurrentNodes = string(currentNodesJson)
	
	return s.db.Save(instance).Error
}

// createSingleTask 创建单个任务
func (s *WorkflowService) createSingleTask(instance *models.WorkflowInstance, node models.WorkflowNode, assigneeID uint) error {
	task := &models.WorkflowTask{
		InstanceID: instance.ID,
		NodeKey:    node.NodeKey,
		NodeName:   node.Name,
		AssigneeID: assigneeID,
		Status:     models.TaskStatusPending,
	}

	return s.db.Create(task).Error
}

// createSingleTaskInTx 在事务中创建单个任务
func (s *WorkflowService) createSingleTaskInTx(tx *gorm.DB, instance *models.WorkflowInstance, node models.WorkflowNode, assigneeID uint) error {
	task := &models.WorkflowTask{
		InstanceID: instance.ID,
		NodeKey:    node.NodeKey,
		NodeName:   node.Name,
		AssigneeID: assigneeID,
		Status:     models.TaskStatusPending,
	}

	return tx.Create(task).Error
}

// ApproveTask 审批任务
func (s *WorkflowService) ApproveTask(taskID uint, userID uint, comment string) error {
	return s.ApproveTaskWithForm(taskID, userID, comment, "")
}

// ApproveTaskWithForm 带表单数据的审批任务
func (s *WorkflowService) ApproveTaskWithForm(taskID uint, userID uint, comment, formValues string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取任务信息
		var task models.WorkflowTask
		if err := tx.Preload("Instance.Workflow.Nodes").First(&task, taskID).Error; err != nil {
			return fmt.Errorf("任务不存在: %w", err)
		}

		// 检查权限
		if task.AssigneeID != userID {
			return errors.New("无权限处理此任务")
		}

		if task.Status != models.TaskStatusPending {
			return errors.New("任务已处理")
		}

		// 更新任务状态
		task.Status = models.TaskStatusApproved
		task.Comment = comment
		task.FormValues = formValues
		now := time.Now()
		task.ProcessTime = &now

		if err := tx.Save(&task).Error; err != nil {
			return fmt.Errorf("更新任务失败: %w", err)
		}

		// 记录历史
		s.recordHistoryInTx(tx, task.InstanceID, task.NodeKey, "审批通过", userID, comment, formValues, "")

		// 检查节点是否完成
		return s.checkNodeCompletionInTx(tx, &task)
	})
}

// RejectTask 拒绝任务
func (s *WorkflowService) RejectTask(taskID uint, userID uint, comment string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取任务信息
		var task models.WorkflowTask
		if err := tx.Preload("Instance").First(&task, taskID).Error; err != nil {
			return fmt.Errorf("任务不存在: %w", err)
		}

		// 检查权限
		if task.AssigneeID != userID {
			return errors.New("无权限处理此任务")
		}

		if task.Status != models.TaskStatusPending {
			return errors.New("任务已处理")
		}

		// 更新任务状态
		task.Status = models.TaskStatusRejected
		task.Comment = comment
		now := time.Now()
		task.ProcessTime = &now

		if err := tx.Save(&task).Error; err != nil {
			return fmt.Errorf("更新任务失败: %w", err)
		}

		// 记录历史
		s.recordHistoryInTx(tx, task.InstanceID, task.NodeKey, "拒绝", userID, comment, "", "")

		// 拒绝整个工作流实例
		return s.completeWorkflowInTx(tx, &task.Instance, models.InstanceStatusRejected)
	})
}

// checkNodeCompletion 检查节点是否完成
func (s *WorkflowService) checkNodeCompletion(task *models.WorkflowTask) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.checkNodeCompletionInTx(tx, task)
	})
}

// checkNodeCompletionInTx 在事务中检查节点是否完成
func (s *WorkflowService) checkNodeCompletionInTx(tx *gorm.DB, task *models.WorkflowTask) error {
	// 获取当前节点的所有任务
	var allTasks []models.WorkflowTask
	if err := tx.Where("instance_id = ? AND node_key = ?", task.InstanceID, task.NodeKey).Find(&allTasks).Error; err != nil {
		return err
	}

	// 获取节点信息
	var node models.WorkflowNode
	if err := tx.Where("workflow_id = ? AND node_key = ?", task.Instance.WorkflowID, task.NodeKey).First(&node).Error; err != nil {
		return err
	}

	// 根据审批模式判断是否完成
	var completed bool
	switch node.ApprovalMode {
	case models.ApprovalModeSequence:
		completed = s.checkSequenceCompletion(allTasks)
	case models.ApprovalModeAny:
		completed = s.checkAnyCompletion(allTasks)
	case models.ApprovalModeAll:
		completed = s.checkAllCompletion(allTasks)
	case models.ApprovalModeParallel:
		completed = s.checkAllCompletion(allTasks)
	}

	if completed {
		// 节点完成，检查是否使用节点树
		var workflow models.WorkflowDefinition
		if err := tx.First(&workflow, task.Instance.WorkflowID).Error; err != nil {
			return err
		}

		if workflow.NodeData != "" {
			// 使用节点树执行
			return s.continueWorkflowWithTree(tx, &task.Instance, &workflow, task.NodeKey)
		} else {
			// 使用传统节点执行
			var nodes []models.WorkflowNode
			if err := tx.Where("workflow_id = ?", task.Instance.WorkflowID).Find(&nodes).Error; err != nil {
				return err
			}
			return s.executeWorkflow(&task.Instance, nodes, task.NodeKey)
		}
	}

	return nil
}

// continueWorkflowWithTree 继续执行基于节点树的工作流
func (s *WorkflowService) continueWorkflowWithTree(tx *gorm.DB, instance *models.WorkflowInstance, workflow *models.WorkflowDefinition, completedNodeKey string) error {
	// 解析节点树
	nodeTree, err := workflow.ParseNodeTree()
	if err != nil {
		return err
	}

	// 找到已完成的节点，继续执行下一个节点
	nextNode := s.findNextNodeInTree(nodeTree, completedNodeKey)
	if nextNode != nil {
		return s.executeNodeTree(tx, instance, nextNode, nil)
	}

	return nil
}

// findNextNodeInTree 在节点树中查找下一个节点
func (s *WorkflowService) findNextNodeInTree(nodeTree *models.NodeTreeData, completedNodeKey string) *models.NodeTreeData {
	if nodeTree.Key == completedNodeKey {
		return nodeTree.Child
	}

	// 递归查找子节点
	if nodeTree.Child != nil {
		if result := s.findNextNodeInTree(nodeTree.Child, completedNodeKey); result != nil {
			return result
		}
	}

	// 递归查找分支节点
	for _, branch := range nodeTree.Branches {
		if result := s.findNextNodeInTree(&branch, completedNodeKey); result != nil {
			return result
		}
	}

	return nil
}

// 辅助方法
func (s *WorkflowService) checkSequenceCompletion(tasks []models.WorkflowTask) bool {
	// 依次审批：检查是否有审批通过的任务，如果有拒绝则整个流程拒绝
	for _, task := range tasks {
		if task.Status == models.TaskStatusRejected {
			return true // 有拒绝则完成（但是拒绝状态）
		}
		if task.Status == models.TaskStatusApproved {
			return true // 有通过则完成
		}
	}
	return false
}

func (s *WorkflowService) checkAnyCompletion(tasks []models.WorkflowTask) bool {
	// 任意审批：任何一个人审批通过即可
	for _, task := range tasks {
		if task.Status == models.TaskStatusApproved {
			return true
		}
	}
	return false
}

func (s *WorkflowService) checkAllCompletion(tasks []models.WorkflowTask) bool {
	// 全员审批：所有人都必须审批通过
	for _, task := range tasks {
		if task.Status == models.TaskStatusPending {
			return false
		}
		if task.Status == models.TaskStatusRejected {
			return true // 有拒绝则完成（但是拒绝状态）
		}
	}
	return true
}

func (s *WorkflowService) completeWorkflow(instance *models.WorkflowInstance, status models.InstanceStatus) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.completeWorkflowInTx(tx, instance, status)
	})
}

func (s *WorkflowService) completeWorkflowInTx(tx *gorm.DB, instance *models.WorkflowInstance, status models.InstanceStatus) error {
	instance.Status = status
	now := time.Now()
	instance.EndTime = &now
	return tx.Save(instance).Error
}

func (s *WorkflowService) checkNodeCondition(node models.WorkflowNode, instance *models.WorkflowInstance) bool {
	// 这里可以实现复杂的条件判断逻辑
	// 可以根据表单数据、流程变量等进行判断
	return true
}

func (s *WorkflowService) resolveAssignees(config AssigneeConfig, instance *models.WorkflowInstance) ([]uint, error) {
	var assignees []uint

	// 根据配置类型解析审批人
	switch config.Type {
	case "users":
		assignees = config.UserIDs
	case "roles":
		// 获取角色下的所有用户
		var users []models.User
		if err := s.db.Table("users").
			Joins("JOIN user_roles ON users.id = user_roles.user_id").
			Where("user_roles.role_id IN ?", config.RoleIDs).
			Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			assignees = append(assignees, user.ID)
		}
	case "departments":
		// 获取部门下的所有用户
		var users []models.User
		if err := s.db.Table("users").
			Joins("JOIN user_departments ON users.id = user_departments.user_id").
			Where("user_departments.department_id IN ?", config.DepartmentIDs).
			Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			assignees = append(assignees, user.ID)
		}
	}

	return assignees, nil
}

func (s *WorkflowService) recordHistory(instanceID uint, nodeKey, action string, operatorID uint, comment, formValues, variables string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.recordHistoryInTx(tx, instanceID, nodeKey, action, operatorID, comment, formValues, variables)
	})
}

func (s *WorkflowService) recordHistoryInTx(tx *gorm.DB, instanceID uint, nodeKey, action string, operatorID uint, comment, formValues, variables string) error {
	history := &models.WorkflowHistory{
		InstanceID: instanceID,
		NodeKey:    nodeKey,
		Action:     action,
		OperatorID: operatorID,
		Comment:    comment,
		FormValues: formValues,
		Variables:  variables,
	}
	return tx.Create(history).Error
}

// GetDB 获取数据库连接
func (s *WorkflowService) GetDB() *gorm.DB {
	return s.db
}

// 请求结构体
type CreateWorkflowRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Nodes       []CreateWorkflowNodeRequest `json:"nodes" binding:"required"`
}

type CreateWorkflowNodeRequest struct {
	NodeKey        string                `json:"node_key" binding:"required"`
	Name           string                `json:"name" binding:"required"`
	Type           models.NodeType       `json:"type" binding:"required"`
	ApprovalMode   models.ApprovalMode   `json:"approval_mode"`
	Position       string                `json:"position"`
	Assignees      string                `json:"assignees"`
	Conditions     string                `json:"conditions"`
	Settings       string                `json:"settings"`
	NextNodes      string                `json:"next_nodes"`
	FormConditions string                `json:"form_conditions"`
	SortOrder      int                   `json:"sort_order"`
}

type CreateWorkflowWithTreeRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	Category    string                  `json:"category"`
	FormID      uint                    `json:"form_id"`
	NodeTree    *models.NodeTreeData    `json:"node_tree"`
}

type StartWorkflowRequest struct {
	WorkflowID   uint   `json:"workflow_id" binding:"required"`
	Title        string `json:"title" binding:"required"`
	BusinessKey  string `json:"business_key"`
	BusinessType string `json:"business_type"`
	BusinessData string `json:"business_data"`
}

type StartWorkflowWithFormRequest struct {
	WorkflowID   uint   `json:"workflow_id" binding:"required"`
	Title        string `json:"title" binding:"required"`
	BusinessKey  string `json:"business_key"`
	BusinessType string `json:"business_type"`
	BusinessData string `json:"business_data"`
	FormValues   string `json:"form_values"`
	Variables    string `json:"variables"`
}

type AssigneeConfig struct {
	Type          string `json:"type"` // users, roles, departments
	UserIDs       []uint `json:"user_ids"`
	RoleIDs       []uint `json:"role_ids"`
	DepartmentIDs []uint `json:"department_ids"`
} 