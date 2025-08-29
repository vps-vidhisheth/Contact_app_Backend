package contact_detail

import (
	"log"
	"sync"

	"gorm.io/gorm"
)

type ModuleConfig struct {
	DB *gorm.DB
}

func NewContactDetailModuleConfig(db *gorm.DB) *ModuleConfig {
	return &ModuleConfig{
		DB: db,
	}
}

func (config *ModuleConfig) TableMigration(wg *sync.WaitGroup) {
	defer wg.Done()

	if err := config.DB.AutoMigrate(&ContactDetail{}); err != nil {
		log.Println("ContactDetail Auto Migration Error:", err)
	}

	log.Println("ContactDetail Table Migrated")
}
