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

type AsignaturaService struct {
	Repo         *repository.AsignaturaRepository
	MoodleClient *moodle.Client
}

func NewAsignaturaService(repo *repository.AsignaturaRepository, moodleClient *moodle.Client) *AsignaturaService {
	return &AsignaturaService{Repo: repo, MoodleClient: moodleClient}
}

// CreateLocal crea el registro en la BD local.
func (s *AsignaturaService) CreateLocal(a *models.Asignatura) error {
	if err := s.validateAsignatura(a); err != nil {
		return err
	}
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
	if a.ID == 0 {
		return errors.New("ID de Asignatura inválido")
	}
	if err := s.validateAsignatura(a); err != nil {
		return err
	}
	return s.Repo.Update(a)
}

// DeleteLocal elimina el registro en la BD local.
func (s *AsignaturaService) DeleteLocal(id uint) error {
	if id == 0 {
		return errors.New("ID de Asignatura inválido")
	}
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

	// Si ya tiene ID_Moodle, actualizamos en lugar de crear
	if asignatura.ID_Moodle != nil {
		log.Printf("Asignatura ID %d ya sincronizada (Moodle ID: %d). Actualizando en Moodle.", id, *asignatura.ID_Moodle)
		return s.UpdateInMoodle(&asignatura)
	}

	// 1. Construir el array de datos para la función de Moodle
	moodleParentID := *asignatura.Cuatrimestre.ID_Moodle

	// **USAMOS EL STRUCT DE CURSO (CourseRequest)**
	data := []moodle.CourseRequest{
		{
			Fullname:   asignatura.NombreCompleto, // 👈 DATOS DE LA ASIGNATURA
			Shortname:  asignatura.NombreCorto,    // 👈 REQUERIDO: Nombre corto único
			Categoryid: int(moodleParentID),       // 👈 ID MOODLE del Cuatrimestre padre
			IDNumber:   safeString(asignatura.ID_Externo),
			Summary:    safeString(asignatura.Resumen),
		},
	}

	// 2. Ejecutar la llamada a la API de Moodle
	var response []moodle.CourseResponse                                     // 👈 USAMOS EL STRUCT DE RESPUESTA DE CURSO
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

// validateAsignatura aplica validaciones de negocio y límites de longitud
func (s *AsignaturaService) validateAsignatura(a *models.Asignatura) error {
	a.NombreCompleto = strings.TrimSpace(a.NombreCompleto)
	a.NombreCorto = strings.TrimSpace(a.NombreCorto)

	if a.CuatrimestreID == 0 {
		return errors.New("CuatrimestreID es obligatorio")
	}
	if a.NombreCompleto == "" {
		return errors.New("NombreCompleto es obligatorio")
	}
	if a.NombreCorto == "" {
		return errors.New("NombreCorto es obligatorio")
	}
	if utf8.RuneCountInString(a.NombreCompleto) > 255 {
		return errors.New("NombreCompleto excede el máximo de 255 caracteres")
	}
	if utf8.RuneCountInString(a.NombreCorto) > 100 {
		return errors.New("NombreCorto excede el máximo de 100 caracteres")
	}
	if a.ID_Externo != nil {
		trimmed := strings.TrimSpace(*a.ID_Externo)
		if utf8.RuneCountInString(trimmed) > 100 {
			return errors.New("ID_Externo excede el máximo de 100 caracteres")
		}
		// normalize back
		*a.ID_Externo = trimmed
	}
	if a.Resumen != nil {
		// No límite estricto aquí; dejar al tipo TEXT en DB. Se puede normalizar espacios.
		trimmed := strings.TrimSpace(*a.Resumen)
		*a.Resumen = trimmed
	}
	return nil
}

// UpdateInMoodle actualiza una asignatura existente en Moodle
func (s *AsignaturaService) UpdateInMoodle(a *models.Asignatura) error {
	if a.ID_Moodle == nil {
		return errors.New("la asignatura no tiene ID_Moodle, no se puede actualizar")
	}

	data := []moodle.CourseUpdateRequest{
		{
			ID:        uint(*a.ID_Moodle),
			Fullname:  a.NombreCompleto,
			Shortname: a.NombreCorto,
			IDNumber:  safeString(a.ID_Externo),
			Summary:   safeString(a.Resumen),
		},
	}

	var response []moodle.CourseResponse
	err := s.MoodleClient.Call("core_course_update_courses", data, &response)
	if err != nil {
		return fmt.Errorf("fallo al actualizar curso/asignatura en Moodle: %w", err)
	}

	log.Printf(" Asignatura '%s' (ID local: %d, Moodle ID: %d) actualizada exitosamente en Moodle", a.NombreCompleto, a.ID, *a.ID_Moodle)
	return nil
}

// BulkSyncToMoodle sincroniza todas las asignaturas sin ID_Moodle a Moodle
func (s *AsignaturaService) BulkSyncToMoodle() {
	go func() {
		log.Println(" Iniciando sincronización masiva de Asignaturas a Moodle...")

		// Obtener todas las asignaturas sin ID_Moodle
		asignaturas, err := s.Repo.GetUnsynced()
		if err != nil {
			log.Printf(" Error al obtener asignaturas sin sincronizar: %v", err)
			return
		}

		if len(asignaturas) == 0 {
			log.Println(" No hay asignaturas pendientes de sincronización")
			return
		}

		log.Printf(" Encontradas %d asignaturas para sincronizar", len(asignaturas))

		// Agrupar asignaturas por CuatrimestreID para sincronización eficiente
		cuatrimestreGroups := make(map[uint][]models.Asignatura)
		for _, asignatura := range asignaturas {
			cuatrimestreGroups[asignatura.CuatrimestreID] = append(cuatrimestreGroups[asignatura.CuatrimestreID], asignatura)
		}

		successCount := 0
		errorCount := 0

		// Procesar cada grupo de asignaturas por cuatrimestre
		for cuatrimestreID, group := range cuatrimestreGroups {
			log.Printf(" Procesando %d asignaturas del Cuatrimestre ID: %d", len(group), cuatrimestreID)

			// Validar que el cuatrimestre padre esté sincronizado
			if group[0].Cuatrimestre.ID_Moodle == nil {
				log.Printf("  Cuatrimestre ID %d no tiene ID_Moodle. Saltando %d asignaturas.", cuatrimestreID, len(group))
				errorCount += len(group)
				continue
			}

			// Construir array de CourseRequest para este grupo
			data := make([]moodle.CourseRequest, len(group))
			for i, asignatura := range group {
				data[i] = moodle.CourseRequest{
					Fullname:   asignatura.NombreCompleto,
					Shortname:  asignatura.NombreCorto,
					Categoryid: int(*asignatura.Cuatrimestre.ID_Moodle),
					IDNumber:   safeString(asignatura.ID_Externo),
					Summary:    safeString(asignatura.Resumen),
				}
			}

			// Llamar a la API de Moodle para crear cursos en batch
			var response []moodle.CourseResponse
			err := s.MoodleClient.Call("core_course_create_courses", data, &response)
			if err != nil {
				log.Printf(" Error al crear cursos en Moodle para Cuatrimestre ID %d: %v", cuatrimestreID, err)
				errorCount += len(group)
				continue
			}

			// Actualizar ID_Moodle en la base de datos local
			for i, asignatura := range group {
				if i < len(response) {
					moodleID := response[i].ID
					asignatura.ID_Moodle = &moodleID

					if err := s.Repo.Update(&asignatura); err != nil {
						log.Printf(" Error al actualizar ID_Moodle para Asignatura ID %d: %v", asignatura.ID, err)
						errorCount++
					} else {
						log.Printf(" Asignatura '%s' (ID local: %d) sincronizada con Moodle ID: %d", asignatura.NombreCompleto, asignatura.ID, moodleID)
						successCount++
					}
				} else {
					log.Printf(" No se recibió respuesta de Moodle para Asignatura ID %d", asignatura.ID)
					errorCount++
				}
			}
		}

		log.Printf(" Sincronización masiva completada: %d exitosas, %d errores", successCount, errorCount)
	}()
}
