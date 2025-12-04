package models

import "gorm.io/gorm"

// Usuario representa tanto a Docente como a Alumno.
// @Description Modelo de Usuario utilizado en la API y sincronizado como usuario en Moodle. Puede ser Docente o Alumno.
type Usuario struct {
	gorm.Model `swaggerignore:"true"`
	Username   string  `gorm:"type:varchar(100);not null;unique" json:"username" example:"jperez2025" description:"Nombre de usuario único (requerido, máx. 100 caracteres)"`                                 // OBLIGATORIO
	Password   string  `gorm:"type:varchar(255);not null" json:"password" example:"Segura123#" description:"Contraseña (requerido, mín. 8 caracteres, debe incluir mayúscula, número y símbolo)"`             // OBLIGATORIO
	FirstName  string  `gorm:"type:varchar(100);not null" json:"first_name" example:"Juan" description:"Nombre(s) del usuario (requerido, máx. 100 caracteres)"`                                              // OBLIGATORIO
	LastName   string  `gorm:"type:varchar(100);not null" json:"last_name" example:"Pérez García" description:"Apellido(s) del usuario (requerido, máx. 100 caracteres)"`                                     // OBLIGATORIO
	Email      string  `gorm:"type:varchar(255);not null;unique" json:"email" example:"juan.perez@universidad.edu.mx" description:"Correo electrónico único (requerido, máx. 255 caracteres)"`                // OBLIGATORIO
	Matricula  *string `gorm:"type:varchar(50);unique" json:"matricula,omitempty" example:"20250001" description:"Matrícula única del usuario (opcional, máx. 50 caracteres, usado como idnumber en Moodle)"` // Uso como 'idnumber'
	Rol        string  `gorm:"type:varchar(50);not null" json:"rol" example:"Alumno" description:"Rol del usuario (requerido: 'Docente' o 'Alumno')"`                                                         // 'Docente' o 'Alumno'
	ID_Moodle  *uint   `gorm:"unique" json:"id_moodle,omitempty" example:"3456" description:"ID del usuario en Moodle (asignado automáticamente tras sincronización)"`                                        // ID devuelto por Moodle

	Matriculas []Matricula `gorm:"foreignKey:UsuarioID" json:"matriculas,omitempty" swaggerignore:"true"`
	// Relación Many-to-Many con Grupos
	Grupos []Grupo `gorm:"many2many:usuario_grupos;" json:"grupos,omitempty" swaggerignore:"true"` // 👈 NUEVO CAMPO DE RELACIÓN
}

// Nota: Puedes usar el campo 'Rol' para diferenciar la entidad Docente/Alumno en la lógica de negocio.
