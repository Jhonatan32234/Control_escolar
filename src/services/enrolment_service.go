// internal/services/enrolment_service.go (Nuevo Servicio para Matrícula)

package services

import (
	"log"
	"sync"
	"api_concurrencia/src/models"
)

// Este servicio gestionaría las llamadas a enrol_manual_enrol_users
type EnrolmentService struct {
	// Repositorio y Cliente Moodle
}

// ProcessBulkEnrolment matricula a los alumnos en paralelo.
func (s *EnrolmentService) ProcessBulkEnrolment(enrollments []models.Matricula) {
	var wg sync.WaitGroup // Usamos WaitGroup para esperar que todas las goroutines terminen
	
	batchSize := 50 // Llamar a Moodle con lotes de 50 matrículas a la vez
	
	for i := 0; i < len(enrollments); i += batchSize {
		end := i + batchSize
		if end > len(enrollments) {
			end = len(enrollments)
		}
		
		batch := enrollments[i:end]
		wg.Add(1)
		
		// Lanzamos una goroutine para procesar cada lote de enrolamiento
		go func(b []models.Matricula) {
			defer wg.Done()
			
			// **Lógica de llamada a Moodle:**
			// 1. Convertir 'b' a la estructura requerida por enrol_manual_enrol_users
			// 2. Llamar a la API de Moodle.
			log.Printf("Procesando lote de %d matrículas...", len(b))
			
		}(batch)
	}
	
	wg.Wait() // Bloquea hasta que todas las goroutines llamen a Done()
	log.Println("✅ Matrícula masiva finalizada.")
}