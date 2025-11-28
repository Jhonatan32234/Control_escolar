package services

import (
	"api_concurrencia/src/models"
	"api_concurrencia/src/moodle"
	"api_concurrencia/src/repository"
	"fmt"
	"log"
)

type CuatrimestreService struct {
	Repo *repository.CuatrimestreRepository
	MoodleClient *moodle.Client
}

func NewCuatrimestreService(repo *repository.CuatrimestreRepository, moodleClient *moodle.Client) *CuatrimestreService {
	return &CuatrimestreService{Repo: repo, MoodleClient: moodleClient}
}

// CreateLocal crea el registro en la BD local.
func (s *CuatrimestreService) CreateLocal(c *models.Cuatrimestre) error {
	// Lógica de validación, por ejemplo: verificar si ProgramaEstudioID existe.
	return s.Repo.Create(c)
}

// GetAll recupera todos los Cuatrimestres.
func (s *CuatrimestreService) GetAll() ([]models.Cuatrimestre, error) {
	return s.Repo.GetAll()
}

// GetByID recupera un Cuatrimestre.
func (s *CuatrimestreService) GetByID(id uint) (models.Cuatrimestre, error) {
	return s.Repo.GetByID(id)
}

// UpdateLocal actualiza el registro en la BD local.
func (s *CuatrimestreService) UpdateLocal(c *models.Cuatrimestre) error {
	return s.Repo.Update(c)
}

// DeleteLocal elimina el registro en la BD local.
func (s *CuatrimestreService) DeleteLocal(id uint) error {
	return s.Repo.Delete(id)
}

// SyncToMoodle simula la lógica de sincronización para Cuatrimestre.
func (s *CuatrimestreService) SyncToMoodle(id uint) error {
	cuatrimestre, err := s.Repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("cuatrimestre no encontrado en BD local: %w", err)
	}

	// 0. Validación Clave: El ProgramaEstudio padre debe estar sincronizado
	if cuatrimestre.ProgramaEstudio.ID_Moodle == nil {
		return fmt.Errorf("error: El ProgramaEstudio padre (ID: %d) no ha sido sincronizado con Moodle (ID_Moodle es nulo)", cuatrimestre.ProgramaEstudioID)
	}

	// Si ya tiene ID_Moodle, saltamos la creación
	if cuatrimestre.ID_Moodle != nil {
		log.Printf("Cuatrimestre ID %d ya sincronizado (Moodle ID: %d). Saltando creación.", id, *cuatrimestre.ID_Moodle)
		return nil
	}

	// 1. Construir el array de datos para la función de Moodle
    // El Parent es el ID_Moodle del ProgramaEstudio (categoría padre)
	parentID := *cuatrimestre.ProgramaEstudio.ID_Moodle 

	data := []moodle.CategoryRequest{
		{
			Name:        cuatrimestre.Nombre,
			Parent:      int(parentID), // 👈 USAMOS EL ID MOODLE DEL PADRE
			IDNumber:    safeString(cuatrimestre.ID_Externo),
			Description: safeString(cuatrimestre.Descripcion),
		},
	}

	// 2. Ejecutar la llamada a la API de Moodle
	var response []moodle.CategoryResponse
	err = s.MoodleClient.Call("core_course_create_categories", data, &response)
	if err != nil {
		return fmt.Errorf("fallo al crear subcategoría en Moodle: %w", err)
	}

	// 3. Procesar la respuesta y actualizar el ID_Moodle local
	if len(response) == 0 {
		return fmt.Errorf("moodle no devolvió ninguna subcategoría creada")
	}

	moodleID := response[0].ID
	cuatrimestre.ID_Moodle = &moodleID

	if err := s.Repo.Update(&cuatrimestre); err != nil {
		return fmt.Errorf("falla al actualizar ID Moodle local para Cuatrimestre ID %d: %w", id, err)
	}

	log.Printf("✅ Cuatrimestre '%s' (ID local: %d) creado exitosamente en Moodle como subcategoría de ID: %d", cuatrimestre.Nombre, id, moodleID)
	return nil
}