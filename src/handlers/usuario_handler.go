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

// CreateUsuario maneja la creación de un nuevo usuario.
// @Summary Crear un nuevo usuario
// @Description Crea un nuevo usuario (Docente o Alumno) en la base de datos local con validaciones de contraseña, unicidad y rol
// @Tags Usuario
// @Accept json
// @Produce json
// @Param usuario body models.Usuario true "Datos del usuario a crear (Password debe tener mín. 8 caracteres, mayúscula, número y símbolo)"
// @Success 201 {object} models.Usuario "Usuario creado exitosamente"
// @Failure 400 {string} string "Error en los datos de entrada, campos obligatorios faltantes o contraseña inválida"
// @Failure 409 {string} string "Ya existe un usuario con el mismo Username, Email o Matrícula"
// @Failure 500 {string} string "Error interno del servidor al crear el usuario"
// @Router /usuario [post]
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

// SyncUsuario maneja la solicitud de sincronización individual.
// @Summary Sincronizar usuario con Moodle
// @Description Sincroniza un usuario local con Moodle de forma asíncrona
// @Tags Usuario
// @Produce plain
// @Param id path int true "ID del usuario a sincronizar"
// @Success 200 {string} string "Sincronización del Usuario iniciada correctamente en segundo plano"
// @Failure 400 {string} string "ID inválido"
// @Router /usuario/sync/{id} [post]
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

// BulkSyncUsuarios maneja la solicitud de sincronización masiva.
// @Summary Sincronización masiva de usuarios por rol
// @Description Sincroniza todos los usuarios de un rol específico (Docente o Alumno) con Moodle de forma asíncrona
// @Tags Usuario
// @Produce plain
// @Param role query string true "Rol a sincronizar: 'Docente' o 'Alumno'"
// @Success 200 {string} string "Sincronización masiva de usuarios iniciada correctamente en segundo plano"
// @Failure 400 {string} string "Rol inválido o no especificado"
// @Router /usuario/bulk-sync [post]
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

// MatricularUsuario maneja la matriculación de un usuario en una asignatura.
// @Summary Matricular usuario en asignatura
// @Description Matricula un usuario en una asignatura de forma asíncrona (crea el enrolamiento en Moodle)
// @Tags Usuario
// @Produce plain
// @Param usuarioID path int true "ID del usuario a matricular"
// @Param asignaturaID path int true "ID de la asignatura"
// @Success 200 {string} string "Matriculación iniciada en segundo plano"
// @Failure 400 {string} string "ID de Usuario o Asignatura inválido"
// @Router /usuario/{usuarioID}/matricular/{asignaturaID} [post]
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

// GetUsuarioByID obtiene un usuario por su ID.
// @Summary Obtener usuario por ID
// @Description Recupera un usuario específico mediante su ID
// @Tags Usuario
// @Produce json
// @Param id path int true "ID del usuario"
// @Success 200 {object} models.Usuario "Usuario encontrado"
// @Failure 400 {string} string "ID inválido"
// @Failure 404 {string} string "Usuario no encontrado"
// @Router /usuario/{id} [get]
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

// GetAllUsuarios obtiene todos los usuarios.
// @Summary Obtener todos los usuarios
// @Description Recupera la lista completa de usuarios (Docentes y Alumnos)
// @Tags Usuario
// @Produce json
// @Success 200 {array} models.Usuario "Lista de usuarios"
// @Failure 500 {string} string "Error al obtener usuarios"
// @Router /usuario [get]
func (h *UsuarioHandler) GetAllUsuarios(w http.ResponseWriter, r *http.Request) {
	cuatrimestres, err := h.Service.GetAll()
	if err != nil {
		http.Error(w, "Error al obtener Usuarios: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cuatrimestres)
}

// GetUnsyncedUsuarios obtiene usuarios no sincronizados por rol.
// @Summary Obtener usuarios no sincronizados por rol
// @Description Recupera la lista de usuarios que aún no han sido sincronizados con Moodle, filtrados por rol
// @Tags Usuario
// @Produce json
// @Param role query string true "Rol a filtrar: 'Docente' o 'Alumno'"
// @Success 200 {array} models.Usuario "Lista de usuarios no sincronizados"
// @Failure 400 {string} string "Rol no especificado o inválido"
// @Failure 500 {string} string "Error al obtener usuarios no sincronizados"
// @Router /usuario/unsynced [get]
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

// GetUsuariosByGroupID obtiene usuarios por ID de grupo.
// @Summary Obtener usuarios por ID de grupo
// @Description Recupera la lista de usuarios que pertenecen a un grupo específico
// @Tags Usuario
// @Produce json
// @Param grupoID path int true "ID del grupo"
// @Success 200 {array} models.Usuario "Lista de usuarios del grupo"
// @Failure 400 {string} string "ID de Grupo inválido"
// @Failure 500 {string} string "Error al obtener usuarios por grupo"
// @Router /usuario/by_group/{grupoID} [get]
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

// UpdateUsuario actualiza un usuario existente.
// @Summary Actualizar usuario
// @Description Actualiza un usuario existente en la base de datos local
// @Tags Usuario
// @Accept json
// @Produce json
// @Param id path int true "ID del usuario a actualizar"
// @Param usuario body models.Usuario true "Datos actualizados del usuario"
// @Success 200 {object} models.Usuario "Usuario actualizado exitosamente"
// @Failure 400 {string} string "ID inválido o error en los datos de entrada"
// @Failure 500 {string} string "Error al actualizar el usuario"
// @Router /usuario/{id} [put]
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

// DeleteUsuario elimina un usuario.
// @Summary Eliminar usuario
// @Description Elimina un usuario de la base de datos local
// @Tags Usuario
// @Param id path int true "ID del usuario a eliminar"
// @Success 204 "Usuario eliminado exitosamente"
// @Failure 400 {string} string "ID inválido"
// @Failure 500 {string} string "Error al eliminar el usuario"
// @Router /usuario/{id} [delete]
func (h *UsuarioHandler) DeleteUsuario(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := h.Service.DeleteLocal(uint(id)); err != nil {
		http.Error(w, "Error al eliminar Asignatura local: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
