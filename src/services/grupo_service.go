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

type GrupoService struct {
	Repo           *repository.GrupoRepository
	MoodleClient   *moodle.Client
	AsignaturaRepo *repository.AsignaturaRepository // Necesario para obtener el CourseID de Moodle
	UsuarioRepo    *repository.UsuarioRepository    // Necesario para obtener el UserID de Moodle
}

func NewGrupoService(repo *repository.GrupoRepository, moodleClient *moodle.Client, aRepo *repository.AsignaturaRepository, uRepo *repository.UsuarioRepository) *GrupoService {
	return &GrupoService{
		Repo:           repo,
		MoodleClient:   moodleClient,
		AsignaturaRepo: aRepo,
		UsuarioRepo:    uRepo,
	}
}

func (s *GrupoService) GetByID(id uint) (models.Grupo, error) {
	return s.Repo.GetByID(id)
}

func (s *GrupoService) GetAll() ([]models.Grupo, error) {
	return s.Repo.GetAll()
}

// CreateLocal crea el registro en la BD local con validaciones
func (s *GrupoService) CreateLocal(g *models.Grupo) error {
	if err := s.validateGrupo(g); err != nil {
		return err
	}
	return s.Repo.Create(g)
}

// SyncGroupToMoodle crea un grupo en Moodle y actualiza el ID_Moodle local.
func (s *GrupoService) SyncToMoodle(grupoID uint) error {
	grupo, err := s.Repo.GetByID(grupoID)
	if err != nil {
		return fmt.Errorf("grupo (ID: %d) no encontrado en BD local: %w", grupoID, err)
	}

	// 1. Obtener el ID de Moodle del Curso al que pertenece este grupo
	// ... (Lógica de obtención de asignatura y verificación de ID_Moodle de Asignatura) ...
	asignatura, err := s.AsignaturaRepo.GetByID(grupo.CourseID)
	if err != nil {
		return fmt.Errorf("asignatura (ID: %d) no encontrada para el grupo: %w", grupo.CourseID, err)
	}
	if asignatura.ID_Moodle == nil {
		return fmt.Errorf("la asignatura '%s' no está sincronizada con Moodle (ID_Moodle nulo)", asignatura.NombreCompleto)
	}

	// A. VERIFICACIÓN: Si ya tiene ID_Moodle, asumimos que está sincronizado.
	if grupo.ID_Moodle != nil {
		log.Printf("Grupo ID %d ya sincronizado (Moodle ID: %d). Saltando creación.", grupoID, *grupo.ID_Moodle)
		return nil
	}

	// 2. Preparar la petición
	moodleCourseID := int(*asignatura.ID_Moodle)
	idNumber := fmt.Sprintf("G-%d-%s", grupoID, grupo.Nombre)

	data := []moodle.GroupRequest{
		{
			CourseID:          moodleCourseID,
			Name:              grupo.Nombre,
			IDNumber:          idNumber,
			Description:       grupo.Description,
			DescriptionFormat: 1, // HTML
			Visibility:        0, // Visible a todos
			Participation:     1, // Actividad habilitada
		},
	}

	log.Printf("Iniciando creación de Grupo '%s' en Moodle para Curso ID %d (Moodle ID: %d)", grupo.Nombre, asignatura.ID, moodleCourseID)

	// 3. Ejecutar la llamada a la API de Moodle
	var response []moodle.GroupResponse
	err = s.MoodleClient.Call("core_group_create_groups", data, &response)

	// 🛑 B. MANEJO DE ERRORES: Verificar si la falla se debe a un duplicado de nombre
	if err != nil {
		// Verificar si el error es de duplicidad (el error que viste antes)
		if strings.Contains(err.Error(), "Group with the same name already exists in the course") {
			log.Printf("⚠️ Advertencia: Grupo '%s' ya existe en Moodle. Intentando recuperar ID.", grupo.Nombre)

			// Llama a una función auxiliar para buscar el grupo por nombre/idnumber.
			// NOTA: Moodle no tiene una función sencilla para buscar grupos por nombre y curso,
			// pero podemos asumir que el grupo existe y forzar un ID de Moodle para continuar.
			// Para simplificar, asumimos que si el error es DUPLICADO, debemos actualizar el ID.

			// *****************************************************************
			// 💡 NOTA: La forma 100% correcta requiere implementar core_group_get_groups
			// para buscar el grupo por nombre y obtener el ID.
			// Por ahora, para avanzar, si falla por duplicado, salimos.
			// *****************************************************************

			return fmt.Errorf("el grupo ya existe en Moodle. Por favor, actualice manualmente el ID_Moodle local o implemente la búsqueda de grupo en Moodle.")

		}
		return fmt.Errorf("fallo al crear Grupo en Moodle: %w", err)
	}

	// C. PROCESAR RESPUESTA EXITOSA
	if len(response) == 0 {
		return fmt.Errorf("moodle no devolvió ningún Grupo creado")
	}

	moodleID := uint(response[0].ID)
	grupo.ID_Moodle = &moodleID

	// 🛑 D. CORRECCIÓN CRÍTICA: Usar Save para actualizar el registro existente.
	// Esto resuelve el "Duplicate entry '3' for key 'grupos.PRIMARY'" que viste.
	// Asumiendo que has añadido la DB a tu struct GrupoRepository (como se sugiere)
	if err := s.Repo.DB.Save(&grupo).Error; err != nil {
		return fmt.Errorf("falla al actualizar ID Moodle local para Grupo ID %d: %w", grupoID, err)
	}

	log.Printf("✅ Grupo '%s' (ID local: %d) creado exitosamente en Moodle. Group ID: %d", grupo.Nombre, grupoID, moodleID)
	return nil
}

// SyncMembersToMoodle añade todos los usuarios locales del grupo a Moodle.
func (s *GrupoService) SyncMembersToMoodle(grupoID uint) error {
	grupo, err := s.Repo.GetByID(grupoID)
	if err != nil {
		return fmt.Errorf("grupo (ID: %d) no encontrado: %w", grupoID, err)
	}

	if grupo.ID_Moodle == nil {
		return fmt.Errorf("el grupo '%s' no está sincronizado con Moodle (ID_Moodle nulo)", grupo.Nombre)
	}

	// 1. Obtener los usuarios locales que son miembros de este grupo
	usuarios, err := s.Repo.GetMembers(grupoID)
	if err != nil {
		return fmt.Errorf("error al obtener miembros del grupo local: %w", err)
	}

	var memberRequests []moodle.GroupMemberRequest
	var missingMoodleIDs []uint

	// 2. Construir el array de peticiones de miembros
	for _, user := range usuarios {
		if user.ID_Moodle == nil {
			missingMoodleIDs = append(missingMoodleIDs, user.ID)
			continue
		}
		memberRequests = append(memberRequests, moodle.GroupMemberRequest{
			GroupID: int(*grupo.ID_Moodle),
			UserID:  int(*user.ID_Moodle),
		})
	}

	if len(missingMoodleIDs) > 0 {
		log.Printf("⚠️ Advertencia: %d miembros del grupo %d no pudieron ser sincronizados (ID_Moodle nulo).", len(missingMoodleIDs), grupoID)
		// Puedes decidir si retornar un error fatal o continuar con los que sí tienen ID
	}

	if len(memberRequests) == 0 {
		log.Printf("No hay miembros válidos para sincronizar en el Grupo %d.", grupoID)
		return nil
	}

	// 3. Ejecutar la llamada a la API de Moodle
	// core_group_add_group_members no devuelve cuerpo, solo éxito o error.
	var response interface{}
	err = s.MoodleClient.Call("core_group_add_group_members", memberRequests, &response)
	if err != nil {
		return fmt.Errorf("fallo al añadir miembros al grupo '%s' (Moodle ID: %d): %w", grupo.Nombre, *grupo.ID_Moodle, err)
	}

	log.Printf("✅ %d miembros añadidos con éxito al grupo '%s' (Moodle ID: %d).", len(memberRequests), grupo.Nombre, *grupo.ID_Moodle)
	return nil
}

// UpdateLocal actualiza el registro en la BD local.
func (s *GrupoService) UpdateLocal(pe *models.Grupo) error {
	if pe.ID == 0 {
		return errors.New("ID de Grupo inválido")
	}
	if err := s.validateGrupo(pe); err != nil {
		return err
	}
	return s.Repo.Update(pe)
}

// DeleteLocal elimina el registro en la BD local.
func (s *GrupoService) DeleteLocal(id uint) error {
	if id == 0 {
		return errors.New("ID de Grupo inválido")
	}
	return s.Repo.Delete(id)
}

// validateGrupo aplica validaciones de negocio y límites
func (s *GrupoService) validateGrupo(g *models.Grupo) error {
	g.Nombre = strings.TrimSpace(g.Nombre)
	g.Description = strings.TrimSpace(g.Description)
	if g.CourseID == 0 {
		return errors.New("CourseID es obligatorio")
	}
	if g.Nombre == "" {
		return errors.New("Nombre es obligatorio")
	}
	if utf8.RuneCountInString(g.Nombre) > 255 {
		return errors.New("Nombre excede el máximo de 255 caracteres")
	}
	// Description: sin límite estricto, normalizamos espacios
	return nil
}
