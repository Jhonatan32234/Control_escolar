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
	uService := services.NewUsuarioService(uRepo, moodleClient, aRepo)
	uHandler := NewUsuarioHandler(uService)

	// --- GRUPO ---
	gRepo := repository.NewGrupoRepository(db)
	gService := services.NewGrupoService(gRepo, moodleClient, aRepo, uRepo)
	gHandler := NewGrupoHandler(gService)

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
		r.Post("/enrol/{usuarioID}/{asignaturaID}", uHandler.MatricularUsuario)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", uHandler.GetUsuarioByID) 
			r.Put("/", uHandler.UpdateUsuario)   
			r.Delete("/", uHandler.DeleteUsuario)
		})
	})


	r.Route("/grupo", func(r chi.Router) {
		r.Post("/", gHandler.CreateGrupo) 
		r.Get("/", gHandler.GetAllGrupo)
		r.Post("/sync/{id}",gHandler.SyncGrupo)
		r.Post("/add-members/{grupoID}", gHandler.AddMembersToGroup) 
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", gHandler.GetGrupoByID) 
			r.Put("/", gHandler.UpdateGrupo)   
			r.Delete("/", gHandler.DeleteGrupo)
		})
	})



	return r
}