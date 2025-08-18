// package service

// import (
// 	"Contact_App/apperror"
// 	"Contact_App/db"
// 	"Contact_App/models/contact"
// 	"Contact_App/repository"
// 	"strings"
// )

// type UpdateFieldInput struct {
// 	Field string `json:"field"`
// 	Value string `json:"value"`
// }

// var contactRepo = repository.NewGormRepository()

// func CreateContact(userID int, fname, lname string) (*contact.Contact, error) {
// 	if strings.TrimSpace(fname) == "" {
// 		return nil, apperror.NewValidationError("fname", "first name cannot be empty")
// 	}
// 	if strings.TrimSpace(lname) == "" {
// 		return nil, apperror.NewValidationError("lname", "last name cannot be empty")
// 	}

// 	newContact := &contact.Contact{
// 		UserID:   userID,
// 		FName:    strings.TrimSpace(fname),
// 		LName:    strings.TrimSpace(lname),
// 		IsActive: true,
// 	}

// 	uow := repository.NewUnitOfWork(db.GetDB(), false)
// 	defer uow.Rollback()

// 	if err := contactRepo.Add(uow, newContact); err != nil {
// 		return nil, err
// 	}

// 	uow.Commit()
// 	return newContact, nil
// }

// func GetContacts(userID int) ([]*contact.Contact, error) {
// 	var contacts []*contact.Contact
// 	uow := repository.NewUnitOfWork(db.GetDB(), true)

// 	err := contactRepo.GetAll(uow, &contacts, repository.Filter("user_id = ?", userID))
// 	if err != nil {
// 		return nil, err
// 	}

// 	return contacts, nil
// }

// func GetContactByID(userID, contactID int) (*contact.Contact, error) {
// 	var contacts []*contact.Contact
// 	uow := repository.NewUnitOfWork(db.GetDB(), true)

// 	err := contactRepo.GetAll(uow, &contacts,
// 		repository.Filter("contact_id = ? AND user_id = ?", contactID, userID),
// 	)

// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(contacts) == 0 || contacts[0].ContactID == 0 {
// 		return nil, apperror.NewNotFoundError("contact", contactID)
// 	}

// 	return contacts[0], nil
// }

// func UpdateContactByID(userID, contactID int, field string, value interface{}) error {

// 	if _, err := GetContactByID(userID, contactID); err != nil {
// 		return err
// 	}

// 	updates := make(map[string]interface{})
// 	switch strings.ToLower(field) {
// 	case "fname", "firstname":
// 		strVal, ok := value.(string)
// 		if !ok || strings.TrimSpace(strVal) == "" {
// 			return apperror.NewValidationError("fname", "must be a non-empty string")
// 		}
// 		updates["f_name"] = strings.TrimSpace(strVal)

// 	case "lname", "lastname":
// 		strVal, ok := value.(string)
// 		if !ok || strings.TrimSpace(strVal) == "" {
// 			return apperror.NewValidationError("lname", "must be a non-empty string")
// 		}
// 		updates["l_name"] = strings.TrimSpace(strVal)

// 	default:
// 		return apperror.NewValidationError("field", "unknown contact field")
// 	}

// 	uow := repository.NewUnitOfWork(db.GetDB(), false)
// 	defer uow.Rollback()

// 	err := contactRepo.UpdateWithMap(uow, &contact.Contact{}, updates,
// 		repository.Filter("contact_id = ? AND user_id = ?", contactID, userID),
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	uow.Commit()
// 	return nil
// }

// func DeleteContactByID(userID, contactID int) error {
// 	uow := repository.NewUnitOfWork(db.GetDB(), false)
// 	defer uow.Rollback()

// 	var contacts []*contact.Contact
// 	err := contactRepo.GetAll(uow, &contacts,
// 		repository.Filter("contact_id = ? AND user_id = ?", contactID, userID),
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	if len(contacts) == 0 || contacts[0].ContactID == 0 {
// 		return apperror.NewNotFoundError("contact", contactID)
// 	}

// 	err = contactRepo.Delete(uow, &contact.Contact{},
// 		repository.Filter("contact_id = ? AND user_id = ?", contactID, userID),
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	uow.Commit()
// 	return nil
// }

package service

import (
	"Contact_App/apperror"
	"Contact_App/db"
	"Contact_App/models/contact"
	"Contact_App/repository"
	"strings"
)

type UpdateFieldInput struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

var contactRepo = repository.NewGormRepository()

func CreateContact(userID int, fname, lname string) (*contact.Contact, error) {
	if strings.TrimSpace(fname) == "" {
		return nil, apperror.NewValidationError("fname", "first name cannot be empty")
	}
	if strings.TrimSpace(lname) == "" {
		return nil, apperror.NewValidationError("lname", "last name cannot be empty")
	}

	newContact := &contact.Contact{
		UserID:   userID,
		FName:    strings.TrimSpace(fname),
		LName:    strings.TrimSpace(lname),
		IsActive: true,
	}

	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	if err := contactRepo.Add(uow, newContact); err != nil {
		return nil, err
	}

	uow.Commit()
	return newContact, nil
}

func GetContacts(userID int, filters ...map[string]string) ([]*contact.Contact, error) {
	var contacts []*contact.Contact
	uow := repository.NewUnitOfWork(db.GetDB(), true)

	// Always filter by userID
	processors := []repository.QueryProcessor{
		repository.Filter("user_id = ?", userID),
	}

	// Optional filters (if provided via handler)
	if len(filters) > 0 {
		f := filters[0]
		if name, ok := f["name"]; ok && strings.TrimSpace(name) != "" {
			processors = append(processors, repository.Filter("f_name LIKE ? OR l_name LIKE ?", "%"+name+"%", "%"+name+"%"))
		}
		if email, ok := f["email"]; ok && strings.TrimSpace(email) != "" {
			processors = append(processors, repository.Filter("email = ?", email))
		}
		if phone, ok := f["phone"]; ok && strings.TrimSpace(phone) != "" {
			processors = append(processors, repository.Filter("phone = ?", phone))
		}
	}

	err := contactRepo.GetAll(uow, &contacts, processors...)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

func GetContactByID(userID, contactID int) (*contact.Contact, error) {
	var contacts []*contact.Contact
	uow := repository.NewUnitOfWork(db.GetDB(), true)

	err := contactRepo.GetAll(uow, &contacts,
		repository.Filter("contact_id = ? AND user_id = ?", contactID, userID),
	)

	if err != nil {
		return nil, err
	}
	if len(contacts) == 0 || contacts[0].ContactID == 0 {
		return nil, apperror.NewNotFoundError("contact", contactID)
	}

	return contacts[0], nil
}

func UpdateContactByID(userID, contactID int, field string, value interface{}) error {

	if _, err := GetContactByID(userID, contactID); err != nil {
		return err
	}

	updates := make(map[string]interface{})
	switch strings.ToLower(field) {
	case "fname", "firstname":
		strVal, ok := value.(string)
		if !ok || strings.TrimSpace(strVal) == "" {
			return apperror.NewValidationError("fname", "must be a non-empty string")
		}
		updates["f_name"] = strings.TrimSpace(strVal)

	case "lname", "lastname":
		strVal, ok := value.(string)
		if !ok || strings.TrimSpace(strVal) == "" {
			return apperror.NewValidationError("lname", "must be a non-empty string")
		}
		updates["l_name"] = strings.TrimSpace(strVal)

	default:
		return apperror.NewValidationError("field", "unknown contact field")
	}

	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	err := contactRepo.UpdateWithMap(uow, &contact.Contact{}, updates,
		repository.Filter("contact_id = ? AND user_id = ?", contactID, userID),
	)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func DeleteContactByID(userID, contactID int) error {
	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	var contacts []*contact.Contact
	err := contactRepo.GetAll(uow, &contacts,
		repository.Filter("contact_id = ? AND user_id = ?", contactID, userID),
	)
	if err != nil {
		return err
	}
	if len(contacts) == 0 || contacts[0].ContactID == 0 {
		return apperror.NewNotFoundError("contact", contactID)
	}

	err = contactRepo.Delete(uow, &contact.Contact{},
		repository.Filter("contact_id = ? AND user_id = ?", contactID, userID),
	)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func GetContactsFiltered(userID int, fName, lName, phone string, contactRepo repository.Repository, db *repository.UnitOfWork) ([]*contact.Contact, error) {
	var contacts []*contact.Contact
	uow := repository.NewUnitOfWork(db.DB, true) // readonly

	// Start with base filter (always by userID)
	filters := []repository.QueryProcessor{
		repository.Filter("user_id = ?", userID),
	}

	// Add filters only if query params are present
	if strings.TrimSpace(fName) != "" {
		filters = append(filters, repository.Filter("f_name LIKE ?", "%"+fName+"%"))
	}
	if strings.TrimSpace(lName) != "" {
		filters = append(filters, repository.Filter("l_name LIKE ?", "%"+lName+"%"))
	}
	if strings.TrimSpace(phone) != "" {
		filters = append(filters, repository.Filter("phone LIKE ?", "%"+phone+"%"))
	}

	// Fetch filtered contacts
	err := contactRepo.GetAll(uow, &contacts, filters...)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}
