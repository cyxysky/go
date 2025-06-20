package main

import (
	"log"

	"gin-web-api/config"
	"gin-web-api/database"
	"gin-web-api/models"
	redisClient "gin-web-api/redis"
	"gin-web-api/routes"
	"gin-web-api/services"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	database.InitDatabase(cfg)

	// 初始化Redis
	redisClient.InitRedis(cfg)

	// 自动迁移数据库表
	if err := database.AutoMigrate(
		// 原有模型
		&models.User{}, 
		
		// 权限相关模型
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.Department{},
		&models.UserDepartment{},
		
		// 表单相关模型
		&models.FormDefinition{},
		&models.FormCard{},
		&models.FormAttribute{},
		&models.FieldAttribute{},
		&models.FormButton{},
		&models.FormData{},
		
		// 工作流相关模型
		&models.WorkflowDefinition{},
		&models.WorkflowNode{},
		&models.WorkflowBranch{},
		&models.WorkflowInstance{},
		&models.WorkflowTask{},
		&models.WorkflowHistory{},
	); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 初始化默认权限数据
	permissionService := services.NewPermissionService()
	if err := permissionService.InitializeDefaultData(); err != nil {
		log.Printf("权限数据初始化警告: %v", err)
	} else {
		log.Println("权限数据初始化完成")
	}

	// 设置路由
	r := routes.SetupRoutes(cfg)

	// 启动服务器
	log.Printf("服务器启动在端口 %s", cfg.Port)
	log.Println("增强审批流系统已启动，功能包括:")
	log.Println("- 工作流定义管理（支持复杂节点树结构）")
	log.Println("- 自定义表单设计和管理")
	log.Println("- 动态表单数据处理")
	log.Println("- 工作流实例管理（支持表单集成）") 
	log.Println("- 任务审批处理（支持表单数据）")
	log.Println("- 条件分支和复杂流转逻辑")
	log.Println("- 基于角色的细粒度权限控制")
	log.Println("- 完整的审批历史记录")
	log.Println("- 支持node.txt格式的导入导出")
	
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
