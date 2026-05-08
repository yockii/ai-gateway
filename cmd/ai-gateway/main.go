package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/yockii/ai-gateway/internal/config"
	"github.com/yockii/ai-gateway/internal/gateway"
	"github.com/yockii/ai-gateway/pkg/router"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化网关
	app, err := gateway.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize gateway: %v", err)
	}

	// 设置路由
	router.Setup(app.GetApp(), app)

	// 启动服务
	go func() {
		if err := app.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("🚀 AI Gateway started on %s", cfg.Server.Address)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Server stopped")
}
