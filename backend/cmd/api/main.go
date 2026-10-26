package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"orenkod/backend/internal/athletes"
	"orenkod/backend/internal/auth"
	"orenkod/backend/internal/competitions"
	"orenkod/backend/internal/config"
	"orenkod/backend/internal/dashboard"
	"orenkod/backend/internal/database"
	"orenkod/backend/internal/httpserver"
	"orenkod/backend/internal/integrations/max"
	"orenkod/backend/internal/notifications"
	"orenkod/backend/internal/platform"
	"orenkod/backend/internal/ratings"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}
	defer db.Close()

	competitionsRepo := competitions.NewRepository(db)
	competitionsService := competitions.NewService(competitionsRepo)
	athletesRepo := athletes.NewRepository(db)
	ratingsRepo := ratings.NewRepository(db)
	dashboardRepo := dashboard.NewRepository(db, competitionsRepo, ratingsRepo)
	maxClient := max.NewClient(cfg)
	authService := auth.New(db)
	platformHandler := platform.New(db, authService, competitionsService, athletesRepo, maxClient)

	competitionsHandler := competitions.NewHandler(competitionsService, maxClient)
	athletesHandler := athletes.NewHandler(athletesRepo)
	ratingsHandler := ratings.NewHandler(ratingsRepo)
	dashboardHandler := dashboard.NewHandler(db, dashboardRepo)
	notificationsHandler := notifications.NewHandler()
	maxHandler := max.NewWebhookHandler(cfg.MaxWebhookSecret, maxClient, competitionsService, ratingsRepo)

	handler := httpserver.NewRouter(httpserver.Routes{
		Health:            dashboardHandler.Health,
		Dashboard:         dashboardHandler.Get,
		Competitions:      competitionsHandler.List,
		CreateCompetition: platformHandler.CreateCompetition,
		Athletes:          athletesHandler.List,
		Rating:            ratingsHandler.List,
		Team:              ratingsHandler.Team,
		Notifications:     notificationsHandler.List,
		MaxStatus:         maxHandler.Status,
		MaxWebhook:        maxHandler.Receive,
		Auth:              authService,
		Platform: httpserver.PlatformRoutes{
			Login:             platformHandler.Login,
			Register:          platformHandler.Register,
			Logout:            platformHandler.Logout,
			Me:                platformHandler.Me,
			Switch:            platformHandler.SwitchRole,
			CreateCompetition: platformHandler.CreateCompetition,
			ApproveAthlete:    platformHandler.ApproveAthlete,
			Catalogs:          platformHandler.Catalogs,
			PublicHome:        platformHandler.PublicHome,
			Verify:            platformHandler.VerifyCertificate,
			Publish:           platformHandler.PublishCompetition,
			Apply:             platformHandler.Apply,
			Applications:      platformHandler.Applications,
			DecideApp:         platformHandler.DecideApplication,
			RequestLink:       platformHandler.RequestLink,
			DecideLink:        platformHandler.DecideLink,
			Links:             platformHandler.Links,
			Judges:            platformHandler.Judges,
			AssignJudge:       platformHandler.AssignJudge,
			Assignments:       platformHandler.Assignments,
			SaveResults:       platformHandler.SaveResults,
			Protocol:          platformHandler.Protocol,
			History:           platformHandler.ResultsHistory,
			RankApply:         platformHandler.RankApply,
			RankList:          platformHandler.RankList,
			RankDecide:        platformHandler.RankDecide,
			CatApply:          platformHandler.CategoryApply,
			CatList:           platformHandler.CategoryList,
			CatDecide:         platformHandler.CategoryDecide,
			TeamSave:          platformHandler.TeamSave,
			TeamOrder:         platformHandler.TeamOrder,
			Enrollment:        platformHandler.Enrollment,
			AthleteReport:     platformHandler.AthleteReport,
			RegistryXLSX:      platformHandler.RegistryXLSX,
			AddType:           platformHandler.AddType,
			Development:       platformHandler.Development,
			Audit:             platformHandler.Audit,
			Certificate:       platformHandler.IssueCertificate,
			Inbox:             platformHandler.Inbox,
			Rating:            platformHandler.Rating,
			TeamGet:           platformHandler.TeamGet,
		},
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      45 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("ОренКод API слушает :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
}
