package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	router "github.com/reguluswee/walletus/cmd/modapi/router"
	"github.com/reguluswee/walletus/common/config"
	"github.com/reguluswee/walletus/common/log"
)

func main() {
	fmt.Println("starting...")

	// 创建主上下文和取消函数
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建等待组，用于等待所有goroutine完成
	var wg sync.WaitGroup

	// 启动HTTP服务器
	server := router.Init()

	// 创建HTTP服务器实例
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GetConfig().Http.Port),
		Handler: server,
	}

	// 启动HTTP服务器
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("HTTP server starting...")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server failed to start", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-sigChan
	log.Info("Received signal:", sig)
	log.Info("Starting graceful shutdown...")

	// 取消上下文，通知所有goroutine停止
	cancel()

	// 创建关闭超时上下文
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 优雅关闭HTTP服务器
	log.Info("Shutting down HTTP server...")

	// 先停止接受新连接
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server graceful shutdown failed:", err)

		// 如果优雅关闭失败，强制关闭
		log.Warn("Forcing HTTP server to close...")
		forceCtx, forceCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer forceCancel()

		if err := httpServer.Shutdown(forceCtx); err != nil {
			log.Error("HTTP server forced shutdown failed:", err)
		}
	}

	// 等待所有goroutine完成
	log.Info("Waiting for all tasks to complete...")
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// 等待任务完成或超时
	select {
	case <-done:
		log.Info("All tasks completed successfully")
	case <-shutdownCtx.Done():
		log.Warn("Shutdown timeout reached, forcing exit")
		log.Warn("Some tasks may not have completed properly")
	}

	log.Info("Server shutdown complete")
}
