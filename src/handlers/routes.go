package handlers

import (
	//"api_concurrencia/src/middleware"
	"api_concurrencia/src/moodle"
	"api_concurrencia/src/repository"
	"api_concurrencia/src/services"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"gorm.io/gorm"
)

func Routes(db *gorm.DB, moodleClient *moodle.Client) *chi.Mux {
	r := chi.NewRouter()

	// Configuración de CORS
	allowedOrigins := []string{"https://*", "http://*"} // Valor por defecto
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Inicialización de Repositorios, Servicios y Handlers

	// --- PROGRAMA ESTUDIO (PE) ---
	peRepo := repository.NewProgramaEstudioRepository(db)
	peService := services.NewProgramaEstudioService(peRepo, moodleClient)
	peHandler := NewProgramaEstudioHandler(peService)

	// --- CUATRIMESTRE ---
	cRepo := repository.NewCuatrimestreRepository(db)
	cService := services.NewCuatrimestreService(cRepo, moodleClient)
	cHandler := NewCuatrimestreHandler(cService)

	// --- ASIGNATURA ---
	aRepo := repository.NewAsignaturaRepository(db)
	aService := services.NewAsignaturaService(aRepo, moodleClient)
	aHandler := NewAsignaturaHandler(aService)

	// --- USUARIO ---
	uRepo := repository.NewUsuarioRepository(db)
	uService := services.NewUsuarioService(uRepo, moodleClient, aRepo)
	uHandler := NewUsuarioHandler(uService)

	// --- AUTH ---
	authHandler := NewAuthHandler(uService)

	// --- GRUPO ---
	gRepo := repository.NewGrupoRepository(db)
	gService := services.NewGrupoService(gRepo, moodleClient, aRepo, uRepo)
	gHandler := NewGrupoHandler(gService)

	// Rutas públicas (sin autenticación)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Rutas protegidas (requieren autenticación)
	r.Group(func(r chi.Router) {
		//r.Use(middleware.AuthMiddleware)

		r.Route("/programa-estudio", func(r chi.Router) {
			r.Post("/", peHandler.CreateProgramaEstudio)
			r.Get("/", peHandler.GetAllProgramaEstudio)
			r.Post("/sync/{id}", peHandler.SyncProgramaEstudio)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", peHandler.GetProgramaEstudioByID)
				//r.Put("/", peHandler.UpdateProgramaEstudio)
				//r.Delete("/", peHandler.DeleteProgramaEstudio)
			})
		})

		r.Route("/cuatrimestre", func(r chi.Router) {
			r.Post("/", cHandler.CreateCuatrimestre)
			r.Get("/", cHandler.GetAllCuatrimestres)
			r.Post("/sync/{id}", cHandler.SyncCuatrimestre)
			r.Post("/bulk-sync", cHandler.BulkSyncCuatrimestres)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", cHandler.GetCuatrimestreByID)
				r.Put("/", cHandler.UpdateCuatrimestre)
				r.Delete("/", cHandler.DeleteCuatrimestre)
			})
		})

		r.Route("/asignatura", func(r chi.Router) {
			r.Post("/", aHandler.CreateAsignatura)
			r.Get("/", aHandler.GetAllAsignaturas)
			r.Post("/sync/{id}", aHandler.SyncAsignatura)
			r.Post("/bulk-sync", aHandler.BulkSyncAsignaturas)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", aHandler.GetAsignaturaByID)
				r.Put("/", aHandler.UpdateAsignatura)
				r.Delete("/", aHandler.DeleteAsignatura)
			})
		})

		r.Route("/usuario", func(r chi.Router) {
			r.Post("/", uHandler.CreateUsuario)
			r.Get("/", uHandler.GetAllUsuarios)
			r.Get("/unsynced", uHandler.GetUnsyncedUsuarios)
			r.Get("/by_group/{grupoID}", uHandler.GetUsuariosByGroupID)
			r.Post("/sync/{id}", uHandler.SyncUsuario)
			r.Post("/bulk-sync", uHandler.BulkSyncUsuarios)
			r.Post("/enrol/{usuarioID}/{asignaturaID}", uHandler.MatricularUsuario)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", uHandler.GetUsuarioByID)
			})
		})

		r.Route("/grupo", func(r chi.Router) {
			r.Post("/", gHandler.CreateGrupo)
			r.Get("/", gHandler.GetAllGrupo)
			r.Post("/sync/{id}", gHandler.SyncGrupo)
			r.Post("/bulk-sync", gHandler.BulkSyncGrupos)
			r.Post("/add-members/{grupoID}", gHandler.AddMembersToGroup)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", gHandler.GetGrupoByID)
				r.Put("/", gHandler.UpdateGrupo)
				r.Delete("/", gHandler.DeleteGrupo)
			})
		})
	})

	return r
}
