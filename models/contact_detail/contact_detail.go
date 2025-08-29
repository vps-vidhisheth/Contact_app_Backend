package contact_detail

type ContactDetail struct {
	ContactDetailsID uint   `gorm:"primaryKey;autoIncrement;type:BIGINT UNSIGNED" json:"contact_details_id"`
	UserID           uint   `gorm:"not null;index;type:BIGINT UNSIGNED" json:"user_id"`
	ContactID        uint   `gorm:"not null;index;type:BIGINT UNSIGNED" json:"contact_id"`
	Type             string `gorm:"not null" json:"type"`
	Value            string `gorm:"not null" json:"value"`
	IsActive         bool   `gorm:"default:true" json:"is_active"`
}
