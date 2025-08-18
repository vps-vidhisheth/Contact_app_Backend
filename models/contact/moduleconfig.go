package contact

import (
	"Contact_App/apperror"
	"errors"
	"strings"
)

var ErrContactDetailNotFound = errors.New("contact detail not found")

type User interface {
	IsAdminUser() bool
	IsActiveUser() bool
	GetContactByID(id int) (*Contact, error)
}

type Requester interface {
	IsAdminUser() bool
	IsActiveUser() bool
}

func NewContact(fname, lname string, userID int) (*Contact, error) {
	fname, lname = strings.TrimSpace(fname), strings.TrimSpace(lname)

	if fname == "" {
		return nil, apperror.NewValidationError("f_name", "first name cannot be empty")
	}

	return &Contact{
		FName:    fname,
		LName:    lname,
		UserID:   userID,
		IsActive: true,
	}, nil
}

func (c *Contact) UpdateField(field string, value interface{}) error {
	switch strings.ToLower(field) {
	case "f_name":
		v, ok := value.(string)
		v = strings.TrimSpace(v)
		if !ok || v == "" {
			return apperror.NewValidationError("f_name", "must be a non-empty string")
		}
		c.FName = v
	case "l_name":
		v, ok := value.(string)
		v = strings.TrimSpace(v)
		if !ok {
			return apperror.NewValidationError("l_name", "must be a string")
		}
		c.LName = v
	default:
		return apperror.NewValidationError("field", "unknown field for contact")
	}
	return nil
}
