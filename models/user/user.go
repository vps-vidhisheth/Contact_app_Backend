package user

import (
	"Contact_App/models/contact"
)

type User struct {
	UserID   int    `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`
	FName    string `gorm:"column:f_name;not null" json:"first_name"`
	LName    string `gorm:"column:l_name" json:"last_name"`
	Email    string `gorm:"column:email;unique;not null" json:"email"`
	Password string `gorm:"column:password" json:"password"`
	IsAdmin  bool   `gorm:"column:is_admin;default:false" json:"is_admin"`
	IsActive bool   `gorm:"column:is_active;default:true" json:"is_active"`

	Contacts []*contact.Contact `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"contacts"`
}

func (u *User) IsAdminActive() bool { return u.IsAdmin && u.IsActive }
func (u *User) IsStaffActive() bool { return !u.IsAdmin && u.IsActive }
func (u *User) IsActiveUser() bool  { return u.IsActive }
func (u *User) IsAdminUser() bool   { return u.IsAdmin }
