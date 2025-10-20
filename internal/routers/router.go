package routers

import (
	"xyz-football/internal/handlers"
	"xyz-football/internal/middleware"
	"xyz-football/internal/repositories"
	"xyz-football/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Initialize repositories
	repo := struct {
		team   repositories.TeamRepository
		player repositories.PlayerRepository
		match  repositories.MatchRepository
		goal   repositories.GoalRepository
		admin  repositories.AdminRepository
	}{
		team:   repositories.NewTeamRepository(db),
		player: repositories.NewPlayerRepository(db),
		match:  repositories.NewMatchRepository(db),
		goal:   repositories.NewGoalRepository(db),
		admin:  repositories.NewAdminRepository(db),
	}

	// Initialize services
	svc := struct {
		team   services.TeamService
		player services.PlayerService
		match  services.MatchService
		report services.ReportService
		admin  services.AdminService
	}{
		team:   services.NewTeamService(repo.team),
		player: services.NewPlayerService(repo.player),
		match:  services.NewMatchService(repo.match, repo.goal),
		report: services.NewReportService(repo.match, repo.team),
		admin:  services.NewAdminService(repo.admin),
	}

	// Initialize handlers
	h := struct {
		team   *handlers.TeamHandler
		player *handlers.PlayerHandler
		match  *handlers.MatchHandler
		report *handlers.ReportHandler
		admin  *handlers.AdminHandler
	}{
		team:   handlers.NewTeamHandler(svc.team),
		player: handlers.NewPlayerHandler(svc.player),
		match:  handlers.NewMatchHandler(svc.match),
		report: handlers.NewReportHandler(svc.report),
		admin:  handlers.NewAdminHandler(svc.admin),
	}

	// Public routes (no authentication required)
	public := r.Group("/api/v1")
	{
		// Authentication endpoints
		auth := public.Group("/admin")
		{
			auth.POST("/login", h.admin.Login)
			auth.POST("/register", h.admin.Register)
		}

	}

	// Protected routes (require authentication)
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuthMiddleware())
	{
		// Team management
		teams := api.Group("/teams")
		{
			teams.GET("", h.team.List)
			teams.GET("/:id", h.team.Get)
			teams.POST("", h.team.Create)
			teams.PUT("/:id", h.team.Update)
			teams.DELETE("/:id", h.team.Delete)
		}

		// Player management
		players := api.Group("/players")
		{
			players.GET("", h.player.List)
			players.GET("/:id", h.player.Get)
			players.POST("", h.player.Create)
			players.PUT("/:id", h.player.Update)
			players.DELETE("/:id", h.player.Delete)
			players.GET("/by-team/:teamId", h.player.ListByTeam)
		}

		// Match management
		matches := api.Group("/matches")
		{
			matches.GET("", h.match.List)
			matches.GET("/:id", h.match.Get)
			matches.GET("/by-team/:teamId", h.match.GetByTeam)
			matches.POST("", h.match.Create)
			matches.PUT("/:id", h.match.Update)
			matches.DELETE("/:id", h.match.Delete)
			matches.POST("/:id/report", h.match.ReportResult)
		}

		reports := api.Group("/reports")
		{
			reports.GET("/standings", h.report.GetStandings)
			reports.GET("/top-scorers", h.report.GetTopScorers)
			reports.GET("/matches/:id", h.report.GetMatchReport)
		}

		// Admin management
		// admin := api.Group("/admin")
		// {
		// 	// Add admin-only endpoints here if needed
		// }
	}

	return r
}
