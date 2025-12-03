package models

import "gorm.io/gorm"

// Grupo representa un Grupo de Moodle (dentro de un Curso/Asignatura).
// @Description Modelo de Grupo utilizado en la API y sincronizado con Moodle.
type Grupo struct {
	gorm.Model        `swaggerignore:"true"`
	Nombre            string `gorm:"type:varchar(255);not null" json:"nombre" example:"Grupo A - Turno Matutino" description:"Nombre del grupo (requerido, máx. 255 caracteres)"`
	CourseID          uint   `gorm:"not null" json:"course_id" example:"12" description:"ID de la asignatura/curso al que pertenece el grupo (requerido)"`                // ID de la Asignatura/Curso local
	ID_Moodle         *uint  `gorm:"unique" json:"id_moodle,omitempty" example:"888" description:"ID del grupo en Moodle (asignado automáticamente tras sincronización)"` // ID del Grupo devuelto por Moodle
	Description       string `json:"description,omitempty" example:"Grupo de clases matutinas para el curso de Programación" description:"Descripción del grupo (opcional)"`
	DescriptionFormat int    `json:"descriptionformat,omitempty" example:"1" description:"Formato de la descripción (1=HTML, 0=texto plano)"`
	// Relación Many-to-Many (Inversa)
	Usuarios []Usuario `gorm:"many2many:usuario_grupos;" json:"usuarios,omitempty" swaggerignore:"true"`
}
