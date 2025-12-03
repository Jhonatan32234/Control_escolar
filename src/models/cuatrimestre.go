package models

import "gorm.io/gorm"

// Cuatrimestre representa la subcategoría en Moodle.
// @Description Modelo de Cuatrimestre utilizado en la API y sincronizado como subcategoría en Moodle.
type Cuatrimestre struct {
	gorm.Model  `swaggerignore:"true"`
	Nombre      string  `gorm:"type:varchar(255);not null" json:"nombre" example:"Primer Cuatrimestre 2025" description:"Nombre del cuatrimestre (requerido, máx. 255 caracteres)"`
	Descripcion *string `gorm:"type:text" json:"descripcion,omitempty" example:"Cuatrimestre correspondiente al periodo enero-abril 2025" description:"Descripción del cuatrimestre (opcional)"`
	ID_Externo  *string `gorm:"type:varchar(100);unique" json:"id_externo,omitempty" example:"CUATR-2025-01" description:"Identificador externo único (opcional, máx. 100 caracteres)"`
	ID_Moodle   *uint   `gorm:"unique" json:"id_moodle,omitempty" example:"5678" description:"ID de la subcategoría en Moodle (asignado automáticamente tras sincronización)"`

	// Campo de la Clave Foránea
	ProgramaEstudioID uint `json:"programa_estudio_id" example:"3" description:"ID del programa de estudio al que pertenece (requerido)"` // <- Asegura que el valor esté presente
	// Relación: Perteneciente a un Programa de Estudio (FK)
	ProgramaEstudio ProgramaEstudio `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" swaggerignore:"true"`

	// Relación: Un Cuatrimestre tiene muchas Asignaturas
	Asignaturas []Asignatura `swaggerignore:"true"` // <- Nueva línea
}
