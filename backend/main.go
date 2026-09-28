package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"mbti-backend/config"
	"mbti-backend/database"
	"mbti-backend/handlers"
	"mbti-backend/middleware"
)

func main() {
	database.Init()
	database.Seed()

	r := setupRouter()
	log.Printf("MBTI 测评服务启动: http://localhost%s", config.AppAddr)
	if err := r.Run(config.AppAddr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	// CORS：前端 Vite dev server 直连
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	// 公开接口
	api.POST("/auth/register", handlers.Register)
	api.POST("/auth/login", handlers.Login)
	api.GET("/questions", handlers.ListQuestions)
	api.GET("/profiles", handlers.ListProfiles)
	api.GET("/results/:id", handlers.GetResult)
	api.GET("/requests", handlers.ListRequests)
	api.GET("/requests/:id", handlers.GetRequest)

	// 需要登录
	auth := api.Group("")
	auth.Use(middleware.Auth())
	{
		auth.GET("/me", handlers.Me)
		auth.POST("/assessments", handlers.SubmitAssessment)
		auth.GET("/my/results", handlers.MyResults)
		auth.GET("/my/transactions", handlers.MyTransactions)
		auth.GET("/my/requests", handlers.ListRequests)
		auth.GET("/my/interpretations", handlers.InterpretedByMe)

		auth.POST("/requests", handlers.CreateRequest)
		auth.POST("/requests/:id/append", handlers.AppendReward)
		auth.POST("/requests/:id/close", handlers.CloseRequest)
		auth.POST("/requests/:id/interpretations", handlers.SubmitInterpretation)
		auth.POST("/requests/:id/interpretations/:interpId/accept", handlers.AcceptInterpretation)
	}

	return r
}
