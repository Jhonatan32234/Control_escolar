package models

import "gorm.io/gorm"

// Grupo representa un grupo de una asignatura con un docente y alumnos.
type Grupo struct {
	gorm.Model
	Nombre string `gorm:"type:varchar(255);not null" json:"nombre"`

	// Relación principal con Asignatura
	AsignaturaID uint       `gorm:"not null" json:"asignatura_id"`
	Asignatura   Asignatura `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// Docente asignado
	DocenteID uint    `gorm:"not null" json:"docente_id"`
	Docente   Docente `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// Alumnos (Usuarios con rol Alumno)
	Alumnos []Usuario `gorm:"many2many:grupo_alumnos;" json:"alumnos"`
}
