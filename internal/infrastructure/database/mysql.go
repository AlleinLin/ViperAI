package database

import (
	"fmt"
	"time"

	"viperai/internal/config"
	"viperai/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Initialize() error {
	cfg := config.Get().Database

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset,
	)

	var logMode logger.Interface
	if gin.Mode() == gin.DebugMode {
		logMode = logger.Default.LogMode(logger.Info)
	} else {
		logMode = logger.Default
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                      dsn,
		DefaultStringSize:        256,
		DisableDatetimePrecision: true,
		DontSupportRenameIndex:   true,
		DontSupportRenameColumn:  true,
	}), &gorm.Config{
		Logger: logMode,
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db

	return autoMigrate()
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&domain.User{},
		&domain.Conversation{},
		&domain.ChatMessage{},
	)
}

func GetDB() *gorm.DB {
	return DB
}
