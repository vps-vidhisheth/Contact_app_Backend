package contact

import (
	"Contact_App/models/contact_detail"
)

type Contact struct {
	ContactID int    `gorm:"primaryKey;autoIncrement" json:"contact_id"`
	FName     string `gorm:"not null" json:"f_name"`
	LName     string `json:"l_name"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`

	UserID int `gorm:"not null" json:"user_id"`

	Details []*contact_detail.ContactDetail `gorm:"foreignKey:ContactID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"details"`

	detailCount int `gorm:"-" json:"-"`
}

func (c *Contact) GetDetailCounterAndIncrement() int {
	c.detailCount++
	return c.detailCount
}

func (c *Contact) AddContactDetail(d *contact_detail.ContactDetail) {
	c.Details = append(c.Details, d)
}

func (c *Contact) GetDetails() []*contact_detail.ContactDetail {
	return c.Details
}

func (c *Contact) DeleteDetailByID(detailID int) error {
	for i, d := range c.Details {
		if d.ContactDetailsID == detailID {
			c.Details = append(c.Details[:i], c.Details[i+1:]...)
			return nil
		}
	}
	return ErrContactDetailNotFound
}
