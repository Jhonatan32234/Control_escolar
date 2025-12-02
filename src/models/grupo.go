package models

import "gorm.io/gorm"

// Grupo representa un Grupo de Moodle (dentro de un Curso/Asignatura).
type Grupo struct {
    gorm.Model
    Nombre      string `gorm:"type:varchar(255);not null" json:"nombre"`
    CourseID    uint   `gorm:"not null" json:"course_id"` // ID de la Asignatura/Curso local
    ID_Moodle   *uint  `gorm:"unique" json:"id_moodle"`  // ID del Grupo devuelto por Moodle
    Description       string `json:"description"` // 👈 CLAVE: QUITAR omitempty
    DescriptionFormat int    `json:"descriptionformat,omitempty"`
    // Relación Many-to-Many (Inversa)
    Usuarios    []Usuario `gorm:"many2many:usuario_grupos;" json:"usuarios,omitempty"`
}