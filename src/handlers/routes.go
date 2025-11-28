package handlers

import (
	"api_concurrencia/src/moodle"
	"api_concurrencia/src/repository"
	"api_concurrencia/src/services"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func Routes(db *gorm.DB, moodleClient *moodle.Client) *chi.Mux {
	r := chi.NewRouter()

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
	uService := services.NewUsuarioService(uRepo)
	uHandler := NewUsuarioHandler(uService)

	r.Route("/programa-estudio", func(r chi.Router) {
		r.Post("/", peHandler.CreateProgramaEstudio)
		r.Get("/", peHandler.GetAllProgramaEstudio)
		r.Post("/sync/{id}", peHandler.SyncProgramaEstudio) 
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", peHandler.GetProgramaEstudioByID)
			r.Put("/", peHandler.UpdateProgramaEstudio)
			r.Delete("/", peHandler.DeleteProgramaEstudio)
		})
	})
    
	r.Route("/cuatrimestre", func(r chi.Router) {
		r.Post("/", cHandler.CreateCuatrimestre)
		r.Get("/", cHandler.GetAllCuatrimestres)
		r.Post("/sync/{id}", cHandler.SyncCuatrimestre) 
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", cHandler.GetCuatrimestreByID)
			r.Put("/", cHandler.UpdateCuatrimestre)
			r.Delete("/", cHandler.DeleteCuatrimestre)
		})
	})

	r.Route("/asignatura", func(r chi.Router) {
		r.Post("/", aHandler.CreateAsignatura)
		r.Get("/", aHandler.GetAllAsignaturas) // Asumiendo que implementaste GetAll
		r.Post("/sync/{id}", aHandler.SyncAsignatura) 
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", aHandler.GetAsignaturaByID) // Asumiendo que implementaste GetByID
			r.Put("/", aHandler.UpdateAsignatura)   // Asumiendo que implementaste Update
			r.Delete("/", aHandler.DeleteAsignatura) // Asumiendo que implementaste Delete
		})
	})

	r.Route("/usuario", func(r chi.Router) {
		r.Post("/", uHandler.CreateUsuario)
		r.Get("/", uHandler.GetAllUsuarios) 
		r.Post("/sync/{id}", uHandler.SyncUsuario)
		r.Post("/bulk-sync", uHandler.BulkSyncUsuarios) // RUTA MASIVA
		
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", uHandler.GetUsuarioByID) 
			r.Put("/", uHandler.UpdateUsuario)   
			r.Delete("/", uHandler.DeleteUsuario)
		})
	})
	return r
}