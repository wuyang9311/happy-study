package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT 密钥（生产环境应放在环境变量中）
var jwtSecret = []byte("happy-study-jwt-secret-2026")

// Claims JWT 声明
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// SetJWTSecret 设置JWT密钥（从环境变量读取）
func SetJWTSecret(secret string) {
	if secret != "" {
		jwtSecret = []byte(secret)
	}
}

// GenerateToken 生成JWT Token
func GenerateToken(user *User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7天过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "happy-study",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析JWT Token
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HashPassword bcrypt 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tokenStr := c.GetHeader("Authorization")
		if len(tokenStr) == 0 {
			c.JSON(http.StatusUnauthorized, utils.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 去除 "Bearer " 前缀
		tokenStr = tokenStr[7:] // len("Bearer ") = 7

		claims, err := ParseToken(string(tokenStr))
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.H{"error": "认证令牌无效或已过期"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next(ctx)
	}
}

// Register 注册
func Register(ctx context.Context, c *app.RequestContext) {
	var req RegisterReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数格式错误"})
		return
	}

	// 参数校验
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, utils.H{"error": "用户名和密码不能为空"})
		return
	}

	// 清洗用户名
	req.Username = SanitizeUsername(req.Username)
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, utils.H{"error": "用户名包含非法字符"})
		return
	}

	// 用户名合法性检查
	if ok, msg := ValidateUsername(req.Username); !ok {
		c.JSON(http.StatusBadRequest, utils.H{"error": msg})
		return
	}

	// 检查保留用户名
	if IsReservedUsername(req.Username) {
		c.JSON(http.StatusBadRequest, utils.H{"error": "该用户名已被系统保留"})
		return
	}

	// 敏感词检查
	if ContainsSensitiveWord(req.Username) || ContainsSensitiveWord(req.Nickname) {
		c.JSON(http.StatusBadRequest, utils.H{"error": "用户名或昵称包含敏感词"})
		return
	}

	// 密码强度检查
	if ok, msg := ValidatePassword(req.Password); !ok {
		c.JSON(http.StatusBadRequest, utils.H{"error": msg})
		return
	}

	// 用户名查重
	existing, err := FindUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "服务器内部错误"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, utils.H{"error": "用户名已被注册"})
		return
	}

	// 加密密码
	hash, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "服务器内部错误"})
		return
	}

	// 默认昵称
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	user := &User{
		Username:     req.Username,
		PasswordHash: hash,
		Nickname:     nickname,
		Email:        req.Email,
		Role:         "user",
		Status:       1,
	}

	if err := CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "注册失败，请稍后重试"})
		return
	}

	// 生成 token
	token, err := GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, LoginResp{
		Token:    token,
		UserInfo: user,
	})
}

// Login 登录
func Login(ctx context.Context, c *app.RequestContext) {
	var req LoginReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数格式错误"})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, utils.H{"error": "用户名和密码不能为空"})
		return
	}

	// 清理用户名
	req.Username = SanitizeUsername(req.Username)

	// 查找用户
	user, err := FindUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "服务器内部错误"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, utils.H{"error": "用户名或密码错误"})
		return
	}

	// 检查账号状态
	if user.Status == 0 {
		c.JSON(http.StatusForbidden, utils.H{"error": "账号已被禁用"})
		return
	}

	// 验证密码
	if !CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, utils.H{"error": "用户名或密码错误"})
		return
	}

	// 更新最后登录时间
	_ = UpdateLastLogin(user.ID)

	// 生成 token
	token, err := GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, LoginResp{
		Token:    token,
		UserInfo: user,
	})
}

// GetProfile 获取当前用户信息
func GetProfile(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.H{"error": "认证信息无效"})
		return
	}

	user, err := FindUserByID(uid)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, UserInfoResp{User: user})
}

// GetUserSettings 获取用户设置
func GetUserSettings(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.H{"error": "认证信息无效"})
		return
	}

	user, err := FindUserByID(uid)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, utils.H{
		"preferred_model": user.PreferredModel,
	})
}

// UpdateUserModel 更新用户首选模型
func UpdateUserModel(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.H{"error": "认证信息无效"})
		return
	}

	var req struct {
		Model string `json:"model"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数格式错误"})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, utils.H{"error": "模型名称不能为空"})
		return
	}

	if err := UpdatePreferredModel(uid, req.Model); err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, utils.H{"preferred_model": req.Model})
}
