package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api_concurrencia/src/models"
	"api_concurrencia/src/services"

	"github.com/go-chi/chi/v5"
)

type UsuarioHandler struct {
	Service *services.UsuarioService
}

func NewUsuarioHandler(s *services.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{Service: s}
}

// CreateUsuario maneja la creación local. (POST /usuario)
func (h *UsuarioHandler) CreateUsuario(w http.ResponseWriter, r *http.Request) {
	var u models.Usuario
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ejemplo de validación de Rol
	if u.Rol != "Docente" && u.Rol != "Alumno" {
		http.Error(w, "Rol debe ser 'Docente' o 'Alumno'", http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateLocal(&u); err != nil {
		http.Error(w, "Error al crear Usuario local: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

// SyncUsuario maneja la solicitud de sincronización individual. (POST /usuario/sync/{id})
func (h *UsuarioHandler) SyncUsuario(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Tarea asíncrona (aunque es individual, por consistencia)
	go h.Service.SyncToMoodle(uint(id))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sincronización del Usuario iniciada correctamente en segundo plano."))
}

// BulkSyncUsuarios maneja la solicitud de sincronización masiva. (POST /usuario/bulk-sync)
func (h *UsuarioHandler) BulkSyncUsuarios(w http.ResponseWriter, r *http.Request) {
	// Leer el parámetro de consulta para determinar qué rol sincronizar (ej: ?role=Alumno)
	role := r.URL.Query().Get("role")

	if role == "" || (role != "Docente" && role != "Alumno") {
		http.Error(w, "Debe especificar un rol válido (role=Docente o role=Alumno)", http.StatusBadRequest)
		return
	}

	// Lanzar la sincronización masiva en segundo plano
	h.Service.BulkSyncToMoodle(role)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sincronización masiva de usuarios iniciada correctamente en segundo plano para el rol: " + role))
}

// ... (Implementar GetByID, GetAll, Update, Delete) ...


func (h *UsuarioHandler) GetUsuarioByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	
	c, err := h.Service.GetByID(uint(id))
	if err != nil {
		http.Error(w, "Usuario no encontrado: "+err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

func (h *UsuarioHandler) GetAllUsuarios(w http.ResponseWriter, r *http.Request) {
	cuatrimestres, err := h.Service.GetAll()
	if err != nil {
		http.Error(w, "Error al obtener Usuarios: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cuatrimestres)
}


func (h *UsuarioHandler) UpdateUsuario(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	
	var c models.Usuario
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.ID = uint(id)

	if err := h.Service.UpdateLocal(&c); err != nil {
		http.Error(w, "Error al actualizar Asignatura local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

// DeleteCuatrimestre maneja la eliminación local. (DELETE /cuatrimestre/{id})
func (h *UsuarioHandler) DeleteUsuario(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	
	if err := h.Service.DeleteLocal(uint(id)); err != nil {
		http.Error(w, "Error al eliminar Asignatura local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}