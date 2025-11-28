package models

import "gorm.io/gorm"

// Asignatura representa un Curso en Moodle.
type Asignatura struct {
	gorm.Model
	
	// Datos del Curso (Obligatorios para Moodle)
	NombreCompleto 	string 	`gorm:"type:varchar(255);not null" json:"nombre_completo"` // Moodle: fullname
	NombreCorto 	string 	`gorm:"type:varchar(100);unique;not null" json:"nombre_corto"` // Moodle: shortname
	
	// Información adicional (Opcional)
	Resumen 		*string `gorm:"type:text" json:"resumen"` // Moodle: summary
	ID_Externo 		*string `gorm:"type:varchar(100);unique" json:"id_externo"` // Moodle: idnumber
	
	// Sincronización con Moodle
	ID_Moodle 		*uint 	`gorm:"unique" json:"id_moodle"`
	
	// Relación de Pertenencia (Clave Foránea)
	CuatrimestreID 	uint 	`gorm:"not null" json:"cuatrimestre_id"` // <- ID local del Cuatrimestre
	// Relación: Perteneciente a un Cuatrimestre (FK)
	Cuatrimestre 	Cuatrimestre `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}