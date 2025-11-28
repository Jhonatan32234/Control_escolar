package main

import (
	"log"
	"net/http"
	"os"

	"api_concurrencia/pkg/migration"
	"api_concurrencia/src/handlers"
	"api_concurrencia/src/moodle"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const defaultPort = "8080"

func main() {
	// 1. Configuración de la Base de Datos
	godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// DSN de ejemplo, ¡DEBE SER REEMPLAZADO con tu configuración!
		dsn = "admin:admin@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
		log.Println("⚠️ Usando DATABASE_URL por defecto. Asegúrate de configurar la variable de entorno.")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ No se pudo conectar a la base de datos MySQL: %v", err)
	}
	log.Println("✅ Conexión a la base de datos establecida exitosamente.")

	// 2. Ejecución de Migraciones
	// Esto creará o actualizará todas las tablas (ProgramaEstudio, Cuatrimestre, etc.)
	migration.AutoMigrateTables(db)

	moodleClient := moodle.NewClient()
    log.Println("✅ Cliente de Moodle inicializado.")
	// 3. Inicialización del Router y las Rutas
	router := handlers.Routes(db, moodleClient)

	// 4. Inicialización del Servidor HTTP
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("🌐 Servidor escuchando en http://localhost:%s", port)
	
	// El router (chi.Mux) implementa la interfaz http.Handler
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}