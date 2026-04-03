package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/williamprado/foto-magica-profissional/internal/auth"
	"github.com/williamprado/foto-magica-profissional/internal/billing"
	"github.com/williamprado/foto-magica-profissional/internal/config"
	"github.com/williamprado/foto-magica-profissional/internal/credits"
	"github.com/williamprado/foto-magica-profissional/internal/generation"
	"github.com/williamprado/foto-magica-profissional/internal/middleware"
	"github.com/williamprado/foto-magica-profissional/internal/notifications"
	"github.com/williamprado/foto-magica-profissional/internal/providers/ai"
	"github.com/williamprado/foto-magica-profissional/internal/providers/payment"
	"github.com/williamprado/foto-magica-profissional/internal/providers/storage"
	"github.com/williamprado/foto-magica-profissional/internal/tenant"
)

type Deps struct {
	Config        config.Config
	Logger        *slog.Logger
	DB            *pgxpool.Pool
	Storage       storage.Provider
	AI            ai.Provider
	Payment       map[string]payment.Provider
	Notifications notifications.Service
}

func NewRouter(deps Deps) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(deps.Logger))
	router.Use(cors.New(cors.Config{
		AllowOrigins:     deps.Config.AllowedOrigins,
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Tenant-ID"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok"}})
	})

	jwtManager := auth.NewJWTManager(deps.Config.JWTSecret, deps.Config.JTTTL)
	tenantRepo := tenant.NewRepository(deps.DB)
	authService := auth.NewService(deps.DB, tenantRepo, jwtManager, deps.Config.DefaultPlanCode)
	creditService := credits.NewService(deps.DB)
	billingService := billing.NewService(deps.DB, creditService, deps.Payment, "mock")
	generationService := generation.NewService(deps.DB, creditService, deps.Storage, deps.AI, deps.Notifications)

	registerAuthRoutes(router, authService, jwtManager)
	registerBillingRoutes(router, billingService, middleware.Auth(jwtManager))
	registerGenerationRoutes(router, generationService, creditService, middleware.Auth(jwtManager), middleware.RateLimit(deps.Config.GenerationRateLimitRPM))

	router.GET("/storage/*key", func(c *gin.Context) {
		key := c.Param("key")
		key = strings.TrimPrefix(key, "/")
		path := filepath.Join(deps.Config.LocalStoragePath, key)
		if _, err := os.Stat(path); err != nil {
			Fail(c, NewError(404, "not_found", "asset not found"))
			return
		}
		c.File(path)
	})

	return router
}

func registerAuthRoutes(router *gin.Engine, authService auth.Service, jwtManager auth.JWTManager) {
	router.POST("/v1/auth/register", func(c *gin.Context) {
		var input auth.RegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			Fail(c, NewError(400, "invalid_payload", err.Error()))
			return
		}
		session, err := authService.Register(c.Request.Context(), input)
		if err != nil {
			Fail(c, err)
			return
		}
		OK(c, http.StatusCreated, session)
	})

	router.POST("/v1/auth/login", func(c *gin.Context) {
		var input auth.LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			Fail(c, NewError(400, "invalid_payload", err.Error()))
			return
		}
		session, err := authService.Login(c.Request.Context(), input)
		if err != nil {
			Fail(c, err)
			return
		}
		OK(c, http.StatusOK, session)
	})

	authorized := router.Group("/v1")
	authorized.Use(middleware.Auth(jwtManager), middleware.TenantScoped())
	authorized.GET("/profile/me", func(c *gin.Context) {
		claims := middleware.MustUser(c)
		session, err := authService.Me(c.Request.Context(), auth.SessionClaims{
			UserID:   claims.UserID,
			TenantID: claims.TenantID,
			Role:     claims.Role,
			Email:    claims.Email,
		})
		if err != nil {
			Fail(c, err)
			return
		}
		OK(c, http.StatusOK, session)
	})
}

func registerBillingRoutes(router *gin.Engine, service billing.Service, authMiddleware gin.HandlerFunc) {
	router.GET("/v1/billing/plans", func(c *gin.Context) {
		plans, err := service.Plans(c.Request.Context())
		if err != nil {
			Fail(c, err)
			return
		}
		OK(c, http.StatusOK, plans)
	})

	authorized := router.Group("/v1")
	authorized.Use(authMiddleware, middleware.TenantScoped())
	authorized.POST("/billing/checkout", func(c *gin.Context) {
		claims := middleware.MustUser(c)
		var payload struct {
			PlanCode string `json:"planCode" binding:"required"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			Fail(c, NewError(http.StatusBadRequest, "invalid_payload", err.Error()))
			return
		}
		session, err := service.CreateCheckout(c.Request.Context(), claims.TenantID, payload.PlanCode)
		if err != nil {
			Fail(c, err)
			return
		}
		OK(c, http.StatusCreated, session)
	})

	router.POST("/webhooks/:provider", func(c *gin.Context) {
		raw, err := c.GetRawData()
		if err != nil {
			Fail(c, err)
			return
		}
		headers := map[string]string{}
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		if err := service.HandleWebhook(c.Request.Context(), c.Param("provider"), headers, raw); err != nil {
			Fail(c, err)
			return
		}
		OK(c, http.StatusOK, gin.H{"received": true})
	})
}

func registerGenerationRoutes(router *gin.Engine, service generation.Service, creditService credits.Service, authMiddleware gin.HandlerFunc, rateLimit gin.HandlerFunc) {
	group := router.Group("/v1")
	group.Use(authMiddleware, middleware.TenantScoped())
	group.GET("/generation/jobs", func(c *gin.Context) {
		claims := middleware.MustUser(c)
		jobs, err := service.ListJobs(c.Request.Context(), claims.TenantID)
		if err != nil {
			Fail(c, err)
			return
		}
		OK(c, http.StatusOK, jobs)
	})
	group.POST("/generation/jobs", rateLimit, func(c *gin.Context) {
		claims := middleware.MustUser(c)
		var input generation.CreateJobInput
		if err := c.ShouldBindJSON(&input); err != nil {
			Fail(c, NewError(http.StatusBadRequest, "invalid_payload", err.Error()))
			return
		}
		job, err := service.CreateJob(c.Request.Context(), claims.TenantID, claims.UserID, input)
		if err != nil {
			switch {
			case errors.Is(err, generation.ErrInvalidReferenceImage):
				Fail(c, NewError(http.StatusBadRequest, "invalid_reference_image", err.Error()))
			case errors.Is(err, generation.ErrInvalidUserImage):
				Fail(c, NewError(http.StatusBadRequest, "invalid_user_image", err.Error()))
			case errors.Is(err, credits.ErrInsufficientCredits):
				Fail(c, NewError(http.StatusPaymentRequired, "insufficient_credits", err.Error()))
			default:
				Fail(c, err)
			}
			return
		}
		OK(c, http.StatusCreated, job)
	})
	group.GET("/gallery", func(c *gin.Context) {
		claims := middleware.MustUser(c)
		jobs, err := service.ListJobs(c.Request.Context(), claims.TenantID)
		if err != nil {
			Fail(c, err)
			return
		}
		var gallery []map[string]any
		for _, job := range jobs {
			if job.ResultURL == "" {
				continue
			}
			gallery = append(gallery, map[string]any{
				"id":         job.ID,
				"title":      firstPrompt(job.PromptSections),
				"previewUrl": job.ResultURL,
				"favorite":   false,
				"createdAt":  job.CreatedAt,
			})
		}
		OK(c, http.StatusOK, gallery)
	})
	group.GET("/credits/wallet", func(c *gin.Context) {
		claims := middleware.MustUser(c)
		wallet, err := creditService.Wallet(c.Request.Context(), claims.TenantID)
		if err != nil {
			Fail(c, err)
			return
		}
		OK(c, http.StatusOK, wallet)
	})
}

func firstPrompt(sections []ai.PromptSection) string {
	if len(sections) == 0 {
		return "Nova geração"
	}
	return sections[0].Content
}

func Serve(ctx context.Context, cfg config.Config, handler http.Handler) error {
	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}
