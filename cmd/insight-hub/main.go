package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kitsnail/insight-hub/internal/api"
	"github.com/kitsnail/insight-hub/internal/config"
	"github.com/kitsnail/insight-hub/internal/repository"
	"github.com/kitsnail/insight-hub/internal/service"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "", "配置文件路径")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本
	if *showVersion {
		fmt.Printf("insight-hub %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	log.Printf("启动 Insight Hub...")
	log.Printf("数据目录: %s", cfg.Storage.DataDir)
	log.Printf("监听地址: %s:%d", cfg.Server.Host, cfg.Server.Port)

	// 初始化数据库
	db, err := repository.NewDatabase(cfg.Storage.DataDir)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	log.Printf("数据库初始化完成")

	// 初始化 Repositories
	itemRepo := repository.NewItemRepository(db)
	tagRepo := repository.NewTagRepository(db)

	// 初始化 Services
	itemSvc := service.NewItemService(itemRepo, tagRepo)
	tagSvc := service.NewTagService(tagRepo)

	// 初始化 API Handler
	handler := api.NewHandler(itemSvc, tagSvc)

	// 创建路由
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 静态文件服务（Web UI）
	webDir := "web/static"
	if _, err := os.Stat(webDir); err == nil {
		fs := http.FileServer(http.Dir(webDir))
		mux.Handle("/", fs)
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("正在关闭...")
		server.Close()
	}()

	log.Printf("服务器启动: http://%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器错误: %v", err)
	}
}
