package service

import (
	"Contact_App/apperror"
	"Contact_App/db"
	"Contact_App/models/contact"
	"Contact_App/models/contact_detail"
	"Contact_App/repository"
	"strings"

	"gorm.io/gorm"
)

type UpdateFieldInput struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

var contactRepo = repository.NewGormRepository()
var contactDetailRepo = repository.NewGormRepository()

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

	query := uow.DB.Preload("Details").Where("user_id = ?", userID)

	if len(filters) > 0 {
		f := filters[0]
		if name, ok := f["name"]; ok && strings.TrimSpace(name) != "" {
			query = query.Where("f_name LIKE ? OR l_name LIKE ?", "%"+name+"%", "%"+name+"%")
		}
		if email, ok := f["email"]; ok && strings.TrimSpace(email) != "" {
			query = query.Joins("JOIN contact_details ON contacts.contact_id = contact_details.contact_id").
				Where("contact_details.type = 'email' AND contact_details.value LIKE ?", "%"+email+"%")
		}
		if phone, ok := f["phone"]; ok && strings.TrimSpace(phone) != "" {
			query = query.Joins("JOIN contact_details ON contacts.contact_id = contact_details.contact_id").
				Where("contact_details.type = 'phone' AND contact_details.value LIKE ?", "%"+phone+"%")
		}
	}

	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}

	return contacts, nil
}

func GetContactByID(userID, contactID int) (*contact.Contact, error) {
	var contact contact.Contact
	uow := repository.NewUnitOfWork(db.GetDB(), true)

	err := uow.DB.Preload("Details").
		Where("contact_id = ? AND user_id = ?", contactID, userID).
		First(&contact).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.NewNotFoundError("contact", contactID)
		}
		return nil, err
	}

	return &contact, nil

}

func AddOrUpdateContactDetail(contactID int, detailType, value string) error {
	if strings.TrimSpace(detailType) == "" || strings.TrimSpace(value) == "" {
		return apperror.NewValidationError("detail", "type and value cannot be empty")
	}

	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	var existingDetail contact_detail.ContactDetail
	err := uow.DB.Where("contact_id = ? AND type = ?", contactID, detailType).First(&existingDetail).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new detail
			newDetail := &contact_detail.ContactDetail{
				ContactID: contactID,
				Type:      strings.TrimSpace(detailType),
				Value:     strings.TrimSpace(value),
			}
			if err := contactDetailRepo.Add(uow, newDetail); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Update existing detail
		updates := map[string]interface{}{
			"value": strings.TrimSpace(value),
		}
		if err := contactDetailRepo.UpdateWithMap(uow, &contact_detail.ContactDetail{}, updates,
			repository.Filter("contact_details_id = ? AND contact_id = ?", existingDetail.ContactDetailsID, contactID),
		); err != nil {
			return err
		}
	}

	uow.Commit()
	return nil
}

// Update contact fields
func UpdateContactByID(userID, contactID int, field string, value interface{}) error {
	contactObj, err := GetContactByID(userID, contactID)
	if err != nil {
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

	case "is_active":
		boolVal, ok := value.(bool)
		if !ok {
			return apperror.NewValidationError("is_active", "must be a boolean")
		}
		updates["is_active"] = boolVal

	default:
		return apperror.NewValidationError("field", "unknown contact field")
	}

	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	err = contactRepo.UpdateWithMap(uow, contactObj, updates,
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

func GetContactsWithDetails(userID int, filters map[string]string) ([]*contact.Contact, error) {
	var contacts []*contact.Contact
	uow := repository.NewUnitOfWork(db.GetDB(), true)

	query := uow.DB.Preload("Details").Where("user_id = ?", userID)

	if fName := strings.TrimSpace(filters["f_name"]); fName != "" {
		query = query.Where("f_name LIKE ?", "%"+fName+"%")
	}
	if lName := strings.TrimSpace(filters["l_name"]); lName != "" {
		query = query.Where("l_name LIKE ?", "%"+lName+"%")
	}
	if phone := strings.TrimSpace(filters["phone"]); phone != "" {
		query = query.Joins("JOIN contact_details ON contacts.contact_id = contact_details.contact_id").
			Where("contact_details.value LIKE ?", "%"+phone+"%")
	}

	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}

	return contacts, nil
}

func GetContactByIDWithDetails(userID, contactID int) (*contact.Contact, error) {
	var contact contact.Contact
	uow := repository.NewUnitOfWork(db.GetDB(), true) // readonly

	err := uow.DB.Preload("Details").
		Where("contact_id = ? AND user_id = ?", contactID, userID).
		First(&contact).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.NewNotFoundError("contact", contactID)
		}
		return nil, err
	}

	return &contact, nil
}
