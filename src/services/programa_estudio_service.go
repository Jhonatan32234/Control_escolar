package services

import (
	"fmt"
	"log"

	"api_concurrencia/src/models"
	"api_concurrencia/src/moodle"
	"api_concurrencia/src/repository"
)

type ProgramaEstudioService struct {
	Repo *repository.ProgramaEstudioRepository
	// Aquí se inyectaría el cliente de Moodle API
	MoodleClient *moodle.Client
}

func NewProgramaEstudioService(repo *repository.ProgramaEstudioRepository, client *moodle.Client) *ProgramaEstudioService {
	return &ProgramaEstudioService{Repo: repo, MoodleClient: client}
}

// CreateLocal crea el registro en la BD local y lo prepara.
func (s *ProgramaEstudioService) CreateLocal(pe *models.ProgramaEstudio) error {
	return s.Repo.Create(pe)
}

func safeString(s *string) string {
    if s != nil {
        return *s
    }
    return ""
}

// SyncToMoodle realiza la lógica de sincronización (Pasos 1 a 3 del flujo).
func (s *ProgramaEstudioService) SyncToMoodle(id uint) error {
    pe, err := s.Repo.GetByID(id)
    if err != nil {
        return fmt.Errorf("PE no encontrado en BD local: %w", err)
    }

    // Si ya tiene ID_Moodle, no lo creamos de nuevo (solo actualizaríamos, pero por ahora solo crearemos)
    if pe.ID_Moodle != nil {
        log.Printf("PE ID %d ya sincronizado (Moodle ID: %d). Saltando creación.", id, *pe.ID_Moodle)
        return nil 
    }

    // 1. Construir el array de datos para la función de Moodle
    data := []moodle.CategoryRequest{
        {
            Name: pe.Nombre,
            Parent: 0, // 0 para categoría padre, como se requiere
            IDNumber: safeString(pe.ID_Externo), // SafeString maneja punteros nulos
            Description: safeString(pe.Descripcion),
        },
    }
    
    // 2. Ejecutar la llamada a la API de Moodle
    var response []moodle.CategoryResponse
    err = s.MoodleClient.Call("core_course_create_categories", data, &response)
    if err != nil {
        return fmt.Errorf("fallo al crear categoría en Moodle: %w", err)
    }

    // 3. Procesar la respuesta y actualizar el ID_Moodle local
    if len(response) == 0 {
        return fmt.Errorf("moodle no devolvió ninguna categoría creada")
    }

    moodleID := response[0].ID
    pe.ID_Moodle = &moodleID
    
    if err := s.Repo.Update(&pe); err != nil {
        return fmt.Errorf("falla al actualizar ID Moodle local para PE ID %d: %w", id, err)
    }

    log.Printf("✅ Programa Estudio '%s' (ID local: %d) creado exitosamente en Moodle con ID: %d", pe.Nombre, id, moodleID)
    return nil
}

// GetByID recupera un PE.
func (s *ProgramaEstudioService) GetByID(id uint) (models.ProgramaEstudio, error) {
    return s.Repo.GetByID(id) 
}

// GetAll recupera todos los PE, delegando al repo.
func (s *ProgramaEstudioService) GetAll() ([]models.ProgramaEstudio, error) {
    return s.Repo.GetAll() 
}

// UpdateLocal actualiza el registro en la BD local.
func (s *ProgramaEstudioService) UpdateLocal(pe *models.ProgramaEstudio) error {
	return s.Repo.Update(pe)
}

// DeleteLocal elimina el registro en la BD local.
func (s *ProgramaEstudioService) DeleteLocal(id uint) error {
	// Nota: Idealmente, aquí se verificaría que no tenga hijos antes de borrar.
	return s.Repo.Delete(id)
}

