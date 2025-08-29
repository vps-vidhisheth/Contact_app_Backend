package user

import (
	"log"
	"sync"

	"gorm.io/gorm"
)

type ModuleConfig struct {
	DB *gorm.DB
}

func NewUserModuleConfig(db *gorm.DB) *ModuleConfig {
	return &ModuleConfig{DB: db}
}

func (config *ModuleConfig) TableMigration(wg *sync.WaitGroup) {
	defer wg.Done()

	if err := config.DB.AutoMigrate(&User{}); err != nil {
		log.Println("User Auto Migration Error:", err)
	}

	log.Println("User Table Migrated")
}
