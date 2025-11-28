package models

// Matricula representa la relación de enrolamiento (Grupo Alumno).
// No usaremos gorm.Model aquí ya que es una tabla intermedia, pero mantenemos un ID.
type Matricula struct {
	ID             uint `gorm:"primaryKey"`

    // FKs Locales (Opcional, pero útil para GORM)
	AsignaturaID   uint  `json:"asignatura_id"`
	UsuarioID      uint  `json:"usuario_id"`
    
    // Datos de Moodle (Claves foráneas lógicas, no de BD)
	CourseMoodleID uint  `gorm:"not null;uniqueIndex:idx_unique_enrollment" json:"course_moodle_id"` // <-- Campo 1 con la definición del índice
	UserMoodleID   uint  `gorm:"not null;uniqueIndex:idx_unique_enrollment" json:"user_moodle_id"` // <-- Campo 2, usa el MISMO nombre
	RoleID         uint  `gorm:"not null" json:"role_id"` 
	Timestart      *int64 `json:"timestart"`
	Timeend        *int64 `json:"timeend"`
    
}