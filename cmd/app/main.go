package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
	"github.com/joho/godotenv"

	"github.com/wuyang9311/happy-study/internal/agent"
	"github.com/wuyang9311/happy-study/internal/agent/interviewer"
	"github.com/wuyang9311/happy-study/internal/agent/teacher"
	"github.com/wuyang9311/happy-study/internal/auth"
	"github.com/wuyang9311/happy-study/internal/handler"
	"github.com/wuyang9311/happy-study/internal/llm"
	"github.com/wuyang9311/happy-study/internal/middleware"
	"github.com/wuyang9311/happy-study/internal/service"
	"github.com/wuyang9311/happy-study/internal/store"
)

func main() {
	_ = godotenv.Load()

	// ========== 数据库初始化 ==========
	mysqlDSN := getEnv("MYSQL_DSN", "root:qwer1234@tcp(172.30.48.1:3306)/happy_study?charset=utf8mb4&parseTime=True&loc=Local")
	if err := auth.InitDB(mysqlDSN); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	log.Println("✅ MySQL 数据库连接成功")

	// JWT 密钥
	auth.SetJWTSecret(getEnv("JWT_SECRET", ""))

	// ========== 会话存储（MySQL） ==========
	s, err := store.NewMySQLStore(auth.GetDB())
	if err != nil {
		log.Fatalf("初始化会话存储失败: %v", err)
	}
	log.Printf("📦 会话存储已初始化（MySQL）")
	log.Printf("🔄 已加载历史会话: %d 个", s.Count())

	ctx := context.Background()
	llmCfg := llm.DefaultConfig()
	provider, err := llm.NewProvider(ctx, llmCfg)
	if err != nil {
		log.Fatalf("创建 LLM Provider 失败: %v", err)
	}

	prompts, err := agent.NewPromptManager()
	if err != nil {
		log.Fatalf("加载 Prompt 模板失败: %v", err)
	}

	intv := interviewer.NewInterviewer(provider, prompts)
	tchr := teacher.NewTeacher(provider, prompts)

	sessionMgr := service.NewSessionManager(s, intv, tchr)

	// ========== HTTP 服务 ==========
	port := getEnv("PORT", "8080")
	hz := server.Default(server.WithHostPorts("0.0.0.0:" + port))

	// 中间件
	hz.Use(middleware.Recovery(), middleware.Logger())
	hz.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// ========== 认证路由（无需登录） ==========
	hz.POST("/api/auth/register", auth.Register)
	hz.POST("/api/auth/login", auth.Login)

	h := handler.NewHandler(sessionMgr, intv, tchr)
	hz.GET("/api/health", h.HealthCheck)

	// ========== 需要登录的认证路由 ==========
	apiAuth := hz.Group("/api/auth")
	apiAuth.Use(auth.AuthMiddleware())
	{
		apiAuth.GET("/profile", auth.GetProfile)
	}

	// ========== 原有 API 路由（需要登录） ==========
	api := hz.Group("/api")
	api.Use(auth.AuthMiddleware())
	{
		api.POST("/diagnosis/start", h.StartDiagnosis)
		api.POST("/diagnosis/answer", h.SubmitAnswer)
		api.GET("/diagnosis/report/:session_id", h.GetReport)

		api.POST("/curriculum/generate", h.GenerateCurriculum)
		api.GET("/curriculum/:session_id", h.GetCurriculum)
		api.POST("/curriculum/lesson", h.GenerateLesson)

		api.POST("/diagnosis/stop", h.StopDiagnosis)

		api.GET("/session/:session_id", h.GetSessionInfo)

		// 用户课程
		api.GET("/user/courses", h.ListUserCourses)
		api.GET("/user/courses/:session_id", h.GetUserCourseDetail)

		// 流式出题
		api.GET("/diagnosis/question/:session_id", h.StreamFirstQuestion)
	}

	// ========== 启动 ==========
	log.Printf("🚀 Happy Study API 启动成功，监听端口 %s", port)
	fmt.Println("  注册: POST   /api/auth/register")
	fmt.Println("  登录: POST   /api/auth/login")
	fmt.Println("  个人信息: GET  /api/auth/profile (需 Bearer Token)")
	hz.Spin()
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
