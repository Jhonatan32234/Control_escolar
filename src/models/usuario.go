package models

import "gorm.io/gorm"

// Usuario representa tanto a Docente como a Alumno.
type Usuario struct {
	gorm.Model
	Username    string `gorm:"type:varchar(100);not null;unique" json:"username"` // OBLIGATORIO
	Password    string `gorm:"type:varchar(255);not null" json:"password"`       // OBLIGATORIO
	FirstName   string `gorm:"type:varchar(100);not null" json:"first_name"`       // OBLIGATORIO
	LastName    string `gorm:"type:varchar(100);not null" json:"last_name"`        // OBLIGATORIO
	Email       string `gorm:"type:varchar(255);not null;unique" json:"email"` // OBLIGATORIO
	Matricula   *string `gorm:"type:varchar(50);unique" json:"matricula"`          // Uso como 'idnumber'
	Rol         string `gorm:"type:varchar(50);not null" json:"rol"`         // 'Docente' o 'Alumno'
	ID_Moodle   *uint  `gorm:"unique" json:"id_moodle"`                           // ID devuelto por Moodle
}

// Nota: Puedes usar el campo 'Rol' para diferenciar la entidad Docente/Alumno en la lógica de negocio.