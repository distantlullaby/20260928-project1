package main

import (
	"log"

	"mbti-backend/config"
	"mbti-backend/handlers"
	"mbti-backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()

	r := gin.Default()
	r.Use(middleware.CORS())

	api := r.Group("/api")
	{
		// 认证
		api.POST("/auth/register", handlers.Register)
		api.POST("/auth/login", handlers.Login)

		// 测评题库（公开）
		api.GET("/quiz/questions", handlers.GetQuestions)

		// 求解读信息流（公开浏览）
		api.GET("/posts", handlers.ListPosts)
		api.GET("/posts/:id", handlers.GetPost)

		// 需登录
		auth := api.Group("")
		auth.Use(middleware.Auth())
		{
			auth.POST("/quiz/submit", handlers.SubmitAssessment)
			auth.GET("/quiz/latest", handlers.GetLatestAssessment)

			auth.POST("/posts", handlers.CreatePost)
			auth.POST("/posts/:id/append", handlers.AppendBounty)
			auth.POST("/posts/:id/responses", handlers.CreateResponse)
			auth.POST("/posts/:id/responses/:responseId/accept", handlers.AcceptResponse)

			auth.GET("/me", handlers.GetMe)
			auth.GET("/me/overview", handlers.GetMyPostOverview)
			auth.GET("/me/transactions", handlers.GetCoinTransactions)
		}
	}

	log.Println("MBTI 服务已启动: http://localhost:8088")
	if err := r.Run(":8088"); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
