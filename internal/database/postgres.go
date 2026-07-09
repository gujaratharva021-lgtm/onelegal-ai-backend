package database

import (
	"fmt"
	"log"
	"time"

	"legaltech-backend/internal/config"
	"legaltech-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// Left at driver defaults (unlimited open conns, MaxIdleConns=2) this pool
	// can exhaust Postgres's max_connections under concurrent request load and
	// churns idle connections needlessly — bound it explicitly.
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to access underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.PasswordResetToken{},
		&models.Session{},
		&models.Client{},
		&models.Case{},
		&models.CaseTimelineEvent{},
		&models.Task{},
		&models.Hearing{},
		&models.CalendarEvent{},
		&models.Draft{},
		&models.ResearchHistory{},
		&models.Document{},
		&models.Meeting{},
		&models.Note{},
		&models.Notification{},
		&models.AIConversation{},
		&models.AIMessage{},
		&models.AIDocumentAnalysis{},
		&models.AIRecommendation{},
		&models.VideoRoom{},
		&models.Invoice{},
		&models.Payment{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	DB = db
	repairStaleVideoRooms()
}

// repairStaleVideoRooms is a one-time data fix for rooms created before the
// ringing/active status split existed: they were stamped status="active"
// (the old default) the instant they were created, even though the client
// never actually answered (started_at is still null). Left alone, those
// rows would incorrectly keep matching the "is this client on an active
// call" busy check forever. Anything never answered and not already ended
// is closed out here as a missed call so busy detection only ever reflects
// calls someone genuinely picked up.
func repairStaleVideoRooms() {
	result := DB.Model(&models.VideoRoom{}).
		Where("status = ? AND started_at IS NULL AND ended_at IS NULL", models.VideoRoomStatusActive).
		Updates(map[string]interface{}{
			"status":   models.VideoRoomStatusEnded,
			"outcome":  models.VideoRoomOutcomeMissed,
			"ended_at": time.Now(),
		})
	if result.Error != nil {
		log.Printf("Failed to repair stale video rooms: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("Repaired %d stale video call room(s) stuck in a false 'active/busy' state", result.RowsAffected)
	}
}
