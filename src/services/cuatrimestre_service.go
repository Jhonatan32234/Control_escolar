package services

import (
	"api_concurrencia/src/models"
	"api_concurrencia/src/moodle"
	"api_concurrencia/src/repository"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

type CuatrimestreService struct {
	Repo         *repository.CuatrimestreRepository
	MoodleClient *moodle.Client
}

func NewCuatrimestreService(repo *repository.CuatrimestreRepository, moodleClient *moodle.Client) *CuatrimestreService {
	return &CuatrimestreService{Repo: repo, MoodleClient: moodleClient}
}

// CreateLocal crea el registro en la BD local.
func (s *CuatrimestreService) CreateLocal(c *models.Cuatrimestre) error {
	if err := s.validateCuatrimestre(c); err != nil {
		return err
	}
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
	if c.ID == 0 {
		return errors.New("ID de Cuatrimestre inválido")
	}
	if err := s.validateCuatrimestre(c); err != nil {
		return err
	}
	return s.Repo.Update(c)
}

// DeleteLocal elimina el registro en la BD local.
func (s *CuatrimestreService) DeleteLocal(id uint) error {
	if id == 0 {
		return errors.New("ID de Cuatrimestre inválido")
	}
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

	// Si ya tiene ID_Moodle, llamamos a UPDATE en lugar de CREATE
	if cuatrimestre.ID_Moodle != nil {
		log.Printf("Cuatrimestre ID %d ya sincronizado (Moodle ID: %d). Actualizando en Moodle...", id, *cuatrimestre.ID_Moodle)
		return s.UpdateInMoodle(&cuatrimestre)
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

// UpdateInMoodle actualiza un cuatrimestre que ya existe en Moodle
func (s *CuatrimestreService) UpdateInMoodle(cuatrimestre *models.Cuatrimestre) error {
	if cuatrimestre.ID_Moodle == nil {
		return fmt.Errorf("el cuatrimestre no tiene ID de Moodle, debe crearse primero")
	}

	data := []moodle.CategoryUpdateRequest{
		{
			ID:          *cuatrimestre.ID_Moodle,
			Name:        cuatrimestre.Nombre,
			IDNumber:    safeString(cuatrimestre.ID_Externo),
			Description: safeString(cuatrimestre.Descripcion),
		},
	}

	var response interface{}
	err := s.MoodleClient.Call("core_course_update_categories", data, &response)
	if err != nil {
		return fmt.Errorf("fallo al actualizar Cuatrimestre en Moodle: %w", err)
	}

	log.Printf(" Cuatrimestre '%s' (Moodle ID: %d) actualizado exitosamente en Moodle", cuatrimestre.Nombre, *cuatrimestre.ID_Moodle)
	return nil
}

// BulkSyncToMoodle sincroniza masivamente todos los cuatrimestres no sincronizados
func (s *CuatrimestreService) BulkSyncToMoodle() {
	go func() {
		cuatrimestres, err := s.Repo.GetUnsynced()
		if err != nil {
			log.Printf("ERROR: No se pudieron obtener cuatrimestres no sincronizados: %v", err)
			return
		}

		if len(cuatrimestres) == 0 {
			log.Printf("No hay cuatrimestres pendientes de sincronizar.")
			return
		}

		log.Printf("Iniciando sincronización masiva para %d cuatrimestres...", len(cuatrimestres))

		// Separar por programa de estudio para sincronizar en grupos
		programaGroups := make(map[uint][]models.Cuatrimestre)
		for _, c := range cuatrimestres {
			programaGroups[c.ProgramaEstudioID] = append(programaGroups[c.ProgramaEstudioID], c)
		}

		successCount := 0
		errorCount := 0

		for programaID, group := range programaGroups {
			log.Printf("Procesando %d cuatrimestres del Programa ID %d...", len(group), programaID)

			// Verificar que el programa padre esté sincronizado
			if len(group) > 0 && group[0].ProgramaEstudio.ID_Moodle == nil {
				log.Printf(" ADVERTENCIA: ProgramaEstudio ID %d no está sincronizado. Saltando %d cuatrimestres.", programaID, len(group))
				errorCount += len(group)
				continue
			}

			// Construir array para batch create
			parentID := *group[0].ProgramaEstudio.ID_Moodle
			data := make([]moodle.CategoryRequest, len(group))
			for i, c := range group {
				data[i] = moodle.CategoryRequest{
					Name:        c.Nombre,
					Parent:      int(parentID),
					IDNumber:    safeString(c.ID_Externo),
					Description: safeString(c.Descripcion),
				}
			}

			// Llamar a Moodle
			var response []moodle.CategoryResponse
			err := s.MoodleClient.Call("core_course_create_categories", data, &response)
			if err != nil {
				log.Printf(" Error al procesar cuatrimestres del Programa ID %d: %v", programaID, err)
				errorCount += len(group)
				continue
			}

			// Actualizar IDs en BD local
			for i, categoryResp := range response {
				if i < len(group) {
					moodleID := categoryResp.ID
					group[i].ID_Moodle = &moodleID
					if err := s.Repo.Update(&group[i]); err != nil {
						log.Printf(" Error al actualizar cuatrimestre ID %d con Moodle ID %d: %v", group[i].ID, moodleID, err)
						errorCount++
					} else {
						log.Printf(" Cuatrimestre '%s' sincronizado con Moodle ID: %d", group[i].Nombre, moodleID)
						successCount++
					}
				}
			}
		}

		log.Printf(" Sincronización masiva de cuatrimestres finalizada. Exitosos: %d, Errores: %d", successCount, errorCount)
	}()
}

// validateCuatrimestre aplica validaciones de negocio y límites de longitud
func (s *CuatrimestreService) validateCuatrimestre(c *models.Cuatrimestre) error {
	c.Nombre = strings.TrimSpace(c.Nombre)
	if c.ProgramaEstudioID == 0 {
		return errors.New("ProgramaEstudioID es obligatorio")
	}
	if c.Nombre == "" {
		return errors.New("Nombre es obligatorio")
	}
	if utf8.RuneCountInString(c.Nombre) > 255 {
		return errors.New("Nombre excede el máximo de 255 caracteres")
	}
	if c.ID_Externo != nil {
		trimmed := strings.TrimSpace(*c.ID_Externo)
		if utf8.RuneCountInString(trimmed) > 100 {
			return errors.New("ID_Externo excede el máximo de 100 caracteres")
		}
		*c.ID_Externo = trimmed
	}
	if c.Descripcion != nil {
		trimmed := strings.TrimSpace(*c.Descripcion)
		*c.Descripcion = trimmed
	}
	return nil
}
