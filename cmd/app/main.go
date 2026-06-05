package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"

	"github.com/joho/godotenv"

	agent "github.com/wuyang9311/happy-study/internal/agent"
	"github.com/wuyang9311/happy-study/internal/agent/interviewer"
	"github.com/wuyang9311/happy-study/internal/agent/teacher"
	"github.com/wuyang9311/happy-study/internal/handler"
	"github.com/wuyang9311/happy-study/internal/service"
)

func main() {
	// 加载 .env
	_ = godotenv.Load()

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 DEEPSEEK_API_KEY 环境变量（或在 .env 文件中配置）")
	}

	ctx := context.Background()
	llmConfig := &agent.LLMConfig{
		APIKey:  apiKey,
		BaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1"),
		Model:   getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
	}

	// 初始化 Agent
	intv, err := interviewer.NewInterviewer(ctx, llmConfig)
	if err != nil {
		log.Fatalf("创建面试官失败: %v", err)
	}
	tchr, err := teacher.NewTeacher(ctx, llmConfig)
	if err != nil {
		log.Fatalf("创建讲师失败: %v", err)
	}

	// 初始化服务
	sessionMgr := service.NewSessionManager()
	h := handler.NewHandler(sessionMgr, intv, tchr)

	// 启动 Hertz 服务
	port := getEnv("PORT", "8080")
	hz := server.Default(server.WithHostPorts("0.0.0.0:" + port))

	// CORS
	hz.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// API 路由
	api := hz.Group("/api")
	{
		api.POST("/diagnosis/start", h.StartDiagnosis)
		api.POST("/diagnosis/answer", h.SubmitAnswer)
		api.GET("/diagnosis/report/:session_id", h.GetReport)

		api.POST("/curriculum/generate", h.GenerateCurriculum)
		api.GET("/curriculum/:session_id", h.GetCurriculum)
		api.POST("/curriculum/lesson", h.GenerateLesson)

		api.GET("/session/:session_id", h.GetSessionInfo)
		api.GET("/health", h.HealthCheck)
	}

	log.Printf("🚀 Happy Study API 启动成功，监听端口 %s", port)
	log.Printf("   健康检查：http://localhost:%s/api/health", port)
	log.Printf("   诊断接口：POST http://localhost:%s/api/diagnosis/start", port)
	hz.Spin()
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
