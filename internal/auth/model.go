package auth

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string         `gorm:"type:varchar(100);default:''" json:"nickname"`
	Email        string         `gorm:"type:varchar(255);default:''" json:"email"`
	Avatar       string         `gorm:"type:varchar(500);default:''" json:"avatar"`
	Role         string         `gorm:"type:varchar(20);default:'user'" json:"role"`
	Status       int8           `gorm:"type:tinyint;default:1;index" json:"status"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// RegisterReq 注册请求
type RegisterReq struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=6,max=64"`
	Nickname string `json:"nickname" validate:"max=50"`
	Email    string `json:"email" validate:"omitempty,email"`
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResp 登录响应
type LoginResp struct {
	Token    string `json:"token"`
	UserInfo *User  `json:"user_info"`
}

// UserInfoResp 用户信息响应
type UserInfoResp struct {
	User *User `json:"user"`
}
