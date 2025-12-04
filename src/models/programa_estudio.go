package models

import "gorm.io/gorm"

// ProgramaEstudio representa la categoría padre en Moodle.
// @Description Modelo de Programa de Estudio utilizado en la API y sincronizado como categoría padre en Moodle.
type ProgramaEstudio struct {
	gorm.Model    `swaggerignore:"true"`
	Nombre        string         `gorm:"type:varchar(255);not null" json:"nombre" example:"Ingeniería en Sistemas Computacionales" description:"Nombre del programa de estudio (requerido, máx. 255 caracteres)"`
	Descripcion   *string        `gorm:"type:text" json:"descripcion,omitempty" example:"Programa de estudios enfocado en el desarrollo de software y sistemas de información" description:"Descripción del programa de estudio (opcional)"`
	ID_Externo    *string        `gorm:"type:varchar(100);unique" json:"id_externo,omitempty" example:"PROG-ISC-2025" description:"Identificador externo único (opcional, máx. 100 caracteres)"`
	ID_Moodle     *uint          `gorm:"unique" json:"id_moodle,omitempty" example:"9012" description:"ID de la categoría en Moodle (asignado automáticamente tras sincronización)"`
	Cuatrimestres []Cuatrimestre `json:"cuatrimestres,omitempty" swaggerignore:"true"` // <- Nueva línea
}
