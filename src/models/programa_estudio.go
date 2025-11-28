package models

import "gorm.io/gorm"

// ProgramaEstudio representa la categoría padre en Moodle.
type ProgramaEstudio struct {
	gorm.Model
	Nombre          string `gorm:"type:varchar(255);not null" json:"nombre"`
	Descripcion     *string `gorm:"type:text" json:"descripcion"`
	ID_Externo      *string `gorm:"type:varchar(100);unique" json:"id_externo"`
	ID_Moodle       *uint   `gorm:"unique" json:"id_moodle"`
	Cuatrimestres   []Cuatrimestre `json:"cuatrimestres"` // <- Nueva línea
}