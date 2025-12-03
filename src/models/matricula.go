package models

// Matricula representa la relación de enrolamiento (Usuario - Asignatura - Rol).
type Matricula struct {
	ID uint `gorm:"primaryKey"`

	// 🛑 FKs Locales (RELACIONES)
	AsignaturaID uint `gorm:"not null" json:"asignatura_id"`
	// Relación 1: Perteneciente a una Asignatura
	Asignatura Asignatura `gorm:"foreignKey:AsignaturaID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	UsuarioID uint `gorm:"not null" json:"usuario_id"`
	// Relación 2: Perteneciente a un Usuario (Docente/Alumno)
	Usuario Usuario `gorm:"foreignKey:UsuarioID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// Datos de Moodle (Claves de sincronización)
	// Creamos un índice único compuesto para evitar dobles enrolamientos.
	CourseMoodleID uint `gorm:"not null;uniqueIndex:idx_unique_enrollment" json:"course_moodle_id"`
	UserMoodleID   uint `gorm:"not null;uniqueIndex:idx_unique_enrollment" json:"user_moodle_id"`
	RoleID         uint `gorm:"not null" json:"role_id"` // 5=Estudiante, 3=Docente

	// Tiempos de enrolamiento
	Timestart *int64 `json:"timestart"`
	Timeend   *int64 `json:"timeend"`
}