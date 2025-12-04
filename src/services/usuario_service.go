package services

import (
	"api_concurrencia/src/models"
	"api_concurrencia/src/moodle"
	"api_concurrencia/src/repository"
	"fmt"
	"log"
	"sync"
)

type UsuarioService struct {
	Repo           *repository.UsuarioRepository
	MoodleClient   *moodle.Client                   // Cliente para la API de Moodle
	AsignaturaRepo *repository.AsignaturaRepository // Repositorio para Asignaturas
}

func NewUsuarioService(repo *repository.UsuarioRepository, moodleClient *moodle.Client, asignaturaRepo *repository.AsignaturaRepository) *UsuarioService {
	return &UsuarioService{Repo: repo, MoodleClient: moodleClient, AsignaturaRepo: asignaturaRepo}
}

// (Implementar CreateLocal, GetByID, GetAll, UpdateLocal, DeleteLocal) ...

// CreateLocal crea el registro en la BD local.
func (s *UsuarioService) CreateLocal(a *models.Usuario) error {
	// Se requeriría la importación de "errors" si el paquete no lo tiene por defecto.
	return s.Repo.Create(a)
}

// GetAll recupera todas las Asignaturas.
func (s *UsuarioService) GetAll() ([]models.Usuario, error) {
	return s.Repo.GetAll()
}

// GetByID recupera una Asignatura.
func (s *UsuarioService) GetByID(id uint) (models.Usuario, error) {
	return s.Repo.GetByID(id)
}

// UpdateLocal actualiza el registro en la BD local.
func (s *UsuarioService) UpdateLocal(a *models.Usuario) error {
	return s.Repo.Update(a)
}

// DeleteLocal elimina el registro en la BD local.
func (s *UsuarioService) DeleteLocal(id uint) error {
	return s.Repo.Delete(id)
}

// GetUnsyncedByRole recupera los usuarios pendientes de sincronización para un rol específico.
func (s *UsuarioService) GetUnsyncedByRole(role string) ([]models.Usuario, error) {
	return s.Repo.GetUnsyncedByRole(role)
}

// GetByGroupID recupera todos los usuarios que pertenecen a un grupo específico.
func (s *UsuarioService) GetByGroupID(grupoID uint) ([]models.Usuario, error) {
	return s.Repo.GetByGroupID(grupoID)
}

// SyncToMoodle (Para un solo usuario).
func (s *UsuarioService) SyncToMoodle(id uint) error {
	usuario, err := s.Repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("usuario (ID: %d) no encontrado en BD local: %w", id, err)
	}

	// Si ya tiene ID_Moodle, llamamos a UPDATE en lugar de CREATE
	if usuario.ID_Moodle != nil {
		log.Printf("Usuario ID %d ya sincronizado (Moodle ID: %d). Actualizando en Moodle...", id, *usuario.ID_Moodle)
		return s.UpdateInMoodle(&usuario)
	}

	// 1. Construir el array de datos para la función de Moodle
	data := []moodle.UserRequest{
		{
			Username:  usuario.Username,
			Password:  usuario.Password,  // Usamos la contraseña hasheada/temporal de tu modelo
			Firstname: usuario.FirstName, // Usamos FirstName
			Lastname:  usuario.LastName,  // Usamos LastName
			Email:     usuario.Email,
			// 👈 Mapeamos Matricula a IDNumber de Moodle
			IDNumber: safeString(usuario.Matricula),
		},
	}

	// 2. Ejecutar la llamada a la API de Moodle
	var response []moodle.UserResponse
	err = s.MoodleClient.Call("core_user_create_users", data, &response)
	if err != nil {
		return fmt.Errorf("fallo al crear Usuario en Moodle: %w", err)
	}

	// 3. Procesar la respuesta y actualizar el ID_Moodle local
	if len(response) == 0 {
		return fmt.Errorf("moodle no devolvió ningún Usuario creado")
	}

	moodleID := response[0].ID
	usuario.ID_Moodle = &moodleID

	if err := s.Repo.Update(&usuario); err != nil {
		return fmt.Errorf("falla al actualizar ID Moodle local para Usuario ID %d: %w", id, err)
	}

	log.Printf("✅ Usuario '%s' (ID local: %d) creado exitosamente en Moodle. User ID: %d", usuario.Username, id, moodleID)
	return nil
}

// UpdateInMoodle actualiza un usuario que ya existe en Moodle
func (s *UsuarioService) UpdateInMoodle(usuario *models.Usuario) error {
	if usuario.ID_Moodle == nil {
		return fmt.Errorf("el usuario no tiene ID de Moodle, debe crearse primero")
	}

	data := []moodle.UserUpdateRequest{
		{
			ID:        *usuario.ID_Moodle,
			Username:  usuario.Username,
			Firstname: usuario.FirstName,
			Lastname:  usuario.LastName,
			Email:     usuario.Email,
			IDNumber:  safeString(usuario.Matricula),
			// Password solo se envía si se cambió (deberías tener un flag para esto)
		},
	}

	// Moodle NO devuelve datos en core_user_update_users, solo confirma sin errores
	var response interface{}
	err := s.MoodleClient.Call("core_user_update_users", data, &response)
	if err != nil {
		return fmt.Errorf("fallo al actualizar Usuario en Moodle: %w", err)
	}

	log.Printf("Usuario '%s' (Moodle ID: %d) actualizado exitosamente en Moodle", usuario.Username, *usuario.ID_Moodle)
	return nil
}

// BulkSyncToMoodle lanza una tarea masiva y concurrente para crear usuarios.
// Usamos un goroutine para no bloquear el API.
func (s *UsuarioService) BulkSyncToMoodle(role string) {
	go func() {
		usuarios, err := s.Repo.GetUnsyncedByRole(role)
		if err != nil {
			log.Printf("ERROR: No se pudieron obtener usuarios no sincronizados para el rol %s: %v", role, err)
			return
		}

		if len(usuarios) == 0 {
			log.Printf("No hay usuarios de rol %s pendientes de sincronizar.", role)
			return
		}

		log.Printf("Iniciando sincronización masiva para %d usuarios de rol %s...", len(usuarios), role)

		s.processInBatches(usuarios)

		log.Printf("✅ Sincronización masiva de usuarios de rol %s finalizada.", role)
	}()
}

func translateRoleToMoodleID(rol string) (int, error) {
	switch rol {
	case "Docente":
		return 3, nil // 3 = Rol de Profesor (Teacher/Editing teacher)
	case "Alumno":
		return 5, nil // 5 = Rol de Estudiante (Student)
	default:
		return 0, fmt.Errorf("rol local desconocido: %s", rol)
	}
}

func (s *UsuarioService) MatricularUsuario(usuarioID, asignaturaID uint) error {
	// 1. Obtener el Usuario local (para ID_Moodle y Rol)
	usuario, err := s.Repo.GetByID(usuarioID)
	if err != nil {
		return fmt.Errorf("usuario (ID: %d) no encontrado: %w", usuarioID, err)
	}

	// 2. Obtener la Asignatura local (para su ID_Moodle)
	asignatura, err := s.AsignaturaRepo.GetByID(asignaturaID)
	if err != nil {
		return fmt.Errorf("asignatura (ID: %d) no encontrada: %w", asignaturaID, err)
	}

	// 3. Validaciones de IDs de Moodle
	if usuario.ID_Moodle == nil {
		return fmt.Errorf("el usuario '%s' no está sincronizado con Moodle (ID_Moodle local es nulo)", usuario.Username)
	}
	if asignatura.ID_Moodle == nil {
		return fmt.Errorf("la asignatura '%s' no está sincronizada con Moodle (ID_Moodle local es nulo)", asignatura.NombreCompleto)
	}

	// 4. Traducir el Rol de Usuario al RoleID de Moodle
	moodleRoleID, err := translateRoleToMoodleID(usuario.Rol)
	if err != nil {
		return fmt.Errorf("no se pudo matricular: %w", err)
	}

	// Convertir a uint para los modelos
	moodleRoleIDUint := uint(moodleRoleID)

	// 5. Construir el array de datos para la API de Moodle
	data := []moodle.EnrolmentRequest{
		{
			RoleID:   moodleRoleID,
			UserID:   *usuario.ID_Moodle,
			CourseID: *asignatura.ID_Moodle,
		},
	}

	// 6. Ejecutar la llamada a la API de Moodle
	var response interface{}
	err = s.MoodleClient.Call("enrol_manual_enrol_users", data, &response)
	if err != nil {
		return fmt.Errorf("fallo al matricular usuario '%s' en curso '%s' en Moodle: %w", usuario.Username, asignatura.NombreCompleto, err)
	}

	// 7. 🚀 NUEVO PASO: GUARDAR REFERENCIA EN LA TABLA MATRICULA LOCAL
	matricula := models.Matricula{
		UsuarioID:      usuarioID,
		AsignaturaID:   asignaturaID,
		UserMoodleID:   *usuario.ID_Moodle,
		CourseMoodleID: *asignatura.ID_Moodle,
		RoleID:         moodleRoleIDUint,
	}

	if err := s.Repo.SaveMatricula(matricula); err != nil {
		// La restricción de índice único compuesto en la tabla Matricula evita duplicados.
		log.Printf("⚠️ ADVERTENCIA: La matrícula fue exitosa en Moodle, pero falló al guardar la referencia local: %v", err)
		return fmt.Errorf("matrícula exitosa en Moodle, pero falló la referencia local: %w", err)
	}

	log.Printf("✅ Usuario %s (ID Moodle: %d) matriculado con éxito en el curso '%s' (ID Moodle: %d) con RoleID: %d y referencia local guardada.",
		usuario.Username, *usuario.ID_Moodle, asignatura.NombreCompleto, *asignatura.ID_Moodle, moodleRoleID)

	return nil
}

// processInBatches divide los usuarios en lotes y los procesa concurrentemente.
func (s *UsuarioService) processInBatches(usuarios []models.Usuario) {
	var wg sync.WaitGroup
	batchSize := 100 // Lotes de 100 usuarios por llamada a la API de Moodle

	for i := 0; i < len(usuarios); i += batchSize {
		end := i + batchSize
		if end > len(usuarios) {
			end = len(usuarios)
		}

		batch := usuarios[i:end]
		wg.Add(1)

		// Ejecutamos cada lote en una goroutine separada
		go func(b []models.Usuario) {
			defer wg.Done()

			log.Printf("-> Procesando lote de %d usuarios...", len(b))

			// Construir array de UserRequest para Moodle
			data := make([]moodle.UserRequest, len(b))
			for i, usuario := range b {
				data[i] = moodle.UserRequest{
					Username:  usuario.Username,
					Password:  usuario.Password,
					Firstname: usuario.FirstName,
					Lastname:  usuario.LastName,
					Email:     usuario.Email,
					IDNumber:  safeString(usuario.Matricula),
				}
			}

			// Llamar a la API de Moodle
			var response []moodle.UserResponse
			err := s.MoodleClient.Call("core_user_create_users", data, &response)
			if err != nil {
				log.Printf("❌ Error al procesar lote: %v", err)
				return
			}

			// Actualizar ID_Moodle en BD local
			for i, userResp := range response {
				if i < len(b) {
					moodleID := userResp.ID
					b[i].ID_Moodle = &moodleID
					if err := s.Repo.Update(&b[i]); err != nil {
						log.Printf("⚠️ Error al actualizar usuario ID %d con Moodle ID %d: %v", b[i].ID, moodleID, err)
					} else {
						log.Printf("✅ Usuario '%s' sincronizado con Moodle ID: %d", b[i].Username, moodleID)
					}
				}
			}

		}(batch)
	}

	wg.Wait() // Espera a que todas las goroutines del lote terminen.
} // CheckUniqueFields delega la verificación de unicidad al repositorio.
func (s *UsuarioService) CheckUniqueFields(u *models.Usuario) (bool, error) {
	// Nota: El repositorio es responsable de buscar duplicados por Username, Email o Matricula.
	return s.Repo.ExistsByUniqueFields(u)
}

// GetByUsername busca un usuario por su username
func (s *UsuarioService) GetByUsername(username string) (*models.Usuario, error) {
	return s.Repo.GetByUsername(username)
}
