package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api_concurrencia/src/models"
	"api_concurrencia/src/services"

	"github.com/go-chi/chi/v5"
)

type ProgramaEstudioHandler struct {
	Service *services.ProgramaEstudioService
}

func NewProgramaEstudioHandler(s *services.ProgramaEstudioService) *ProgramaEstudioHandler {
	return &ProgramaEstudioHandler{Service: s}
}

// CreateProgramaEstudio maneja la creación local.
func (h *ProgramaEstudioHandler) CreateProgramaEstudio(w http.ResponseWriter, r *http.Request) {
	var pe models.ProgramaEstudio
	if err := json.NewDecoder(r.Body).Decode(&pe); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateLocal(&pe); err != nil {
		http.Error(w, "Error al crear PE local: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pe)
}

// SyncProgramaEstudio maneja la solicitud de sincronización.
func (h *ProgramaEstudioHandler) SyncProgramaEstudio(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sincronización iniciada correctamente."))
}

// GetAllProgramaEstudio obtiene todos los PE.
func (h *ProgramaEstudioHandler) GetAllProgramaEstudio(w http.ResponseWriter, r *http.Request) {
	programas, err := h.Service.Repo.GetAll() // Asumiendo que el servicio llama al repositorio
	if err != nil {
		http.Error(w, "Error al obtener PE: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(programas)
}

// GetProgramaEstudioByID obtiene un PE por ID.
func (h *ProgramaEstudioHandler) GetProgramaEstudioByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	
	pe, err := h.Service.GetByID(uint(id))
	if err != nil {
		http.Error(w, "PE no encontrado: "+err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pe)
}

// UpdateProgramaEstudio maneja la actualización local.
func (h *ProgramaEstudioHandler) UpdateProgramaEstudio(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	
	var pe models.ProgramaEstudio
	if err := json.NewDecoder(r.Body).Decode(&pe); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pe.ID = uint(id) // Asegurar que se actualice el registro correcto

	if err := h.Service.UpdateLocal(&pe); err != nil {
		http.Error(w, "Error al actualizar PE local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pe)
}

// DeleteProgramaEstudio maneja la eliminación local.
func (h *ProgramaEstudioHandler) DeleteProgramaEstudio(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	
	if err := h.Service.DeleteLocal(uint(id)); err != nil {
		http.Error(w, "Error al eliminar PE local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}