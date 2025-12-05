package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api_concurrencia/src/models"
	"api_concurrencia/src/services"

	"github.com/go-chi/chi/v5"
)

type AsignaturaHandler struct {
	Service *services.AsignaturaService
}

func NewAsignaturaHandler(s *services.AsignaturaService) *AsignaturaHandler {
	return &AsignaturaHandler{Service: s}
}

// CreateAsignatura maneja la creación local. (POST /asignatura)
// @Summary Crear Asignatura
// @Description Crea una asignatura local
// @Tags asignatura
// @Accept json
// @Produce json
// @Param asignatura body models.Asignatura true "Datos de la asignatura"
// @Success 201 {object} models.Asignatura
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asignatura/ [post]
func (h *AsignaturaHandler) CreateAsignatura(w http.ResponseWriter, r *http.Request) {
	var a models.Asignatura
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateLocal(&a); err != nil {
		http.Error(w, "Error al crear Asignatura local: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

// @Summary Obtener Asignatura
// @Description Obtiene una asignatura por ID
// @Tags asignatura
// @Produce json
// @Param id path int true "ID de la asignatura"
// @Success 200 {object} models.Asignatura
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /asignatura/{id}/ [get]
func (h *AsignaturaHandler) GetAsignaturaByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	c, err := h.Service.GetByID(uint(id))
	if err != nil {
		http.Error(w, "Asignatura no encontrado: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

// @Summary Listar Asignaturas
// @Description Obtiene todas las asignaturas
// @Tags asignatura
// @Produce json
// @Success 200 {array} models.Asignatura
// @Failure 500 {string} string
// @Router /asignatura/ [get]
func (h *AsignaturaHandler) GetAllAsignaturas(w http.ResponseWriter, r *http.Request) {
	asignaturas, err := h.Service.GetAll()
	if err != nil {
		http.Error(w, "Error al obtener Asignaturas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(asignaturas)
}

// @Summary Actualizar Asignatura
// @Description Actualiza una asignatura por ID
// @Tags asignatura
// @Accept json
// @Produce json
// @Param id path int true "ID de la asignatura"
// @Param asignatura body models.Asignatura true "Datos de la asignatura"
// @Success 200 {object} models.Asignatura
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asignatura/{id}/ [put]
func (h *AsignaturaHandler) UpdateAsignatura(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var c models.Asignatura
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.ID = uint(id)

	if err := h.Service.UpdateLocal(&c); err != nil {
		http.Error(w, "Error al actualizar Asignatura local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

// DeleteCuatrimestre maneja la eliminación local. (DELETE /cuatrimestre/{id})
// @Summary Eliminar Asignatura
// @Description Elimina una asignatura por ID
// @Tags asignatura
// @Param id path int true "ID de la asignatura"
// @Success 204 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asignatura/{id}/ [delete]
func (h *AsignaturaHandler) DeleteAsignatura(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteLocal(uint(id)); err != nil {
		http.Error(w, "Error al eliminar Asignatura local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SyncAsignatura maneja la solicitud de sincronización. (POST /asignatura/sync/{id})
// @Summary Sincronizar Asignatura
// @Description Inicia la sincronización de una asignatura a Moodle
// @Tags asignatura
// @Param id path int true "ID de la asignatura"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Router /asignatura/sync/{id} [post]
func (h *AsignaturaHandler) SyncAsignatura(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Tarea asíncrona para no bloquear el hilo principal
	go h.Service.SyncToMoodle(uint(id))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sincronización de la Asignatura iniciada correctamente en segundo plano."))
}

// BulkSyncAsignaturas maneja la sincronización masiva de asignaturas a Moodle. (POST /asignatura/bulk-sync)
// @Summary Sincronización masiva de Asignaturas
// @Description Sincroniza todas las asignaturas que no tienen ID_Moodle a Moodle
// @Tags asignatura
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /asignatura/bulk-sync [post]
func (h *AsignaturaHandler) BulkSyncAsignaturas(w http.ResponseWriter, r *http.Request) {
	h.Service.BulkSyncToMoodle()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sincronización masiva de asignaturas iniciada correctamente en segundo plano."))
}
