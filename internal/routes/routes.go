package routes

import (
	"legaltech-backend/internal/config"
	"legaltech-backend/internal/handlers"
	"legaltech-backend/internal/middleware"
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/internal/ws"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, cfg *config.Config) {
	services.InitPushService(cfg)
	authService := services.NewAuthService(cfg)
	aiService := services.NewAIService(cfg)
	draftService := services.NewDraftService(aiService)
	researchService := services.NewResearchService(aiService)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(authService)
	dashboardHandler := handlers.NewDashboardHandler()
	clientHandler := handlers.NewClientHandler()
	caseHandler := handlers.NewCaseHandler()
	taskHandler := handlers.NewTaskHandler()
	hearingHandler := handlers.NewHearingHandler()
	calendarHandler := handlers.NewCalendarHandler()
	draftHandler := handlers.NewDraftHandler(draftService)
	researchHandler := handlers.NewResearchHandler(researchService)
	documentService := services.NewDocumentService()
	documentHandler := handlers.NewDocumentHandler()
	caseTimelineHandler := handlers.NewCaseTimelineHandler()
	dashboardStatsHandler := handlers.NewDashboardStatsHandler()
	aiDocumentService := services.NewAIDocumentService(aiService)
	aiRecommendationService := services.NewAIRecommendationService(aiService)
	aiDocumentHandler := handlers.NewAIDocumentHandler(aiDocumentService)
	aiRecommendationHandler := handlers.NewAIRecommendationHandler(aiRecommendationService)
	aiDraftingHandler := handlers.NewAIDraftingHandler(aiService, documentService)
	ocrHandler := handlers.NewOCRHandler(services.NewOCRService())
	meetingHandler := handlers.NewMeetingHandler()
	noteHandler := handlers.NewNoteHandler()
	notificationHandler := handlers.NewNotificationHandler()
	aiHandler := handlers.NewAIHandler(aiService)
	signalingHub := ws.NewHub()
	roomRegistry := ws.NewRoomRegistry()
	signalingHandler := handlers.NewSignalingHandler(signalingHub, cfg, roomRegistry)
	iceHandler := handlers.NewICEHandler(cfg)
	videoRoomService := services.NewVideoRoomService(signalingHub, roomRegistry)
	videoRoomHandler := handlers.NewVideoRoomHandler(videoRoomService)
	invoiceHandler := handlers.NewInvoiceHandler()
	clientPortalHandler := handlers.NewClientPortalHandler()
	adminHandler := handlers.NewAdminHandler()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// WebRTC signaling relay (offer/answer/ICE exchange only — no media).
		// Authenticated via a query-param JWT since WebSocket handshakes
		// can't carry a custom Authorization header from Flutter's client.
		v1.GET("/ws/signaling", signalingHandler.HandleWS)

		protected := v1.Group("")
		protected.Use(middleware.JWTAuthMiddleware(cfg))
		{
			user := protected.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PUT("/profile", userHandler.UpdateProfile)
				user.PUT("/change-password", userHandler.ChangePassword)
				user.PUT("/device-token", userHandler.UpdateDeviceToken)
			}

			protected.GET("/dashboard", dashboardHandler.Get)
			protected.GET("/dashboard/statistics", dashboardStatsHandler.Get)

			clients := protected.Group("/clients")
			{
				clients.POST("", clientHandler.Create)
				clients.GET("", clientHandler.List)
				clients.GET("/:id", clientHandler.Get)
				clients.PUT("/:id", clientHandler.Update)
				clients.PUT("/:id/archive", clientHandler.Archive)
				clients.GET("/:id/documents", clientHandler.ListDocuments)
				clients.GET("/:id/meetings", clientHandler.ListMeetings)
				clients.GET("/:id/invoices", clientHandler.ListInvoices)
				clients.DELETE("/:id", clientHandler.Delete)
				clients.PUT("/:id/reset-password", clientHandler.ResetPassword)
			}

			cases := protected.Group("/cases")
			{
				cases.POST("", caseHandler.Create)
				cases.GET("", caseHandler.List)
				cases.GET("/:id", caseHandler.Get)
				cases.PUT("/:id", caseHandler.Update)
				cases.DELETE("/:id", caseHandler.Delete)
				cases.POST("/:id/timeline", caseTimelineHandler.Create)
				cases.GET("/:id/timeline", caseTimelineHandler.List)
				cases.PUT("/:id/timeline/:eventId", caseTimelineHandler.Update)
				cases.POST("/:id/documents", documentHandler.Upload)
				cases.GET("/:id/documents", documentHandler.ListForCase)
				cases.POST("/:id/recommendations", aiRecommendationHandler.Generate)
				cases.GET("/:id/recommendations", aiRecommendationHandler.List)
			}

			tasks := protected.Group("/tasks")
			{
				tasks.POST("", taskHandler.Create)
				tasks.GET("", taskHandler.List)
				tasks.GET("/:id", taskHandler.Get)
				tasks.PUT("/:id", taskHandler.Update)
				tasks.DELETE("/:id", taskHandler.Delete)
			}

			hearings := protected.Group("/hearings")
			{
				hearings.POST("", hearingHandler.Create)
				hearings.GET("", hearingHandler.List)
				hearings.GET("/:id", hearingHandler.Get)
				hearings.PUT("/:id", hearingHandler.Update)
				hearings.DELETE("/:id", hearingHandler.Delete)
			}

			calendar := protected.Group("/calendar")
			{
				calendar.POST("", calendarHandler.Create)
				calendar.GET("", calendarHandler.List)
				calendar.GET("/:id", calendarHandler.Get)
				calendar.PUT("/:id", calendarHandler.Update)
				calendar.DELETE("/:id", calendarHandler.Delete)
			}
			// Alias matching the spec's literal /calendar/events path; same
			// handlers as /calendar above (kept for backward compatibility).
			calendarEvents := protected.Group("/calendar/events")
			{
				calendarEvents.POST("", calendarHandler.Create)
				calendarEvents.GET("", calendarHandler.List)
				calendarEvents.GET("/:id", calendarHandler.Get)
				calendarEvents.PUT("/:id", calendarHandler.Update)
				calendarEvents.DELETE("/:id", calendarHandler.Delete)
			}

			drafts := protected.Group("/drafts")
			{
				drafts.POST("", draftHandler.Create)
				drafts.POST("/generate", draftHandler.GenerateWithAI)
				drafts.GET("", draftHandler.List)
				drafts.GET("/:id", draftHandler.Get)
				drafts.PUT("/:id", draftHandler.Update)
				drafts.DELETE("/:id", draftHandler.Delete)
			}

			invoices := protected.Group("/invoices")
			{
				invoices.POST("", invoiceHandler.Create)
				invoices.GET("", invoiceHandler.List)
				invoices.GET("/:id", invoiceHandler.Get)
				invoices.GET("/:id/pdf", invoiceHandler.PDF)
				invoices.PUT("/:id", invoiceHandler.Update)
				invoices.POST("/:id/send", invoiceHandler.Send)
				invoices.PUT("/:id/mark-paid", invoiceHandler.MarkPaid)
				invoices.PUT("/:id/payment-status", invoiceHandler.SetPaymentStatus)
				invoices.GET("/:id/payment", invoiceHandler.GetPayment)
				invoices.DELETE("/:id", invoiceHandler.Delete)
			}

			// Lawyer's full Payment History across every invoice/client.
			protected.GET("/payments", invoiceHandler.PaymentHistory)

			research := protected.Group("/research")
			{
				research.POST("", researchHandler.Search)
				research.GET("", researchHandler.List)
				research.PUT("/:id/bookmark", researchHandler.SetBookmark)
				research.DELETE("/:id", researchHandler.Delete)
			}

			documents := protected.Group("/documents")
			{
				documents.POST("", documentHandler.Create)
				documents.GET("", documentHandler.List)
				documents.POST("/sign", documentHandler.Sign)
				documents.GET("/signature", documentHandler.GetSignature)
				documents.GET("/:id", documentHandler.Get)
				documents.GET("/:id/download", documentHandler.Download)
				documents.DELETE("/:id", documentHandler.Delete)
			}

			ocr := protected.Group("/ocr")
			{
				ocr.POST("/extract", ocrHandler.Extract)
			}

			meetings := protected.Group("/meetings")
			{
				meetings.POST("", meetingHandler.Create)
				meetings.GET("", meetingHandler.ListUpcoming)
				meetings.GET("/history", meetingHandler.ListHistory)
				meetings.GET("/:id", meetingHandler.Get)
				meetings.PUT("/:id", meetingHandler.Update)
				meetings.POST("/:id/join", meetingHandler.Join)
				meetings.POST("/:id/end", meetingHandler.End)
				meetings.POST("/:id/cancel", meetingHandler.Cancel)
				meetings.DELETE("/:id", meetingHandler.Delete)
			}

			notes := protected.Group("/notes")
			{
				notes.POST("", noteHandler.Create)
				notes.GET("", noteHandler.List)
				notes.GET("/:id", noteHandler.Get)
				notes.PUT("/:id", noteHandler.Update)
				notes.DELETE("/:id", noteHandler.Delete)
			}

			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notificationHandler.List)
				notifications.POST("", notificationHandler.Create)
				notifications.PUT("/:id/read", notificationHandler.MarkRead)
				notifications.PUT("/read-all", notificationHandler.MarkAllRead)
				notifications.PUT("/:id", notificationHandler.Update)
				notifications.DELETE("/:id", notificationHandler.Delete)
			}

			ai := protected.Group("/ai")
			{
				ai.POST("/chat", aiHandler.Chat)
				ai.GET("/conversations", aiHandler.ListConversations)
				ai.GET("/conversations/:id/messages", aiHandler.GetMessages)
				ai.DELETE("/conversations/:id", aiHandler.DeleteConversation)
				ai.POST("/summarize", aiHandler.Summarize)
				ai.POST("/documents/summarize", aiDocumentHandler.Summarize)
				ai.POST("/documents/contract-analysis", aiDocumentHandler.AnalyzeContract)
				ai.GET("/documents/history", aiDocumentHandler.History)
				ai.POST("/petition", aiDraftingHandler.Petition)
				ai.POST("/agreement", aiDraftingHandler.Agreement)
				ai.POST("/legal-notice", aiDraftingHandler.LegalNotice)
			}

			protected.GET("/ice-servers", iceHandler.GetServers)

			// Private, exactly-two-participant (lawyer + client) WebRTC video
			// calls — distinct from the /meetings CRUD (scheduling) above.
			videoRooms := protected.Group("/video-rooms")
			{
				videoRooms.POST("/start", videoRoomHandler.StartCall)
				videoRooms.GET("/active", videoRoomHandler.ActiveForMe)
				videoRooms.POST("/:id/join", videoRoomHandler.Join)
				videoRooms.POST("/:id/end", videoRoomHandler.End)
			}

			// Read-only client dashboard: My Cases / My Documents / My
			// Meetings / My Invoices / My Profile. Gated to Role == "client"
			// on top of the JWT check every other /api/v1 route already has;
			// every handler resolves "my own data" server-side from the
			// caller's linked Client record, never from client input.
			clientPortal := protected.Group("/client")
			clientPortal.Use(middleware.RequireRole(models.RoleClient))
			{
				clientPortal.GET("/profile", clientPortalHandler.Profile)
				clientPortal.GET("/cases", clientPortalHandler.Cases)
				clientPortal.GET("/documents", clientPortalHandler.Documents)
				clientPortal.GET("/meetings", clientPortalHandler.Meetings)
				clientPortal.GET("/invoices", clientPortalHandler.Invoices)
				clientPortal.GET("/invoices/:id", clientPortalHandler.GetInvoice)
				clientPortal.POST("/invoices/:id/payment-result", clientPortalHandler.SubmitInvoicePayment)
				clientPortal.GET("/invoices/:id/payment", clientPortalHandler.GetPayment)
				clientPortal.GET("/payments", clientPortalHandler.PaymentHistory)
			}

			// Hidden Admin Dashboard — never linked from any lawyer/client UI.
			// Gated to Role == "admin" on top of the JWT check every other
			// /api/v1 route already has. There is no signup/creation path for
			// this role anywhere in the API; an admin account can only be
			// made by updating the users table directly in the database.
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole(models.RoleAdmin))
			{
				admin.GET("/dashboard", adminHandler.DashboardStats)
				admin.GET("/lawyers", adminHandler.Lawyers)
				admin.GET("/clients", adminHandler.Clients)
				admin.GET("/cases", adminHandler.Cases)
				admin.GET("/payments", adminHandler.Payments)
			}
		}
	}
}
