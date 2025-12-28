package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB ahora recibe la cadena de conexión (dsn) como parámetro
func InitDB(dsn string) {
	var err error

	// Configuración de GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 1. Abrir conexión con la URL que recibimos
	DB, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		log.Fatalf("❌ Error crítico: No se pudo abrir la sesión con GORM: %v", err)
	}

	// 2. Configurar el Pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("❌ Error obteniendo instancia SQL: %v", err)
	}

	sqlDB.SetMaxOpenConns(15)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 3. Verificar conexión
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Error haciendo ping a la BD (Revisar credenciales): %v", err)
	}

	log.Println("✅ Conexión a Base de Datos establecida correctamente via GORM")
}
