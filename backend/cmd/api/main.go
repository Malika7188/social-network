package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Athooh/social-network/internal/auth"
	"github.com/Athooh/social-network/internal/config"
	"github.com/Athooh/social-network/internal/follow"
	"github.com/Athooh/social-network/internal/group"
	notifications "github.com/Athooh/social-network/internal/notifcations"
	"github.com/Athooh/social-network/internal/post"
	"github.com/Athooh/social-network/internal/profile"
	"github.com/Athooh/social-network/internal/server"
	wsHandler "github.com/Athooh/social-network/internal/websocket"
	"github.com/Athooh/social-network/pkg/db/sqlite"
	"github.com/Athooh/social-network/pkg/filestore"
	"github.com/Athooh/social-network/pkg/logger"
	"github.com/Athooh/social-network/pkg/websocket"

	"github.com/Athooh/social-network/internal/chat"
	"github.com/Athooh/social-network/internal/event"
	userHandler "github.com/Athooh/social-network/internal/user"
	"github.com/Athooh/social-network/pkg/session"
	"github.com/Athooh/social-network/pkg/user"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Println("Usage: go run cmd/api/main.go ")
		os.Exit(1)
	}

	// Load configuration
	cfg := config.Load()

	// Set up logger
	level := logger.INFO
	switch cfg.Log.Level {
	case "debug":
		level = logger.DEBUG
	case "info":
		level = logger.INFO
	case "warn":
		level = logger.WARN
	case "error":
		level = logger.ERROR
	case "fatal":
		level = logger.FATAL
	}

	log := logger.New(logger.Config{
		Level:       level,
		Output:      os.Stdout,
		TimeFormat:  cfg.Log.TimeFormat,
		ShowCaller:  cfg.Log.ShowCaller,
		FilePath:    cfg.Log.FilePath,
		EnableColor: cfg.Log.EnableColor,
		OutputType:  logger.BothOutput,
	})

	log.Info("Starting social network API server")

	// Set up database
	dbConfig := sqlite.Config{
		DBPath:         cfg.Database.Path,
		MigrationsPath: cfg.Database.MigrationsPath,
	}

	db, err := sqlite.New(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Info("Connected to database")

	if err := sqlite.CreateMigrations(db); err != nil {
		log.Fatal("Failed to create migrations: %v", err)
	}

	// Set up repositories
	userRepo := user.NewSQLiteRepository(db.DB)
	sessionRepo := session.NewSQLiteRepository(db.DB)
	postRepo := post.NewSQLiteRepository(db.DB)
	followRepo := follow.NewSQLiteRepository(db.DB)
	statusRepo := userHandler.NewSQLiteStatusRepository(db.DB)
	groupRepo := group.NewSQLiteRepository(db.DB)
	eventRepo := event.NewSQLiteRepository(db.DB)
	chatRepo := chat.NewSQLiteRepository(db.DB)
	profileRepo := profile.NewSQLiteRepository(db.DB)
	notificationsRepo := notifications.NewSQLiteRepository(db.DB)

	// Set up session manager
	sessionManager := session.NewSessionManager(
		sessionRepo,
		cfg.Auth.SessionCookieName,
		cfg.Auth.SessionCookieDomain,
		cfg.Auth.SessionCookieSecure,
		cfg.Auth.SessionMaxAge,
	)

	// Set up JWT configuration
	jwtConfig := auth.JWTConfig{
		SecretKey:     cfg.Auth.JWTSecretKey,
		TokenDuration: time.Duration(cfg.Auth.JWTTokenDuration) * time.Second,
		Issuer:        "social-network",
	}

	// Set up file store
	fileStore, err := filestore.New(cfg.FileStore.UploadDir)
	if err != nil {
		log.Fatal("Failed to create file store: %v", err)
	}

	// Set up WebSocket hub
	wsHub := websocket.NewHub(log)
	go wsHub.Run()

	// Set up services
	notificationsService := notifications.NewService(notificationsRepo, userRepo, log, wsHub)
	authService := auth.NewService(userRepo, sessionManager, jwtConfig, statusRepo)
	postNotificationSvc := post.NewNotificationService(wsHub, userRepo, notificationsService, log)
	postService := post.NewService(postRepo, fileStore, log, postNotificationSvc)
	statusService := userHandler.NewStatusService(statusRepo, sessionRepo, wsHub, log)
	eventService := event.NewService(eventRepo, fileStore, log, notificationsService, wsHub)
	groupService := group.NewService(groupRepo, fileStore, log, wsHub, notificationsService)
	chatService := chat.NewService(chatRepo, log, wsHub)
	followService := follow.NewService(followRepo, userRepo, statusRepo, notificationsService, log, wsHub)
	profileService := profile.NewService(profileRepo, "./data/uploads")

	// Connect the Hub to the StatusService
	wsHub.SetStatusUpdater(statusService)

	// Run status cleanup to ensure consistency between sessions and online status
	go statusService.CleanupUserStatuses()

	// Set up handlers
	authHandler := auth.NewHandler(authService, fileStore)
	postHandler := post.NewHandler(postService, log)
	wsHandler := wsHandler.NewHandler(wsHub, log, statusService)
	followHandler := follow.NewHandler(followService, log)
	groupHandler := group.NewHandler(groupService, log)
	eventHandler := event.NewHandler(eventService, log)
	chatHandler := chat.NewHandler(chatService, log)
	notificationHanler := notifications.NewHandler(notificationsService, log)
	profileHandler := profile.NewHandler(profileService, log)

	// Set up router with both session and JWT middleware
	router := server.Router(server.RouterConfig{
		AuthHandler:         authHandler,
		PostHandler:         postHandler,
		WSHandler:           wsHandler,
		FollowHandler:       followHandler,
		GroupHandler:        groupHandler,
		EventHandler:        eventHandler,
		ChatHandler:         chatHandler,
		NotificationHanlder: notificationHanler,
		AuthMiddleware:      authService.RequireAuth,
		JWTMiddleware:       authService.RequireJWTAuth,
		Logger:              log,
		UploadDir:           cfg.FileStore.UploadDir,
		ProfileHandler:      profileHandler,
	})

	// Set up server
	serverConfig := server.Config{
		Host:         cfg.Server.Host,
		Port:         cfg.Server.Port,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	srv := server.New(serverConfig, router, log)

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatal("Server error: %v", err)
	}
}
