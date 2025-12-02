package main

import (
	"log"
	"net/http"
	"os"

	"api_concurrencia/pkg/migration"
	"api_concurrencia/src/handlers"
	"api_concurrencia/src/moodle"

	"crypto/tls"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const defaultPort = "8080"

func main() {
	// 1. Configuración de la Base de Datos (TiDB Cloud / MySQL)
	godotenv.Load()
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USERNAME")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_DATABASE")

	var dsn string
	if host != "" && dbPort != "" && user != "" && pass != "" && name != "" {
		// TiDB Cloud (MySQL) requiere conexión segura (TLS)
		// Registramos un perfil TLS y lo referenciamos en el DSN con `tls=tidb`
		tlsConfigName := "tidb"
		_ = mysqlDriver.RegisterTLSConfig(tlsConfigName, &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		})
		dsn = user + ":" + pass + "@tcp(" + host + ":" + dbPort + ")/" + name + "?charset=utf8mb4&parseTime=True&loc=Local&tls=" + tlsConfigName
	} else if os.Getenv("DATABASE_URL") != "" {
		dsn = os.Getenv("DATABASE_URL")
	} else {
		dsn = "admin:admin@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
		log.Println("Variables DB_* no configuradas. Usando DSN local por defecto.")
	}

	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
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
