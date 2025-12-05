package moodle

// Estructura para crear un Usuario en Moodle (core_user_create_users)
type UserRequest struct {
	Username  string `json:"username"` // Debe ser único (ej: matrícula, email)
	Password  string `json:"password"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`

	// Opcionales
	IDNumber string `json:"idnumber,omitempty"` // ID externo (ej: ID de empleado o alumno)
	// Auth        string `json:"auth,omitempty"`     // Método de autenticación (ej: 'manual')
}

// Estructura para actualizar un Usuario en Moodle (core_user_update_users)
type UserUpdateRequest struct {
	ID        uint   `json:"id"`                  // ID de Moodle (requerido)
	Username  string `json:"username,omitempty"`  // Nuevo username
	Password  string `json:"password,omitempty"`  // Nueva contraseña
	Firstname string `json:"firstname,omitempty"` // Nuevo nombre
	Lastname  string `json:"lastname,omitempty"`  // Nuevo apellido
	Email     string `json:"email,omitempty"`     // Nuevo email
	IDNumber  string `json:"idnumber,omitempty"`  // Nuevo ID externo
}

// Estructura de respuesta después de crear un usuario
type UserResponse struct {
	ID       uint   `json:"id"` // ID de Moodle asignado
	Username string `json:"username"`
}

// Estructura para la matriculación manual (enrol_manual_enrol_users)
type EnrolmentRequest struct {
	RoleID   int  `json:"roleid"`   // 5: Estudiante, 3: Profesor
	UserID   uint `json:"userid"`   // ID de Moodle del Usuario
	CourseID uint `json:"courseid"` // ID de Moodle del Curso (Asignatura)
	// Timestart y Timeend (Opcionales para definir un periodo de matrícula)
}

// CategoryRequest representa la estructura esperada por core_course_create_categories.
// Los datos se envían como un array de CategoryRequest.
type CategoryRequest struct {
	Name        string `json:"name"`
	Parent      int    `json:"parent"` // 0 para categoría padre
	IDNumber    string `json:"idnumber,omitempty"`
	Description string `json:"description,omitempty"`
}

// CategoryUpdateRequest para actualizar categorías en Moodle
type CategoryUpdateRequest struct {
	ID          uint   `json:"id"`                    // ID de Moodle (requerido)
	Name        string `json:"name,omitempty"`        // Nuevo nombre
	IDNumber    string `json:"idnumber,omitempty"`    // Nuevo ID externo
	Description string `json:"description,omitempty"` // Nueva descripción
}

// CategoryResponse representa la estructura que Moodle devuelve al crear una categoría.
type CategoryResponse struct {
	ID       uint   `json:"id"` // El ID de Moodle que necesitamos guardar
	Name     string `json:"name"`
	Parent   int    `json:"parent"`
	IDNumber string `json:"idnumber"`
}

// Estructura para crear un Curso (Asignatura) en Moodle
type CourseRequest struct {
	// Requeridos
	Fullname   string `json:"fullname"`
	Shortname  string `json:"shortname"`
	Categoryid int    `json:"categoryid"` // Este será el ID_Moodle del Cuatrimestre padre

	// Opcionales/Recomendados
	IDNumber string `json:"idnumber,omitempty"` // ID Externo (para evitar duplicados)
	Summary  string `json:"summary,omitempty"`  // Resumen/Descripción
	Format   string `json:"format,omitempty"`   // 'topics', 'weeks', etc.
}

// CourseUpdateRequest para actualizar cursos en Moodle
type CourseUpdateRequest struct {
	ID        uint   `json:"id"`                  // ID de Moodle (requerido)
	Fullname  string `json:"fullname,omitempty"`  // Nuevo nombre completo
	Shortname string `json:"shortname,omitempty"` // Nuevo nombre corto
	IDNumber  string `json:"idnumber,omitempty"`  // Nuevo ID externo
	Summary   string `json:"summary,omitempty"`   // Nuevo resumen
}

// Estructura de respuesta después de crear un curso
type CourseResponse struct {
	ID        uint   `json:"id"` // ID de Moodle asignado
	Shortname string `json:"shortname"`
}

// Definiciones para core_group_create_groups
type GroupRequest struct {
	CourseID          int    `json:"courseid"`
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	DescriptionFormat int    `json:"descriptionformat,omitempty"` // 👈 NUEVO: 1=HTML
	EnrolmentKey      string `json:"enrolmentkey,omitempty"`      // 👈 NUEVO: Clave
	IDNumber          string `json:"idnumber,omitempty"`
	Visibility        int    `json:"visibility,omitempty"`    // 👈 NUEVO: 0=Visible
	Participation     int    `json:"participation,omitempty"` // 👈 NUEVO: 1=Habilitado
}

type GroupResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	CourseID int    `json:"courseid"`
	IDNumber string `json:"idnumber"`
}

// Definiciones para core_group_add_group_members
type GroupMemberRequest struct {
	GroupID int `json:"groupid"`
	UserID  int `json:"userid"`
}
