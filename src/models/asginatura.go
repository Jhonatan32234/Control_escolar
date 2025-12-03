package models

import "gorm.io/gorm"

// Asignatura representa un Curso en Moodle.
// @Description Modelo de Asignatura (Curso) utilizado en la API y sincronizado con Moodle.
type Asignatura struct {
	gorm.Model `swaggerignore:"true"`

	// Datos del Curso (Obligatorios para Moodle)
	NombreCompleto string `gorm:"type:varchar(255);not null" json:"nombre_completo" example:"Programación Orientada a Objetos I" description:"Nombre completo del curso (requerido, máx. 255 caracteres)"` // Moodle: fullname
	NombreCorto    string `gorm:"type:varchar(100);unique;not null" json:"nombre_corto" example:"POO1-2025-A" description:"Nombre corto único del curso (requerido, máx. 100 caracteres)"`                 // Moodle: shortname

	// Información adicional (Opcional)
	Resumen    *string `gorm:"type:text" json:"resumen,omitempty" example:"Curso introductorio de programación orientada a objetos que cubre conceptos fundamentales como clases, objetos, herencia y polimorfismo." description:"Descripción del curso (opcional)"` // Moodle: summary
	ID_Externo *string `gorm:"type:varchar(100);unique" json:"id_externo,omitempty" example:"ASIG-POO1-2025" description:"Identificador externo único (opcional, máx. 100 caracteres)"`                                                                              // Moodle: idnumber

	// Sincronización con Moodle
	ID_Moodle *uint `gorm:"unique" json:"id_moodle,omitempty" example:"1234" description:"ID del curso en Moodle (asignado automáticamente tras sincronización)"`

	// Relación de Pertenencia (Clave Foránea)
	CuatrimestreID uint `gorm:"not null" json:"cuatrimestre_id" example:"5" description:"ID del cuatrimestre al que pertenece (requerido)"` // <- ID local del Cuatrimestre
	// Relación: Perteneciente a un Cuatrimestre (FK)
	Cuatrimestre Cuatrimestre `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" swaggerignore:"true"`
}
