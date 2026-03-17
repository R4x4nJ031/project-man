package main
import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/R4x4nJ031/project-man/internal/config"
	"github.com/R4x4nJ031/project-man/internal/database"
	"github.com/R4x4nJ031/project-man/internal/health"
	"github.com/R4x4nJ031/project-man/internal/logger"
	"github.com/R4x4nJ031/project-man/internal/middleware"
	"github.com/R4x4nJ031/project-man/internal/project"
	"github.com/R4x4nJ031/project-man/internal/server"
)

func main() {

	cfg := config.Load()
	err := database.Connect()
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	log := logger.New()

	mux := http.NewServeMux()

	// Public route
	mux.HandleFunc("/healthz", health.Handler)
	// Protected route (MUST be before server start)
	mux.Handle("/secure",
		middleware.AuthMiddleware(cfg)(
			middleware.TenantMiddleware(
				middleware.RequirePermission(middleware.CreateProject)(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("authorized project creation"))
					}),
				),
			),
		),
	)
	mux.Handle("/projects",
		middleware.AuthMiddleware(cfg)(
			middleware.TenantMiddleware(
				middleware.RequirePermission(middleware.CreateProject)(
					http.HandlerFunc(project.CreateProjectHandler),
				),
			),
		),
	)

	mux.Handle("/projects/list",
		middleware.AuthMiddleware(cfg)(
			middleware.TenantMiddleware(
				middleware.RequirePermission(middleware.ReadProject)(
					http.HandlerFunc(project.ListProjectsHandler),
				),
			),
		),
	)
	mux.Handle("/projects/get",
		middleware.AuthMiddleware(cfg)(
			middleware.TenantMiddleware(
				middleware.RequirePermission(middleware.ReadProject)(
					http.HandlerFunc(project.GetProjectHandler),
				),
			),
		),
	)

	mux.Handle("/projects/update",
		middleware.AuthMiddleware(cfg)(
			middleware.TenantMiddleware(
				middleware.RequirePermission(middleware.UpdateProject)(
					http.HandlerFunc(project.UpdateProjectHandler),
				),
			),
		),
	)

	mux.Handle("/projects/delete",
		middleware.AuthMiddleware(cfg)(
			middleware.TenantMiddleware(
				middleware.RequirePermission(middleware.DeleteProject)(
					http.HandlerFunc(project.DeleteProjectHandler),
				),
			),
		),
	)

	srv := server.New(":"+cfg.Port, mux)

	go func() {
		log.Println("Server starting on port", cfg.Port)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
