package handlers

import (
	"net/http"
	"strconv"

	"gin-web-api/database"
	"gin-web-api/models"
	redisClient "gin-web-api/redis"

	"github.com/gin-gonic/gin"
)

type PostHandler struct{}

func NewPostHandler() *PostHandler {
	return &PostHandler{}
}

// CreatePost 创建文章
func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认状态
	if req.Status == "" {
		req.Status = "draft"
	}

	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		Summary: req.Summary,
		Status:  req.Status,
		UserID:  userID.(uint),
	}

	if err := database.GetDB().Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文章创建失败"})
		return
	}

	// 预加载用户信息
	database.GetDB().Preload("User").First(&post, post.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "文章创建成功",
		"post":    post.ToResponse(),
	})
}

// GetPosts 获取文章列表
func (h *PostHandler) GetPosts(c *gin.Context) {
	var posts []models.Post
	query := database.GetDB().Preload("User")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// 状态筛选
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 用户筛选
	userID := c.Query("user_id")
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// 搜索
	search := c.Query("search")
	if search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 执行查询
	var total int64
	query.Model(&models.Post{}).Count(&total)
	
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文章列表失败"})
		return
	}

	// 转换为响应格式
	var postResponses []models.PostResponse
	for _, post := range posts {
		postResponses = append(postResponses, post.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": postResponses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetPost 获取单个文章
func (h *PostHandler) GetPost(c *gin.Context) {
	id := c.Param("id")
	
	var post models.Post
	if err := database.GetDB().Preload("User").First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 增加查看次数（使用Redis计数器）
	viewKey := "post_views:" + id
	redisClient.Client.Incr(redisClient.Client.Context(), viewKey)
	
	// 异步更新数据库中的查看次数
	go func() {
		if count, err := redisClient.Get(viewKey); err == nil {
			if viewCount, err := strconv.Atoi(count); err == nil {
				database.GetDB().Model(&post).Update("view_count", viewCount)
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"post": post.ToResponse(),
	})
}

// UpdatePost 更新文章
func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	id := c.Param("id")
	var post models.Post
	if err := database.GetDB().First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 检查权限
	if post.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限修改此文章"})
		return
	}

	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}  
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := database.GetDB().Model(&post).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文章更新失败"})
		return
	}

	// 重新查询更新后的文章
	database.GetDB().Preload("User").First(&post, id)

	c.JSON(http.StatusOK, gin.H{
		"message": "文章更新成功",
		"post":    post.ToResponse(),
	})
}

// DeletePost 删除文章
func (h *PostHandler) DeletePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	id := c.Param("id")
	var post models.Post
	if err := database.GetDB().First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 检查权限
	if post.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除此文章"})
		return
	}

	if err := database.GetDB().Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文章删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章删除成功"})
} 