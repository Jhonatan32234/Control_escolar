package services

import (
	"api_concurrencia/src/models"
	"api_concurrencia/src/moodle"
	"api_concurrencia/src/repository"
	"errors"
	"fmt"
	"log"
)

type AsignaturaService struct {
	Repo *repository.AsignaturaRepository
	MoodleClient *moodle.Client
}

func NewAsignaturaService(repo *repository.AsignaturaRepository, moodleClient *moodle.Client) *AsignaturaService {
	return &AsignaturaService{Repo: repo, MoodleClient: moodleClient}
}

// CreateLocal crea el registro en la BD local.
func (s *AsignaturaService) CreateLocal(a *models.Asignatura) error {
	// Lógica de validación: Verificar que el CuatrimestreID exista localmente.
	if a.CuatrimestreID == 0 {
		return errors.New("CuatrimestreID es obligatorio")
	}
	// Se requeriría la importación de "errors" si el paquete no lo tiene por defecto.
	return s.Repo.Create(a)
}

// GetAll recupera todas las Asignaturas.
func (s *AsignaturaService) GetAll() ([]models.Asignatura, error) {
	return s.Repo.GetAll()
}

// GetByID recupera una Asignatura.
func (s *AsignaturaService) GetByID(id uint) (models.Asignatura, error) {
	return s.Repo.GetByID(id)
}

// UpdateLocal actualiza el registro en la BD local.
func (s *AsignaturaService) UpdateLocal(a *models.Asignatura) error {
	return s.Repo.Update(a)
}

// DeleteLocal elimina el registro en la BD local.
func (s *AsignaturaService) DeleteLocal(id uint) error {
	return s.Repo.Delete(id)
}

// SyncToMoodle simula la lógica de sincronización para Asignatura (Curso).
func (s *AsignaturaService) SyncToMoodle(id uint) error {
    asignatura, err := s.Repo.GetByID(id)
    if err != nil {
        return fmt.Errorf("asignatura no encontrada en BD local: %w", err)
    }

    // 0. Validación Clave: El Cuatrimestre padre debe estar sincronizado
    // NOTA: Asegúrate que tu Repo.GetByID precarga la relación Cuatrimestre, y este precarga el ID_Moodle.
    if asignatura.Cuatrimestre.ID_Moodle == nil { // 👈 VERIFICAMOS EL CUATRIMESTRE
        return fmt.Errorf("error: El Cuatrimestre padre (ID: %d) no ha sido sincronizado con Moodle (ID_Moodle es nulo)", asignatura.CuatrimestreID)
    }

    // Si ya tiene ID_Moodle, saltamos la creación (Asignatura/Curso ya creado)
    if asignatura.ID_Moodle != nil { // 👈 VERIFICAMOS LA ASIGNATURA
        log.Printf("Asignatura ID %d ya sincronizada (Moodle ID: %d). Saltando creación.", id, *asignatura.ID_Moodle)
        return nil
    }

    // 1. Construir el array de datos para la función de Moodle
    moodleParentID := *asignatura.Cuatrimestre.ID_Moodle 
    
    // **USAMOS EL STRUCT DE CURSO (CourseRequest)**
    data := []moodle.CourseRequest{
        {
            Fullname: asignatura.NombreCompleto, // 👈 DATOS DE LA ASIGNATURA
            Shortname: asignatura.NombreCorto,   // 👈 REQUERIDO: Nombre corto único
            Categoryid: int(moodleParentID),     // 👈 ID MOODLE del Cuatrimestre padre
            IDNumber: safeString(asignatura.ID_Externo),
            Summary: safeString(asignatura.Resumen),
        },
    }

    // 2. Ejecutar la llamada a la API de Moodle
    var response []moodle.CourseResponse // 👈 USAMOS EL STRUCT DE RESPUESTA DE CURSO
    err = s.MoodleClient.Call("core_course_create_courses", data, &response) // 👈 USAMOS LA FUNCIÓN DE CURSOS
    if err != nil {
        return fmt.Errorf("fallo al crear Curso/Asignatura en Moodle: %w", err)
    }

    // 3. Procesar la respuesta y actualizar el ID_Moodle local
    if len(response) == 0 {
        return fmt.Errorf("moodle no devolvió ningún Curso/Asignatura creado")
    }

    moodleID := response[0].ID
    asignatura.ID_Moodle = &moodleID

    if err := s.Repo.Update(&asignatura); err != nil {
        return fmt.Errorf("falla al actualizar ID Moodle local para Asignatura ID %d: %w", id, err)
    }

    log.Printf("✅ Asignatura '%s' (ID local: %d) creada exitosamente en Moodle como Curso de ID: %d", asignatura.NombreCompleto, id, moodleID)
    return nil
}