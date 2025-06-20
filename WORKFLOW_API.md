# 审批流系统 API 文档

## 概述

本系统实现了一个完整的企业级审批流管理系统，包含以下核心功能：

- **工作流定义管理**：创建、编辑、部署工作流模板
- **工作流实例管理**：启动、监控、控制工作流实例
- **任务处理**：审批、拒绝、转办任务
- **权限控制**：基于角色的细粒度权限管理
- **审批历史**：完整的审批过程记录

## 认证

所有API都需要通过JWT token进行认证，在请求头中添加：
```
Authorization: Bearer <your-jwt-token>
```

## 工作流定义管理 API

### 1. 创建工作流定义

```http
POST /api/v1/workflows
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "费用报销审批",
  "description": "员工费用报销审批流程",
  "category": "财务",
  "nodes": [
    {
      "node_key": "start",
      "name": "开始",
      "type": "start",
      "next_nodes": "[\"approval1\"]",
      "sort_order": 1
    },
    {
      "node_key": "approval1",
      "name": "直属领导审批",
      "type": "approval",
      "approval_mode": "sequence",
      "assignees": "{\"type\":\"roles\",\"role_ids\":[2]}",
      "next_nodes": "[\"approval2\"]",
      "sort_order": 2
    },
    {
      "node_key": "approval2",
      "name": "财务审批",
      "type": "approval", 
      "approval_mode": "any",
      "assignees": "{\"type\":\"departments\",\"department_ids\":[1]}",
      "next_nodes": "[\"end\"]",
      "sort_order": 3
    },
    {
      "node_key": "end",
      "name": "结束",
      "type": "end",
      "sort_order": 4
    }
  ]
}
```

### 2. 获取工作流列表

```http
GET /api/v1/workflows
Authorization: Bearer <token>
```

### 3. 获取工作流详情

```http
GET /api/v1/workflows/{id}
Authorization: Bearer <token>
```

### 4. 更新工作流状态

```http
PUT /api/v1/workflows/{id}/status
Content-Type: application/json
Authorization: Bearer <token>

{
  "status": "active"
}
```

## 工作流实例管理 API

### 1. 启动工作流实例

```http
POST /api/v1/instances
Content-Type: application/json
Authorization: Bearer <token>

{
  "workflow_id": 1,
  "title": "张三差旅费报销申请",
  "business_key": "REIMB-2024-001",
  "business_type": "expense_reimbursement",
  "business_data": "{\"amount\":1500,\"purpose\":\"客户拜访差旅费\"}"
}
```

### 2. 获取实例列表

```http
GET /api/v1/instances
Authorization: Bearer <token>
```

### 3. 获取实例详情

```http
GET /api/v1/instances/{id}
Authorization: Bearer <token>
```

### 4. 取消实例

```http
PUT /api/v1/instances/{id}/cancel
Authorization: Bearer <token>
```

### 5. 获取实例历史记录

```http
GET /api/v1/instances/{id}/history
Authorization: Bearer <token>
```

## 任务处理 API

### 1. 获取我的待办任务

```http
GET /api/v1/tasks/my
Authorization: Bearer <token>
```

### 2. 审批通过任务

```http
POST /api/v1/tasks/{id}/approve
Content-Type: application/json
Authorization: Bearer <token>

{
  "comment": "同意报销，金额合理"
}
```

### 3. 拒绝任务

```http
POST /api/v1/tasks/{id}/reject
Content-Type: application/json
Authorization: Bearer <token>

{
  "comment": "缺少发票凭证，请补充后重新提交"
}
```

## 权限管理 API

### 1. 获取角色列表

```http
GET /api/v1/admin/roles
Authorization: Bearer <admin-token>
```

### 2. 创建角色

```http
POST /api/v1/admin/roles
Content-Type: application/json
Authorization: Bearer <admin-token>

{
  "name": "部门经理",
  "code": "dept_manager",
  "description": "部门经理角色"
}
```

### 3. 给用户分配角色

```http
POST /api/v1/admin/users/{user_id}/roles
Content-Type: application/json
Authorization: Bearer <admin-token>

{
  "role_id": 2
}
```

### 4. 给角色分配权限

```http
POST /api/v1/admin/roles/{role_id}/permissions
Content-Type: application/json
Authorization: Bearer <admin-token>

{
  "permission_id": 5
}
```

## 用户个人信息 API

### 1. 获取个人资料

```http
GET /api/v1/profile
Authorization: Bearer <token>
```

### 2. 检查权限

```http
POST /api/v1/check-permission
Content-Type: application/json
Authorization: Bearer <token>

{
  "permission_code": "workflow:create"
}
```

## 统计信息 API

### 获取工作流统计

```http
GET /api/v1/workflow/statistics
Authorization: Bearer <token>
```

## 系统初始化 API

### 初始化默认数据

```http
POST /api/v1/admin/initialize
Authorization: Bearer <admin-token>
```

## 错误处理

所有API都遵循统一的错误响应格式：

```json
{
  "error": "错误描述信息"
}
```

常见HTTP状态码：
- `200` - 成功
- `201` - 创建成功
- `400` - 请求参数错误
- `401` - 未授权
- `403` - 权限不足
- `404` - 资源不存在
- `500` - 服务器内部错误

## 工作流节点类型说明

- **start**: 开始节点，工作流的起点
- **end**: 结束节点，工作流的终点
- **approval**: 审批节点，需要人工审批处理
- **condition**: 条件节点，根据条件判断流向
- **parallel**: 并行节点，启动并行分支
- **merge**: 合并节点，合并并行分支

## 审批模式说明

- **sequence**: 依次审批，按顺序逐个审批
- **parallel**: 并行审批，同时进行审批
- **any**: 任意审批，任何一人审批通过即可
- **all**: 全员审批，所有人都必须审批通过

## 权限代码说明

### 工作流权限
- `workflow:create` - 创建工作流
- `workflow:read` - 查看工作流
- `workflow:update` - 编辑工作流
- `workflow:delete` - 删除工作流
- `workflow:deploy` - 部署工作流

### 实例权限
- `instance:create` - 创建实例
- `instance:read` - 查看实例
- `instance:cancel` - 取消实例
- `instance:suspend` - 挂起实例
- `instance:transfer` - 转办实例

### 任务权限
- `task:approve` - 审批任务
- `task:reject` - 拒绝任务
- `task:claim` - 认领任务
- `task:delegate` - 委派任务

### 系统权限
- `system:admin` - 系统管理
- `user:manage` - 用户管理
- `role:manage` - 角色管理

## 使用示例

### 完整的审批流程示例

1. **管理员创建工作流定义**
2. **激活工作流**
3. **用户启动工作流实例**
4. **审批人处理审批任务**
5. **查看审批历史**

详细示例代码请参考项目中的测试文件。 