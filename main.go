package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"video_feed/models"
	"video_feed/mq_task"
	"video_feed/routes"
	"video_feed/utils"
)

func main() {

	//初始化数据库
	models.InitDB()
	//初始化redis
	models.InitRedis()
	//初始化RabbitMQ
	models.InitMQ()
	defer models.CloseMQ()
	//初始化JWT密钥
	utils.InitJWTSecret()

	mq_task.StartConsumer() // 开启后台消费者协程逻辑
	//设置 Gin 路由
	r := routes.SetupRouter()

	//优雅停机
	srv := &http.Server{
		Addr:    ":8085",
		Handler: r,
	}
	//启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%s\n", err)
		}
	}()
	//等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit //阻塞，直到收到信号
	log.Println("Shutting down server")

	//设置15秒时间，让现有请求处理完成
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
