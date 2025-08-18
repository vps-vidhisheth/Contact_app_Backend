package contact_detail

import (
	"Contact_App/apperror"
	"strings"
)

type ContactDetail struct {
	ContactDetailsID int    `gorm:"primaryKey;autoIncrement" json:"contact_details_id"`
	UserID           int    `gorm:"not null" json:"user_id"`
	ContactID        int    `gorm:"not null" json:"contact_id"`
	Type             string `gorm:"not null" json:"type"`
	Value            string `gorm:"not null" json:"value"`
	IsActive         bool   `json:"is_active"`
}

func (d *ContactDetail) updateType(value interface{}) error {
	v, ok := value.(string)
	v = strings.ToLower(strings.TrimSpace(v))
	if !ok || (v != "email" && v != "phone") {
		return apperror.NewValidationError("type", "must be either 'email' or 'phone'")
	}
	d.Type = v
	return nil
}

func (d *ContactDetail) updateValue(value interface{}) error {
	v, ok := value.(string)
	v = strings.TrimSpace(v)
	if !ok || v == "" {
		return apperror.NewValidationError("value", "must be a non-empty string")
	}
	d.Value = v
	return nil
}
