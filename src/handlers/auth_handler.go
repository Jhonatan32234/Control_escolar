package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"time"

	"api_concurrencia/src/models"
	"api_concurrencia/src/services"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	UsuarioService *services.UsuarioService
}

func NewAuthHandler(us *services.UsuarioService) *AuthHandler {
	return &AuthHandler{UsuarioService: us}
}

type RegisterRequest struct {
	Username  string  `json:"username" example:"jperez2025"`
	Password  string  `json:"password" example:"Segura123#"`
	FirstName string  `json:"first_name" example:"Juan"`
	LastName  string  `json:"last_name" example:"Pérez García"`
	Email     string  `json:"email" example:"juan.perez@universidad.edu.mx"`
	Matricula *string `json:"matricula,omitempty" example:"20250001"`
	Rol       string  `json:"rol" example:"Alumno"`
}

type LoginRequest struct {
	Username string `json:"username" example:"jperez2025"`
	Password string `json:"password" example:"Segura123#"`
}

type AuthResponse struct {
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID    uint   `json:"user_id" example:"1"`
	Username  string `json:"username" example:"jperez2025"`
	Rol       string `json:"rol" example:"Alumno"`
	ExpiresAt int64  `json:"expires_at" example:"1701734400"`
}

// Register maneja el registro de nuevos usuarios
// @Summary Registrar un nuevo usuario
// @Description Registra un nuevo usuario (Docente o Alumno) con validación de contraseña y hash seguro
// @Tags Autenticación
// @Accept json
// @Produce json
// @Param usuario body RegisterRequest true "Datos del usuario para registro"
// @Success 201 {object} AuthResponse "Usuario registrado exitosamente con token JWT"
// @Failure 400 {string} string "Datos inválidos o contraseña débil"
// @Failure 409 {string} string "Usuario ya existe"
// @Failure 500 {string} string "Error interno del servidor"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validaciones
	if req.Username == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Rol == "" {
		http.Error(w, "Todos los campos (Username, Password, FirstName, LastName, Email, Rol) son obligatorios.", http.StatusBadRequest)
		return
	}

	if req.Rol != "Docente" && req.Rol != "Alumno" {
		http.Error(w, "Rol debe ser 'Docente' o 'Alumno'", http.StatusBadRequest)
		return
	}

	// Validaciones de contraseña
	if len(req.Password) < 8 {
		http.Error(w, "La contraseña debe tener al menos 8 caracteres.", http.StatusBadRequest)
		return
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(req.Password) {
		http.Error(w, "La contraseña debe contener al menos una mayúscula.", http.StatusBadRequest)
		return
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(req.Password) {
		http.Error(w, "La contraseña debe contener al menos un número.", http.StatusBadRequest)
		return
	}
	if !regexp.MustCompile(`[\W_]`).MatchString(req.Password) {
		http.Error(w, "La contraseña debe contener al menos un símbolo (*, #, etc.).", http.StatusBadRequest)
		return
	}

	// Hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error al procesar la contraseña", http.StatusInternalServerError)
		return
	}

	usuario := models.Usuario{
		Username:  req.Username,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Matricula: req.Matricula,
		Rol:       req.Rol,
	}

	// Verificar unicidad
	isDuplicate, err := h.UsuarioService.CheckUniqueFields(&usuario)
	if err != nil {
		http.Error(w, "Error interno al validar unicidad de datos.", http.StatusInternalServerError)
		return
	}
	if isDuplicate {
		http.Error(w, "Ya existe un usuario con el mismo Username, Email o Matrícula.", http.StatusConflict)
		return
	}

	// Crear usuario
	if err := h.UsuarioService.CreateLocal(&usuario); err != nil {
		http.Error(w, "Error al crear Usuario: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generar token JWT
	token, expiresAt, err := generateJWT(usuario.ID, usuario.Username, usuario.Rol)
	if err != nil {
		http.Error(w, "Error al generar token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token:     token,
		UserID:    usuario.ID,
		Username:  usuario.Username,
		Rol:       usuario.Rol,
		ExpiresAt: expiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login maneja la autenticación de usuarios
// @Summary Iniciar sesión
// @Description Autentica un usuario con username y password, devuelve un token JWT
// @Tags Autenticación
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Credenciales de inicio de sesión"
// @Success 200 {object} AuthResponse "Autenticación exitosa con token JWT"
// @Failure 400 {string} string "Datos inválidos"
// @Failure 401 {string} string "Credenciales incorrectas"
// @Failure 500 {string} string "Error interno del servidor"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username y Password son obligatorios", http.StatusBadRequest)
		return
	}

	// Buscar usuario por username
	usuario, err := h.UsuarioService.GetByUsername(req.Username)
	if err != nil {
		http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		return
	}

	// Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		return
	}

	// Generar token JWT
	token, expiresAt, err := generateJWT(usuario.ID, usuario.Username, usuario.Rol)
	if err != nil {
		http.Error(w, "Error al generar token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token:     token,
		UserID:    usuario.ID,
		Username:  usuario.Username,
		Rol:       usuario.Rol,
		ExpiresAt: expiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateJWT(userID uint, username, rol string) (string, int64, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"rol":      rol,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}
