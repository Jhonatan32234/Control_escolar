package models

import "gorm.io/gorm"

// Cuatrimestre representa la subcategoría en Moodle.
type Cuatrimestre struct {
	gorm.Model
	Nombre          string `gorm:"type:varchar(255);not null" json:"nombre"`
	Descripcion     *string `gorm:"type:text" json:"descripcion"`
	ID_Externo      *string `gorm:"type:varchar(100);unique" json:"id_externo"`
	ID_Moodle       *uint   `gorm:"unique" json:"id_moodle"`
    
    // Campo de la Clave Foránea
	ProgramaEstudioID uint `json:"programa_estudio_id"` // <- Asegura que el valor esté presente
    // Relación: Perteneciente a un Programa de Estudio (FK)
    ProgramaEstudio ProgramaEstudio `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"` 
    
    // Relación: Un Cuatrimestre tiene muchas Asignaturas
    Asignaturas     []Asignatura // <- Nueva línea
}