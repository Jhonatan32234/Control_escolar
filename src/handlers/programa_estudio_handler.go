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
// @Summary Crear un nuevo programa de estudio
// @Description Crea un nuevo programa de estudio en la base de datos local
// @Tags ProgramaEstudio
// @Accept json
// @Produce json
// @Param programa_estudio body models.ProgramaEstudio true "Datos del programa de estudio a crear"
// @Success 201 {object} models.ProgramaEstudio "Programa de estudio creado exitosamente"
// @Failure 400 {string} string "Error en los datos de entrada o campos obligatorios faltantes"
// @Failure 500 {string} string "Error interno del servidor al crear el programa de estudio"
// @Router /programa_estudio [post]
func (h *ProgramaEstudioHandler) CreateProgramaEstudio(w http.ResponseWriter, r *http.Request) {
	var pe models.ProgramaEstudio
	if err := json.NewDecoder(r.Body).Decode(&pe); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 🛑 VALIDACIÓN DE ENTRADA
	if pe.Nombre == "" || pe.Descripcion == nil || *pe.Descripcion == "" {
		http.Error(w, "Faltan campos obligatorios: nombre o descripción.", http.StatusBadRequest)
		return
	}
	if pe.ID_Externo == nil || *pe.ID_Externo == "" {
		http.Error(w, "El campo 'id_externo' es obligatorio.", http.StatusBadRequest)
		return
	}
	// Prevenir que el cliente establezca ID_Moodle o ID local
	pe.ID_Moodle = nil
	pe.ID = 0

	if err := h.Service.CreateLocal(&pe); err != nil {
		http.Error(w, "Error al crear PE local: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pe)
}

// SyncProgramaEstudio maneja la solicitud de sincronización.
// @Summary Sincronizar programa de estudio con Moodle
// @Description Sincroniza un programa de estudio local con Moodle como categoría padre
// @Tags ProgramaEstudio
// @Produce plain
// @Param id path int true "ID del programa de estudio a sincronizar"
// @Success 200 {string} string "Sincronización iniciada correctamente"
// @Failure 400 {string} string "ID inválido"
// @Failure 500 {string} string "Error durante la sincronización"
// @Router /programa_estudio/sync/{id} [post]
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
// @Summary Obtener todos los programas de estudio
// @Description Recupera la lista completa de programas de estudio
// @Tags ProgramaEstudio
// @Produce json
// @Success 200 {array} models.ProgramaEstudio "Lista de programas de estudio"
// @Failure 500 {string} string "Error al obtener programas de estudio"
// @Router /programa_estudio [get]
func (h *ProgramaEstudioHandler) GetAllProgramaEstudio(w http.ResponseWriter, r *http.Request) {
	programas, err := h.Service.GetAll()
	if err != nil {
		http.Error(w, "Error al obtener PE: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(programas)
}

// GetProgramaEstudioByID obtiene un PE por ID.
// @Summary Obtener programa de estudio por ID
// @Description Recupera un programa de estudio específico mediante su ID
// @Tags ProgramaEstudio
// @Produce json
// @Param id path int true "ID del programa de estudio"
// @Success 200 {object} models.ProgramaEstudio "Programa de estudio encontrado"
// @Failure 400 {string} string "ID inválido"
// @Failure 404 {string} string "Programa de estudio no encontrado"
// @Router /programa_estudio/{id} [get]
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
// @Summary Actualizar programa de estudio
// @Description Actualiza un programa de estudio existente en la base de datos local
// @Tags ProgramaEstudio
// @Accept json
// @Produce json
// @Param id path int true "ID del programa de estudio a actualizar"
// @Param programa_estudio body models.ProgramaEstudio true "Datos actualizados del programa de estudio"
// @Success 200 {object} models.ProgramaEstudio "Programa de estudio actualizado exitosamente"
// @Failure 400 {string} string "ID inválido o error en los datos de entrada"
// @Failure 500 {string} string "Error al actualizar el programa de estudio"
// @Router /programa_estudio/{id} [put]
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
// @Summary Eliminar programa de estudio
// @Description Elimina un programa de estudio de la base de datos local
// @Tags ProgramaEstudio
// @Param id path int true "ID del programa de estudio a eliminar"
// @Success 204 "Programa de estudio eliminado exitosamente"
// @Failure 400 {string} string "ID inválido"
// @Failure 500 {string} string "Error al eliminar el programa de estudio"
// @Router /programa_estudio/{id} [delete]
func (h *ProgramaEstudioHandler) DeleteProgramaEstudio(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := h.Service.DeleteLocal(uint(id)); err != nil {
		http.Error(w, "Error al eliminar PE local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
