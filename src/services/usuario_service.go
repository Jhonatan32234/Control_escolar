package services

import (
	"api_concurrencia/src/models"
	"api_concurrencia/src/repository"
	"log"
	"sync"
)

type UsuarioService struct {
	Repo *repository.UsuarioRepository
	// MoodleClient *moodle.Client // Cliente para la API de Moodle
}

func NewUsuarioService(repo *repository.UsuarioRepository) *UsuarioService {
	return &UsuarioService{Repo: repo}
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


// SyncToMoodle (Para un solo usuario).
func (s *UsuarioService) SyncToMoodle(id uint) error {
	log.Printf("Iniciando sincronización individual del Usuario ID %d...", id)
	// Lógica para enviar a Moodle usando core_user_create_users([un solo usuario])
	// y actualizar el ID_Moodle local.
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
			
			// **Lógica de llamada a Moodle:**
			// 1. Convertir el lote a la estructura requerida por core_user_create_users.
			// 2. Llamar a la API de Moodle.
			
			// 3. (Importante) Recorrer la respuesta de Moodle para actualizar el ID_Moodle de cada usuario en la BD local.
			
		}(batch)
	}

	wg.Wait() // Espera a que todas las goroutines del lote terminen.
}