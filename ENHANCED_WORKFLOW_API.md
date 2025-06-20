# 增强版审批流系统 API 文档 - 完整版

## 概述

本系统基于您提供的 `node.txt` 文件结构，实现了一个功能完整的企业级审批流管理系统，支持：

- **复杂节点树结构**：支持嵌套的条件分支、并行节点等复杂流程设计
- **自定义表单设计**：完全可配置的动态表单，支持多种字段类型和验证规则
- **表单数据集成**：工作流与表单数据深度集成，支持条件判断和数据传递
- **灵活的审批模式**：支持依次审批、并行审批、任意审批、全员审批等多种模式
- **细粒度权限控制**：基于角色的权限管理，支持实例级和任务级权限控制
- **完整的审计日志**：记录所有操作历史，包括表单数据变更
- **完整的CRUD操作**：支持表单和工作流的创建、读取、更新、删除操作

## 认证

所有API都需要通过JWT token进行认证：
```
Authorization: Bearer <your-jwt-token>
```

## 表单管理 API（完整版）

### 1. 创建表单定义

```http
POST /api/v1/forms
Content-Type: application/json
Authorization: Bearer <token>

{
  "object": "custom_obj_2841513138091747",
  "name": "测试表单",
  "key": "custom_form_10162416350399872",
  "cards": [...],
  "buttons": [...]
}
```

### 2. 更新表单定义

```http
PUT /api/v1/forms/1
Content-Type: application/json
Authorization: Bearer <token>

{
  "object": "custom_obj_2841513138091747",
  "name": "更新后的表单名称",
  "key": "custom_form_10162416350399872",
  "cards": [...],
  "buttons": [...]
}
```

### 3. 删除表单定义

```http
DELETE /api/v1/forms/1
Authorization: Bearer <token>
```

### 4. 导出表单定义

```http
GET /api/v1/forms/1/export
Authorization: Bearer <token>
```

### 5. 克隆表单定义

```http
POST /api/v1/forms/1/clone
Content-Type: application/json
Authorization: Bearer <token>

{
  "new_name": "克隆的表单"
}
```

### 6. 激活表单定义

```http
PUT /api/v1/forms/1/activate
Authorization: Bearer <token>
```

### 7. 停用表单定义

```http
PUT /api/v1/forms/1/deactivate
Authorization: Bearer <token>
```

### 8. 获取表单列表

```http
GET /api/v1/forms?page=1&page_size=20
Authorization: Bearer <token>
```

### 9. 获取表单详情

```http
GET /api/v1/forms/1
Authorization: Bearer <token>
```

### 10. 根据key获取表单定义

```http
GET /api/v1/forms/key/custom_form_10162416350399872
Authorization: Bearer <token>
```

## 工作流管理 API（增强版）

### 1. 从JSON导入工作流和表单（支持node.txt格式）

```http
POST /api/v1/workflows/import
Content-Type: application/json
Authorization: Bearer <token>

{
  "workflow": {
    "name": "费用报销审批流程",
    "description": "支持复杂条件分支的费用报销流程",
    "category": "财务",
    "node_tree": {
      "key": "root_node",
      "name": "提交申请",
      "type": "ROOT",
      "child": {
        "key": "approval_node_1",
        "name": "部门经理审批",
        "type": "approval",
        "child": {
          "key": "condition_node_1",
          "name": "金额条件判断",
          "type": "condition",
          "branches": [
            {
              "key": "high_amount_branch",
              "name": "高金额分支",
              "type": "condition",
              "child": {
                "key": "finance_approval",
                "name": "财务总监审批",
                "type": "approval",
                "child": {
                  "key": "end_node",
                  "name": "结束",
                  "type": "END"
                }
              }
            },
            {
              "key": "low_amount_branch",
              "name": "低金额分支",
              "type": "condition",
              "child": {
                "key": "end_node_2",
                "name": "结束",
                "type": "END"
              }
            }
          ]
        }
      }
    }
  },
  "form": {
    "object": "expense_form",
    "name": "费用报销表单",
    "key": "expense_form_001",
    "cards": [
      {
        "name": "基本信息",
        "attributes": [
          {
            "attribute": {
              "object": "Expense",
              "name": "报销金额",
              "key": "amount",
              "type": "NUMBER",
              "element": "number"
            },
            "element": "number",
            "name": "报销金额",
            "required": true,
            "location": { "x": 1, "y": 1 }
          },
          {
            "attribute": {
              "object": "Expense",
              "name": "报销事由",
              "key": "purpose",
              "type": "STRING",
              "element": "text"
            },
            "element": "text",
            "name": "报销事由",
            "required": true,
            "location": { "x": 1, "y": 2 }
          }
        ]
      }
    ],
    "buttons": [
      {
        "name": "提交申请",
        "type": "primary",
        "click": { "operate": "start" }
      }
    ]
  }
}
```

### 2. 更新工作流状态

```http
PUT /api/v1/workflows/1/status
Content-Type: application/json
Authorization: Bearer <token>

{
  "status": "active"
}
```

### 3. 获取工作流节点树

```http
GET /api/v1/workflows/1/node-tree
Authorization: Bearer <token>
```

## 表单数据管理 API

### 1. 创建表单数据

```http
POST /api/v1/form-data
Content-Type: application/json
Authorization: Bearer <token>

{
  "form_id": 1,
  "instance_id": 1,
  "business_key": "REIMB-2024-001",
  "form_values": "{\"amount\":1500,\"purpose\":\"客户拜访差旅费\"}"
}
```

### 2. 更新表单数据

```http
PUT /api/v1/form-data/1
Content-Type: application/json
Authorization: Bearer <token>

{
  "form_values": "{\"amount\":1800,\"purpose\":\"客户拜访差旅费（已更新）\"}",
  "status": "submitted"
}
```

### 3. 获取表单数据

```http
GET /api/v1/form-data/1
Authorization: Bearer <token>
```

## 工作流实例管理 API

### 1. 启动带表单的工作流实例

```http
POST /api/v1/instances/with-form
Content-Type: application/json
Authorization: Bearer <token>

{
  "workflow_id": 1,
  "title": "张三差旅费报销申请",
  "business_key": "REIMB-2024-001",
  "business_type": "expense_reimbursement",
  "form_values": "{\"amount\":1500,\"purpose\":\"客户拜访差旅费\"}",
  "variables": "{\"department\":\"销售部\",\"manager_id\":2}"
}
```

### 2. 获取实例的表单数据

```http
GET /api/v1/instances/1/form-data
Authorization: Bearer <token>
```

### 3. 取消工作流实例

```http
PUT /api/v1/instances/1/cancel
Authorization: Bearer <token>
```

## 任务处理 API

### 1. 带表单数据的审批

```http
POST /api/v1/tasks/1/approve
Content-Type: application/json
Authorization: Bearer <token>

{
  "comment": "同意报销，金额合理",
  "form_values": "{\"approval_amount\":1500,\"approval_note\":\"已核实发票\"}"
}
```

### 2. 拒绝任务

```http
POST /api/v1/tasks/1/reject
Content-Type: application/json
Authorization: Bearer <token>

{
  "comment": "单据不完整，请补充相关材料"
}
```

### 3. 获取我的待办任务

```http
GET /api/v1/tasks/my
Authorization: Bearer <token>
```

## 系统管理 API

### 1. 获取工作流统计信息

```http
GET /api/v1/workflow/statistics
Authorization: Bearer <token>
```

### 2. 权限管理

```http
# 获取角色列表
GET /api/v1/admin/roles
Authorization: Bearer <token>

# 创建角色
POST /api/v1/admin/roles
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "部门经理",
  "description": "部门经理角色",
  "permissions": ["workflow_read", "instance_approve"]
}
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突（如表单正在使用中不能删除） |
| 500 | 服务器内部错误 |

## 完整功能特性

### 1. 表单管理
- ✅ 创建表单定义
- ✅ 更新表单定义（版本管理）
- ✅ 删除表单定义（安全检查）
- ✅ 导出表单定义
- ✅ 克隆表单定义
- ✅ 激活/停用表单
- ✅ 表单预览和验证

### 2. 工作流管理
- ✅ 复杂节点树结构支持
- ✅ 条件分支和并行节点
- ✅ 表单与工作流深度集成
- ✅ 工作流状态管理
- ✅ 支持node.txt格式导入

### 3. 审批流程
- ✅ 多种审批模式
- ✅ 动态审批人分配
- ✅ 表单数据条件判断
- ✅ 完整的审批历史记录

### 4. 权限控制
- ✅ 基于角色的权限管理
- ✅ 实例级权限检查
- ✅ 任务级权限验证

### 5. 数据安全
- ✅ 软删除机制
- ✅ 级联删除配置
- ✅ 权限校验
- ✅ 数据完整性检查

这个系统现在是一个功能完整的企业级审批流管理平台，完全支持您提供的node.txt格式，可以处理复杂的业务场景。 