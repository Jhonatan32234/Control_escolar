package models

import "gorm.io/gorm"

// Docente representa a un profesor y sus asignaturas a impartir.
type Docente struct {
	gorm.Model
	Nombre     string  `gorm:"type:varchar(255);not null" json:"nombre"`
	Email      string  `gorm:"type:varchar(255);not null;unique" json:"email"`
	ID_Externo *string `gorm:"type:varchar(100);unique" json:"id_externo"`
	ID_Moodle  *uint   `gorm:"unique" json:"id_moodle"`

	// Relación: Asignaturas que imparte
	Asignaturas []Asignatura `gorm:"many2many:docente_asignaturas;" json:"asignaturas"`
}
