package router

import (
	"viperai/internal/infrastructure/database"
	"viperai/internal/repository"
	"viperai/internal/service"
	"viperai/internal/transport/http/handler"
	"viperai/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	userRepo := repository.NewUserRepository(database.GetDB())
	convRepo := repository.NewConversationRepository(database.GetDB())
	msgRepo := repository.NewMessageRepository(database.GetDB())

	userService := service.NewUserService(userRepo)
	chatService := service.NewChatService(convRepo, msgRepo)
	fileService := service.NewFileService()
	codeService := service.NewCodeService()
	imageService := service.NewImageService("/root/models/mobilenetv2/mobilenetv2-7.onnx", "/root/imagenet_classes.txt", 224, 224)
	ttsService := service.NewTTSService()

	userHandler := handler.NewUserHandler(userService)
	chatHandler := handler.NewChatHandler(chatService)
	fileHandler := handler.NewFileHandler(fileService)
	codeHandler := handler.NewCodeHandler(codeService)
	imageHandler := handler.NewImageHandler(imageService)
	ttsHandler := handler.NewTTSHandler(ttsService)

	api := r.Group("/api/v1")
	{
		userGroup := api.Group("/user")
		{
			userGroup.POST("/login", userHandler.Login)
			userGroup.POST("/register", userHandler.Register)
			userGroup.POST("/captcha", userHandler.SendCaptcha)
			userGroup.GET("/profile", middleware.AuthRequired(), userHandler.GetProfile)
		}

		chatGroup := api.Group("/chat")
		chatGroup.Use(middleware.AuthRequired())
		{
			chatGroup.GET("/conversations", chatHandler.GetConversations)
			chatGroup.POST("/send-new", chatHandler.CreateAndSend)
			chatGroup.POST("/send", chatHandler.Send)
			chatGroup.POST("/stream-new", chatHandler.CreateAndStream)
			chatGroup.POST("/stream", chatHandler.Stream)
			chatGroup.POST("/history", chatHandler.GetHistory)
		}

		fileGroup := api.Group("/file")
		fileGroup.Use(middleware.AuthRequired())
		{
			fileGroup.POST("/upload", fileHandler.Upload)
		}

		codeGroup := api.Group("/code")
		codeGroup.Use(middleware.AuthRequired())
		{
			codeGroup.POST("/execute", codeHandler.Execute)
			codeGroup.GET("/languages", codeHandler.GetLanguages)
			codeGroup.POST("/analyze", codeHandler.Analyze)
			codeGroup.POST("/format", codeHandler.Format)
			codeGroup.POST("/test", codeHandler.RunTests)
		}

		imageGroup := api.Group("/image")
		imageGroup.Use(middleware.AuthRequired())
		{
			imageGroup.POST("/recognize", imageHandler.Recognize)
		}

		ttsGroup := api.Group("/tts")
		ttsGroup.Use(middleware.AuthRequired())
		{
			ttsGroup.POST("/create", ttsHandler.CreateTask)
			ttsGroup.GET("/query", ttsHandler.QueryTask)
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
