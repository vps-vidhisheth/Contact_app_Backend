package contact

import (
	"log"
	"sync"

	"github.com/jinzhu/gorm"
)

type ModuleConfig struct {
	DB *gorm.DB
}

func NewContactModuleConfig(db *gorm.DB) *ModuleConfig {
	return &ModuleConfig{
		DB: db,
	}
}

func (config *ModuleConfig) TableMigration(wg *sync.WaitGroup) {
	defer wg.Done()

	if err := config.DB.AutoMigrate(&Contact{}).Error; err != nil {
		log.Println("Contact Auto Migration Error:", err)
	}

	if err := config.DB.Model(&Contact{}).
		AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE").Error; err != nil {
		log.Println("Contact Foreign Key Error:", err)
	}

	log.Println("Contact Table Migrated")
}
