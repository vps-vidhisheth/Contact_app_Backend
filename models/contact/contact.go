package contact

import (
	"Contact_App/models/contact_detail"

	"gorm.io/gorm"
)

type Contact struct {
	ContactID uint   `gorm:"column:contact_id;primaryKey;autoIncrement;type:BIGINT UNSIGNED" json:"contact_id"`
	UserID    uint   `gorm:"column:user_id;not null;type:BIGINT UNSIGNED" json:"user_id"`
	FName     string `gorm:"column:f_name;not null" json:"first_name"`
	LName     string `gorm:"column:l_name;not null" json:"last_name"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`

	Details   []*contact_detail.ContactDetail `gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE" json:"details"`
	DeletedAt gorm.DeletedAt                  `gorm:"index" json:"-"`
}
