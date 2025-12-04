package models

// Matricula representa la relación de enrolamiento (Usuario - Asignatura - Rol).
// @Description Modelo de Matricula que representa el enrolamiento de un usuario en una asignatura con un rol específico en Moodle.
type Matricula struct {
	ID uint `gorm:"primaryKey" json:"id" example:"1" description:"ID único de la matrícula"`

	// 🛑 FKs Locales (RELACIONES)
	AsignaturaID uint `gorm:"not null" json:"asignatura_id" example:"10" description:"ID de la asignatura (requerido)"`
	// Relación 1: Perteneciente a una Asignatura
	Asignatura Asignatura `gorm:"foreignKey:AsignaturaID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" swaggerignore:"true"`

	UsuarioID uint `gorm:"not null" json:"usuario_id" example:"25" description:"ID del usuario (requerido)"`
	// Relación 2: Perteneciente a un Usuario (Docente/Alumno)
	Usuario Usuario `gorm:"foreignKey:UsuarioID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" swaggerignore:"true"`

	// Datos de Moodle (Claves de sincronización)
	// Creamos un índice único compuesto para evitar dobles enrolamientos.
	CourseMoodleID uint `gorm:"not null;uniqueIndex:idx_unique_enrollment" json:"course_moodle_id" example:"1234" description:"ID del curso en Moodle (requerido, único por combinación usuario-curso)"`
	UserMoodleID   uint `gorm:"not null;uniqueIndex:idx_unique_enrollment" json:"user_moodle_id" example:"5678" description:"ID del usuario en Moodle (requerido, único por combinación usuario-curso)"`
	RoleID         uint `gorm:"not null" json:"role_id" example:"5" description:"ID del rol en Moodle (requerido: 5=Estudiante, 3=Docente)"` // 5=Estudiante, 3=Docente

	// Tiempos de enrolamiento
	Timestart *int64 `json:"timestart,omitempty" example:"1704067200" description:"Timestamp de inicio del enrolamiento (opcional, UNIX timestamp)"`
	Timeend   *int64 `json:"timeend,omitempty" example:"1719792000" description:"Timestamp de finalización del enrolamiento (opcional, UNIX timestamp)"`
}
