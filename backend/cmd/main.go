package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"rain-yi-backend/config"
	"rain-yi-backend/controller"
	"rain-yi-backend/middleware"
	"rain-yi-backend/model"
	"rain-yi-backend/repository"
	"rain-yi-backend/service"
	"rain-yi-backend/skill"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db := config.InitDatabase()
	model.AutoMigrate(db)

	config.InitRedis(cfg)

	userRepo := repository.NewUserRepository()
	convRepo := repository.NewConversationRepository()
	msgRepo := repository.NewMessageRepository()
	personaRepo := repository.NewPersonaRepository()
	pfRepo := repository.NewPersonaFileRepository()
	fileRepo := repository.NewFileRepository()

	personaStg := service.NewPersonaStorage(pfRepo)
	fileStorage := service.NewMinioStorage(cfg, fileRepo)

	skillManager := skill.NewSkillManager(personaRepo, pfRepo, personaStg)
	if err := skillManager.LoadSkills(); err != nil {
		log.Printf("警告: 技能加载失败: %v", err)
	}

	aiService := service.NewAIService(skillManager)
	contextManager := service.NewContextManager(msgRepo)
	hub := service.NewWebSocketHub()

	authController := controller.NewAuthController(userRepo)
	conversationController := controller.NewConversationController(convRepo, msgRepo, contextManager)
	chatController := controller.NewChatController(msgRepo, convRepo, aiService, contextManager, hub)
	personaController := controller.NewPersonaController(
		personaRepo,
		pfRepo,
		convRepo,
		skillManager.PromptCache(),
		skillManager.PersonaCache(),
		personaStg,
	)
	userController := controller.NewUserController(userRepo)
	uploadController := controller.NewUploadController(fileRepo, userRepo, convRepo, fileStorage)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	r.Static("/static", "./static")

	if fileStorage != nil {
		r.GET("/storage/*filepath", func(c *gin.Context) {
			objPath := strings.TrimPrefix(c.Param("filepath"), "/")
			if objPath == "" {
				c.Status(http.StatusNotFound)
				return
			}

			minioStorage, ok := fileStorage.(*service.MinioStorage)
			if ok {
				objPath = strings.TrimPrefix(objPath, minioStorage.Bucket()+"/")
			}

			obj, err := fileStorage.Get(objPath)
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			defer obj.Close()

			ext := strings.ToLower(filepath.Ext(objPath))
			contentType := "application/octet-stream"
			switch ext {
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".png":
				contentType = "image/png"
			case ".gif":
				contentType = "image/gif"
			case ".webp":
				contentType = "image/webp"
			case ".svg":
				contentType = "image/svg+xml"
			case ".md":
				contentType = "text/markdown; charset=utf-8"
			}

			var fileSize int64 = -1
			if sized, ok := obj.(interface{ Size() int64 }); ok {
				fileSize = sized.Size()
			}

			c.Header("Cache-Control", "public, max-age=86400")
			if fileSize >= 0 {
				c.DataFromReader(http.StatusOK, fileSize, contentType, obj, nil)
			} else {
				c.Header("Content-Type", contentType)
				io.Copy(c.Writer, obj)
			}
		})
	}

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/logout", authController.Logout)
		}

		api.GET("/ws/chat", chatController.HandleWebSocket)

		authorized := api.Group("")
		authorized.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			authorized.GET("/user/profile", userController.GetProfile)
			authorized.PUT("/user/profile", userController.UpdateProfile)

			authorized.GET("/conversations", conversationController.GetConversations)
			authorized.POST("/conversations", conversationController.CreateConversation)
			authorized.GET("/conversations/:id/messages", conversationController.GetMessages)
			authorized.DELETE("/conversations/:id/messages", conversationController.ClearMessages)
			authorized.PUT("/conversations/:id/config", conversationController.UpdateConfig)

			personas := authorized.Group("/personas")
			{
				personas.GET("", personaController.GetPersonas)
				personas.GET("/:id", personaController.GetPersona)
				personas.GET("/:id/debug", personaController.DebugPrompt)
				personas.POST("", personaController.CreatePersona)
				personas.PUT("/:id", personaController.UpdatePersona)
				personas.DELETE("/:id", personaController.DeletePersona)
				personas.POST("/:id/files", personaController.UploadSkillFile)
				personas.DELETE("/:id/files/:fileId", personaController.DeleteSkillFile)
				personas.POST("/:id/avatar", personaController.UploadPersonaAvatar)
				personas.POST("/load", personaController.LoadFromDirectory)
			}

			authorized.GET("/conversations/:id/persona", personaController.GetConversationPersona)
			authorized.PUT("/conversations/:id/persona", personaController.SetConversationPersona)

			uploads := authorized.Group("/upload")
			{
				uploads.POST("", uploadController.UploadFile)
				uploads.POST("/avatar", uploadController.UploadAvatar)
				uploads.POST("/avatar/ai", uploadController.UploadAIAvatar)
				uploads.POST("/image", uploadController.UploadImage)
				uploads.DELETE("/:id", uploadController.DeleteFile)
				uploads.GET("/list", uploadController.ListFiles)
			}
		}
	}

	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("RainYi 服务启动于 %s", addr)
	log.Printf("前端地址: %s", cfg.FrontendURL)

	http.ListenAndServe(addr, r)
}
