package migration

import (
	"api_concurrencia/src/models"
	"log"

	"gorm.io/gorm"
)

// AutoMigrateTables ejecuta las migraciones para todas las entidades.
func AutoMigrateTables(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.ProgramaEstudio{},
		&models.Cuatrimestre{},
		&models.Asignatura{},
		&models.Docente{},
		&models.Grupo{},
		&models.Usuario{},
		&models.Matricula{},
	)

	if err != nil {
		log.Fatalf("Error durante la migración de tablas: %v", err)
	}

	log.Println("✅ Migraciones de tablas completadas exitosamente.")
}
