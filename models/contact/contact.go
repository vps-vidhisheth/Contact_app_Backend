package contact

import (
	"Contact_App/models/contact_detail"
)

type Contact struct {
	ContactID   int                             `gorm:"primaryKey;autoIncrement" json:"contact_id"`
	FName       string                          `gorm:"not null" json:"f_name"`
	LName       string                          `json:"l_name"`
	IsActive    bool                            `gorm:"default:true" json:"is_active"`
	UserID      int                             `gorm:"not null" json:"user_id"`
	Details     []*contact_detail.ContactDetail `gorm:"foreignKey:ContactID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"details"`
	DetailCount int                             `gorm:"-" json:"-"` // Not persisted in DB
}

// Increment and return detail count (runtime only)
func (c *Contact) GetDetailCounterAndIncrement() int {
	c.DetailCount++
	return c.DetailCount
}

// Add a detail to the contact (runtime only)
func (c *Contact) AddContactDetail(d *contact_detail.ContactDetail) {
	c.Details = append(c.Details, d)
}

// Get all details (runtime only)
func (c *Contact) GetDetails() []*contact_detail.ContactDetail {
	return c.Details
}

// Delete a detail by its ID (runtime only)
func (c *Contact) DeleteDetailByID(detailID int) error {
	for i, d := range c.Details {
		if d.ContactDetailsID == detailID { // Correct field name
			c.Details = append(c.Details[:i], c.Details[i+1:]...)
			return nil
		}
	}
	return ErrContactDetailNotFound
}
