package handlers

import (
	"api_concurrencia/src/models"
	"api_concurrencia/src/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type GrupoHandler struct {
	Service *services.GrupoService
}

func NewGrupoHandler(s *services.GrupoService) *GrupoHandler {
	return &GrupoHandler{Service: s}
}

// CreateGrupo maneja la creación local del grupo y su sincronización a Moodle. (POST /grupo)
// @Summary Crear Grupo
// @Description Crea un grupo local
// @Tags grupo
// @Accept json
// @Produce json
// @Param grupo body models.Grupo true "Datos del grupo"
// @Success 201 {object} models.Grupo
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /grupo/ [post]
func (h *GrupoHandler) CreateGrupo(w http.ResponseWriter, r *http.Request) {
	var g models.Grupo
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Crear el grupo en la base de datos local (usando servicio para validaciones)
	if err := h.Service.CreateLocal(&g); err != nil {
		http.Error(w, "Error al crear Grupo local: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

// @Summary Listar Grupos
// @Description Obtiene todos los grupos
// @Tags grupo
// @Produce json
// @Success 200 {array} models.Grupo
// @Failure 500 {string} string
// @Router /grupo/ [get]
func (h *GrupoHandler) GetAllGrupo(w http.ResponseWriter, r *http.Request) {
	grupos, err := h.Service.GetAll()
	if err != nil {
		http.Error(w, "Error al obtener Grupos: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(grupos)
}

// @Summary Obtener Grupo
// @Description Obtiene un grupo por ID
// @Tags grupo
// @Produce json
// @Param id path int true "ID del grupo"
// @Success 200 {object} models.Grupo
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /grupo/{id}/ [get]
func (h *GrupoHandler) GetGrupoByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	pe, err := h.Service.GetByID(uint(id))
	if err != nil {
		http.Error(w, "Grupo no encontrado: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pe)
}

// @Summary Actualizar Grupo
// @Description Actualiza un grupo por ID
// @Tags grupo
// @Accept json
// @Produce json
// @Param id path int true "ID del grupo"
// @Param grupo body models.Grupo true "Datos del grupo"
// @Success 200 {object} models.Grupo
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /grupo/{id}/ [put]
func (h *GrupoHandler) UpdateGrupo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var pe models.Grupo
	if err := json.NewDecoder(r.Body).Decode(&pe); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pe.ID = uint(id) // Asegurar que se actualice el registro correcto

	if err := h.Service.UpdateLocal(&pe); err != nil {
		http.Error(w, "Error al actualizar Grupo local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pe)
}

// DeleteProgramaEstudio maneja la eliminación local.
// @Summary Eliminar Grupo
// @Description Elimina un grupo por ID
// @Tags grupo
// @Param id path int true "ID del grupo"
// @Success 204 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /grupo/{id}/ [delete]
func (h *GrupoHandler) DeleteGrupo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteLocal(uint(id)); err != nil {
		http.Error(w, "Error al eliminar Grupo local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Sincronizar Grupo
// @Description Inicia la sincronización del grupo a Moodle
// @Tags grupo
// @Param id path int true "ID del grupo"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /grupo/sync/{id} [post]
func (h *GrupoHandler) SyncGrupo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := h.Service.SyncToMoodle(uint(id)); err != nil {
		http.Error(w, "Error durante la sincronización: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sincronización iniciada correctamente."))
}

// AddMembersToGroup maneja la adición de miembros a un grupo local y su sincronización a Moodle.
// POST /grupo/add-members/{grupoID}
// @Summary Añadir Miembros a Grupo
// @Description Añade miembros al grupo y sincroniza con Moodle
// @Tags grupo
// @Accept json
// @Produce plain
// @Param grupoID path int true "ID del grupo"
// @Param usuarios body []uint true "IDs de usuarios"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /grupo/add-members/{grupoID} [post]
func (h *GrupoHandler) AddMembersToGroup(w http.ResponseWriter, r *http.Request) {
	grupoIDStr := chi.URLParam(r, "grupoID")
	grupoID, err := strconv.ParseUint(grupoIDStr, 10, 32)
	if err != nil {
		http.Error(w, "ID de Grupo inválido", http.StatusBadRequest)
		return
	}

	// Esperamos un array de IDs de usuario en el cuerpo
	var usuarioIDs []uint
	if err := json.NewDecoder(r.Body).Decode(&usuarioIDs); err != nil {
		http.Error(w, "Error al decodificar IDs de usuarios: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(usuarioIDs) == 0 {
		http.Error(w, "Debe proporcionar al menos un ID de usuario.", http.StatusBadRequest)
		return
	}

	// 1. Añadir miembros en la tabla de unión local (Many-to-Many)
	if err := h.Service.Repo.AddMembers(uint(grupoID), usuarioIDs); err != nil {
		http.Error(w, "Error al añadir miembros localmente: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Iniciar Sincronización Asíncrona con Moodle
	// Esto sólo funciona si el grupo ya está sincronizado.
	go func(id uint) {
		if err := h.Service.SyncMembersToMoodle(id); err != nil {
			log.Printf("ERROR ASÍNCRONO al añadir miembros a Moodle para Grupo ID %d: %v", id, err)
		} else {
			log.Printf("✅ Sincronización asíncrona de miembros para Grupo ID %d finalizada.", id)
		}
	}(uint(grupoID))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Miembros añadidos localmente e iniciada sincronización a Moodle para Grupo ID %d.", grupoID)))
}

// ... (Aquí podrías añadir GetByID, GetAll, etc. si fueran necesarios)
