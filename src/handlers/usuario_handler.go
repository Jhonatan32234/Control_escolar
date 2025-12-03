package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
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

func (h *UsuarioHandler) CreateUsuario(w http.ResponseWriter, r *http.Request) {
    var u models.Usuario
    if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
        http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    // 0. Bloquear el envío de ID_Moodle por el cliente
    if u.ID_Moodle != nil {
        http.Error(w, "No se permite enviar el campo ID_Moodle en la creación de un usuario. Este campo es asignado por el sistema.", http.StatusBadRequest)
        return
    }

    // 0.1. Validar campos obligatorios
    if u.Username == "" || u.Password == "" || u.FirstName == "" || u.LastName == "" || u.Email == "" || u.Rol == "" {
        http.Error(w, "Todos los campos (Username, Password, FirstName, LastName, Email, Rol) son obligatorios.", http.StatusBadRequest)
        return
    }

    isDuplicate, err := h.Service.CheckUniqueFields(&u)
    if err != nil {
        log.Println("Error al verificar unicidad:", err)
        http.Error(w, "Error interno al validar unicidad de datos.", http.StatusInternalServerError)
        return
    }
    if isDuplicate {
        http.Error(w, "Ya existe un usuario con el mismo Username, Email o Matrícula.", http.StatusConflict) // Status 409 Conflict
        return
    }

    // 1. Validación de Rol (Existente)
    if u.Rol != "Docente" && u.Rol != "Alumno" {
        http.Error(w, "Rol debe ser 'Docente' o 'Alumno'", http.StatusBadRequest)
        return
    }

    // Al menos una mayúscula (A-Z)
    if !regexp.MustCompile(`[A-Z]`).MatchString(u.Password) {
        http.Error(w, "La contraseña debe contener al menos una mayúscula.", http.StatusBadRequest)
        return
    }

    // Al menos un dígito (0-9)
    if !regexp.MustCompile(`[0-9]`).MatchString(u.Password) {
        http.Error(w, "La contraseña debe contener al menos un número.", http.StatusBadRequest)
        return
    }

    // Al menos un caracter especial no alfanumérico (usamos \W que incluye _)
    if !regexp.MustCompile(`[\W_]`).MatchString(u.Password) {
        http.Error(w, "La contraseña debe contener al menos un símbolo (*, #, etc.).", http.StatusBadRequest)
        return
    }

    // Validar longitud mínima
    if len(u.Password) < 8 {
        http.Error(w, "La contraseña debe tener al menos 8 caracteres.", http.StatusBadRequest)
        return
    }
    
    // 3. Crear el registro en la base de datos local
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

func (h *UsuarioHandler) MatricularUsuario(w http.ResponseWriter, r *http.Request) {
    usuarioIDStr := chi.URLParam(r, "usuarioID")
    asignaturaIDStr := chi.URLParam(r, "asignaturaID")

    usuarioID, err := strconv.ParseUint(usuarioIDStr, 10, 32)
    if err != nil {
        http.Error(w, "ID de Usuario inválido", http.StatusBadRequest)
        return
    }

    asignaturaID, err := strconv.ParseUint(asignaturaIDStr, 10, 32)
    if err != nil {
        http.Error(w, "ID de Asignatura inválido", http.StatusBadRequest)
        return
    }

    // Ejecutamos la función de servicio en segundo plano (asíncrona)
    go func() {
        if err := h.Service.MatricularUsuario(uint(usuarioID), uint(asignaturaID)); err != nil {
            // Es importante registrar errores en la goroutine, ya que no podemos devolverlos al cliente HTTP
            log.Printf("ERROR de Matrícula (U:%d, A:%d): %v", usuarioID, asignaturaID, err)
        }
    }()

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf("Matriculación del Usuario %d en la Asignatura %d iniciada en segundo plano.", usuarioID, asignaturaID)))
}

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


// GetUnsyncedUsuarios maneja la solicitud GET para obtener usuarios no sincronizados por rol. (GET /usuario/unsynced?role=Alumno)
func (h *UsuarioHandler) GetUnsyncedUsuarios(w http.ResponseWriter, r *http.Request) {
    // Leer el parámetro de consulta para determinar qué rol filtrar (ej: ?role=Alumno)
    role := r.URL.Query().Get("role")

    if role == "" {
        http.Error(w, "Debe especificar el parámetro 'role' (Docente o Alumno).", http.StatusBadRequest)
        return
    }
    if role != "Docente" && role != "Alumno" {
        http.Error(w, "El rol especificado es inválido. Use 'Docente' o 'Alumno'.", http.StatusBadRequest)
        return
    }

    usuarios, err := h.Service.GetUnsyncedByRole(role)
    if err != nil {
        log.Printf("Error al obtener usuarios no sincronizados: %v", err)
        http.Error(w, "Error al obtener usuarios no sincronizados: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(usuarios)
}

// GetUsuariosByGroupID maneja la solicitud GET para obtener usuarios por ID de Grupo. (GET /usuario/by_group/{grupoID})
func (h *UsuarioHandler) GetUsuariosByGroupID(w http.ResponseWriter, r *http.Request) {
    grupoIDStr := chi.URLParam(r, "grupoID")
    grupoID, err := strconv.ParseUint(grupoIDStr, 10, 32)
    if err != nil {
        http.Error(w, "ID de Grupo inválido", http.StatusBadRequest)
        return
    }

    usuarios, err := h.Service.GetByGroupID(uint(grupoID))
    if err != nil {
        log.Printf("Error al obtener usuarios por grupo: %v", err)
        http.Error(w, "Error al obtener usuarios por grupo: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(usuarios)
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