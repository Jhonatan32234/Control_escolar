package moodle

// CategoryRequest representa la estructura esperada por core_course_create_categories.
// Los datos se envían como un array de CategoryRequest.
type CategoryRequest struct {
	Name        string `json:"name"`
	Parent      int    `json:"parent"` // 0 para categoría padre
	IDNumber    string `json:"idnumber,omitempty"`
	Description string `json:"description,omitempty"`
}

// CategoryResponse representa la estructura que Moodle devuelve al crear una categoría.
type CategoryResponse struct {
	ID       uint   `json:"id"`        // El ID de Moodle que necesitamos guardar
	Name     string `json:"name"`
	Parent   int    `json:"parent"`
	IDNumber string `json:"idnumber"`
}

// Estructura para crear un Curso (Asignatura) en Moodle
type CourseRequest struct {
	// Requeridos
	Fullname 	string 	`json:"fullname"`
	Shortname 	string 	`json:"shortname"`
	Categoryid 	int 	`json:"categoryid"` // Este será el ID_Moodle del Cuatrimestre padre
	
	// Opcionales/Recomendados
	IDNumber 	string 	`json:"idnumber,omitempty"` // ID Externo (para evitar duplicados)
	Summary 	string 	`json:"summary,omitempty"` 	// Resumen/Descripción
	Format 		string 	`json:"format,omitempty"` // 'topics', 'weeks', etc.
}

// Estructura de respuesta después de crear un curso
type CourseResponse struct {
	ID 			uint 	`json:"id"` // ID de Moodle asignado
	Shortname 	string 	`json:"shortname"`
}
