package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/application"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/project"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/user"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/infra/http/handler"
	custommiddleware "github.com/AugustoGuapo/concretrack-backoffice-be/internal/infra/http/middleware"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/infra/storage"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
    // --- Inicialización de base de datos ---
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file %v", err)
    }

    db, err := sqlx.Open("sqlite", fmt.Sprintf("file:%s?cache=shared&_fk=1", os.Getenv("DB_PATH")))
    if err != nil {
        log.Fatalf("error al abrir la base de datos: %v", err)
    }
    defer db.Close()

    // --- Inyección de dependencias ---
    userRepo := storage.NewUserRepository(db)
    userService := user.NewService(userRepo)
    authHandler := handler.NewAuthHandler(userService)

    projectRepo := storage.NewProjectRepository(db)
    projectService := project.NewService(projectRepo)
    projectHandler := handler.NewProjectHandler(projectService)

    reportsService := application.NewReportsService(projectRepo)
    reportsHandler := handler.NewReportsHandler(*reportsService)

    // --- Router Chi ---
    r := chi.NewRouter()

    // Middlewares globales
    r.Use(middleware.Logger)          // log bonito
    r.Use(middleware.Recoverer)       // recupera de panics
    r.Use(custommiddleware.CORS)      // tu middleware de CORS

    // --- Rutas ---
    r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
        authHandler.Login(w, r)
    })

    r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        response := map[string]string{"response": "pong"}
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    })

    // Grupo de proyectos
    r.Route("/projects", func(r chi.Router) {

        r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
            ID := chi.URLParam(r, "id")
			numericID, err := strconv.Atoi(ID)
			if err != nil {
				http.Error(w, "project id must be numeric", http.StatusBadRequest)
			}
            projectHandler.GetProjectByID(w, r, numericID)
        })

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			pageStr := r.URL.Query().Get("page")
			if pageStr == "" {
				// Por defecto puedes poner 1 o lanzar error
				pageStr = "1"
			}

			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				http.Error(w, "invalid page", http.StatusBadRequest)
				return
			}

			projectHandler.GetProjects(w, r, page)
		})

        r.Get("/{ID}/families/{familyID}/report", func(w http.ResponseWriter, r *http.Request) {


            reportsHandler.GenerateReportForOneFamily(w, r)
        })
    })

    // --- Inicio del servidor ---
    log.Println("Servidor corriendo en http://localhost:8080")
    if err := http.ListenAndServe("0.0.0.0:8080", r); err != nil {
        log.Fatalf("error al iniciar servidor: %v", err)
    }
}
