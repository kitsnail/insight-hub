package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kitsnail/insight-hub/internal/api"
	"github.com/kitsnail/insight-hub/internal/config"
	"github.com/kitsnail/insight-hub/internal/repository"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// loggingMiddleware 请求日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 包装 ResponseWriter 以捕获状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		log.Printf("[%s] %s %s %d %v", r.Method, r.URL.Path, r.URL.RawQuery, wrapped.statusCode, duration)
	})
}

// responseWriter 包装 http.ResponseWriter 以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

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

	// 初始化 PostgreSQL
	if cfg.Database.Name == "" {
		log.Fatalf("请配置 PostgreSQL 数据库")
	}

	log.Printf("连接 PostgreSQL: %s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	pgDB, err := repository.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("初始化 PostgreSQL 失败: %v", err)
	}
	defer pgDB.Close()

	log.Printf("PostgreSQL 连接成功")

	// 初始化 Repositories
	itemRepo := repository.NewItemRepoV3(pgDB.Pool)
	tagRepo := repository.NewTagRepoPG(pgDB.Pool)
	taskRepo := repository.NewTaskRepoPG(pgDB.Pool)

	// 创建路由
	mux := http.NewServeMux()

	// 注册 v3 API (新版统一 API)
	handlerV3 := api.NewHandlerV3(itemRepo)
	handlerV3.RegisterRoutes(mux)

	// 请求日志中间件
	loggedMux := loggingMiddleware(mux)

	// 注册 v2 API (兼容旧版)
	handler := api.NewHandler(itemRepo, tagRepo, taskRepo)
	handler.RegisterRoutes(mux)

	// 静态文件服务（Web UI）
	webDir := "web/static"
	if _, err := os.Stat(webDir); err == nil {
		// v2 UI 重定向
		mux.HandleFunc("GET /v2", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/index-v2.html", http.StatusFound)
		})
		
		fs := http.FileServer(http.Dir(webDir))
		mux.Handle("/", fs)
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: loggedMux,
	}

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("正在关闭...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		server.Shutdown(ctx)
	}()

	log.Printf("服务器启动: http://%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器错误: %v", err)
	}
}
