package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api_concurrencia/src/models"
	"api_concurrencia/src/services"

	"github.com/go-chi/chi/v5"
)

type CuatrimestreHandler struct {
	Service *services.CuatrimestreService
}

func NewCuatrimestreHandler(s *services.CuatrimestreService) *CuatrimestreHandler {
	return &CuatrimestreHandler{Service: s}
}

// CreateCuatrimestre maneja la creación local. (POST /cuatrimestre)
func (h *CuatrimestreHandler) CreateCuatrimestre(w http.ResponseWriter, r *http.Request) {
	var c models.Cuatrimestre
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateLocal(&c); err != nil {
		http.Error(w, "Error al crear Cuatrimestre local: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// GetCuatrimestreByID obtiene un Cuatrimestre por ID. (GET /cuatrimestre/{id})
func (h *CuatrimestreHandler) GetCuatrimestreByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	c, err := h.Service.GetByID(uint(id))
	if err != nil {
		http.Error(w, "Cuatrimestre no encontrado: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

// GetAllCuatrimestres obtiene todos los Cuatrimestres. (GET /cuatrimestre)
func (h *CuatrimestreHandler) GetAllCuatrimestres(w http.ResponseWriter, r *http.Request) {
	cuatrimestres, err := h.Service.GetAll()
	if err != nil {
		http.Error(w, "Error al obtener Cuatrimestres: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cuatrimestres)
}

// UpdateCuatrimestre maneja la actualización local. (PUT /cuatrimestre/{id})
func (h *CuatrimestreHandler) UpdateCuatrimestre(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var c models.Cuatrimestre
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.ID = uint(id)

	if err := h.Service.UpdateLocal(&c); err != nil {
		http.Error(w, "Error al actualizar Cuatrimestre local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

// DeleteCuatrimestre maneja la eliminación local. (DELETE /cuatrimestre/{id})
func (h *CuatrimestreHandler) DeleteCuatrimestre(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteLocal(uint(id)); err != nil {
		http.Error(w, "Error al eliminar Cuatrimestre local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SyncCuatrimestre maneja la solicitud de sincronización. (POST /cuatrimestre/sync/{id})
func (h *CuatrimestreHandler) SyncCuatrimestre(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Al igual que con PE, lanzamos la tarea asíncrona para no bloquear la petición HTTP
	go h.Service.SyncToMoodle(uint(id))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sincronización del Cuatrimestre iniciada correctamente en segundo plano."))
}
