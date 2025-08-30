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

type ContactService struct {
	contactRepo       repository.Repository
	contactDetailRepo repository.Repository
}

func NewContactService() *ContactService {
	return &ContactService{
		contactRepo:       repository.NewGormRepository(),
		contactDetailRepo: repository.NewGormRepository(),
	}
}

// CreateContact creates a new contact for a user
func (s *ContactService) CreateContact(userID uint, fname, lname string) (*contact.Contact, error) {
	fname = strings.TrimSpace(fname)
	lname = strings.TrimSpace(lname)
	if fname == "" {
		return nil, apperror.NewValidationError("first_name", "cannot be empty")
	}
	if lname == "" {
		return nil, apperror.NewValidationError("last_name", "cannot be empty")
	}

	newContact := &contact.Contact{
		UserID:   userID,
		FName:    fname,
		LName:    lname,
		IsActive: true,
	}

	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	if err := s.contactRepo.Add(uow, newContact); err != nil {
		return nil, err
	}

	uow.Commit()
	return newContact, nil
}

// GetContactsWithDetails retrieves all contacts belonging to the given user
func (s *ContactService) GetContactsWithDetails(userID uint, filters map[string]string) ([]*contact.Contact, error) {
	var contacts []*contact.Contact
	uow := repository.NewUnitOfWork(db.GetDB(), true)

	query := uow.DB.Preload("Details").Where("user_id = ? AND is_active = ?", userID, true)

	if name := strings.TrimSpace(filters["f_name"]); name != "" {
		query = query.Where("f_name LIKE ? OR l_name LIKE ?", "%"+name+"%", "%"+name+"%")
	}
	if phone := strings.TrimSpace(filters["phone"]); phone != "" {
		query = query.Joins("JOIN contact_details ON contacts.contact_id = contact_details.contact_id").
			Where("contacts.user_id = ? AND contact_details.type = 'phone' AND contact_details.value LIKE ?", userID, "%"+phone+"%")
	}

	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

// GetContactByIDWithDetails retrieves a single contact by ID ensuring ownership
func (s *ContactService) GetContactByIDWithDetails(userID, contactID uint) (*contact.Contact, error) {
	var c contact.Contact
	uow := repository.NewUnitOfWork(db.GetDB(), true)

	err := uow.DB.Preload("Details").
		Where("contact_id = ? AND user_id = ? AND is_active = ?", contactID, userID, true).
		First(&c).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.NewNotFoundError("contact", int(contactID))
		}
		return nil, err
	}
	return &c, nil
}

// AddOrUpdateContactDetail adds or updates a contact detail using the provided transaction
func (s *ContactService) AddOrUpdateContactDetail(uow *repository.UnitOfWork, userID, contactID uint, detailType, value string) error {
	detailType = strings.ToLower(strings.TrimSpace(detailType))
	value = strings.TrimSpace(value)
	if detailType == "" || value == "" {
		return apperror.NewValidationError("detail", "type and value cannot be empty")
	}

	var detail contact_detail.ContactDetail
	err := uow.DB.Where("contact_id = ? AND type = ?", contactID, detailType).First(&detail).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == gorm.ErrRecordNotFound {
		newDetail := &contact_detail.ContactDetail{
			ContactID: contactID,
			UserID:    userID,
			Type:      detailType,
			Value:     value,
			IsActive:  true,
		}
		if err := s.contactDetailRepo.Add(uow, newDetail); err != nil {
			return err
		}
	} else {
		detail.Value = value
		if err := s.contactDetailRepo.Update(uow, &detail); err != nil {
			return err
		}
	}

	return nil
}

func (s *ContactService) UpdateContactByID(userID, contactID uint, updates map[string]interface{}) error {
	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	c, err := s.GetContactByIDWithDetails(userID, contactID)
	if err != nil {
		return err
	}

	updateMap := make(map[string]interface{})

	if v, ok := updates["first_name"]; ok {
		strVal, _ := v.(string)
		strVal = strings.TrimSpace(strVal)
		if strVal == "" {
			return apperror.NewValidationError("first_name", "cannot be empty")
		}
		updateMap["f_name"] = strVal
	}

	if v, ok := updates["last_name"]; ok {
		strVal, _ := v.(string)
		strVal = strings.TrimSpace(strVal)
		if strVal == "" {
			return apperror.NewValidationError("last_name", "cannot be empty")
		}
		updateMap["l_name"] = strVal
	}

	if v, ok := updates["is_active"]; ok {
		switch val := v.(type) {
		case bool:
			updateMap["is_active"] = val
		case string:
			lower := strings.ToLower(strings.TrimSpace(val))
			if lower == "true" {
				updateMap["is_active"] = true
			} else if lower == "false" {
				updateMap["is_active"] = false
			} else {
				return apperror.NewValidationError("is_active", "must be a boolean")
			}
		default:
			return apperror.NewValidationError("is_active", "must be a boolean")
		}
	}

	if len(updateMap) > 0 {
		if err := s.contactRepo.UpdateWithMap(uow, c, updateMap, repository.Filter("contact_id = ? AND user_id = ?", contactID, userID)); err != nil {
			return err
		}
	}

	if v, ok := updates["details"]; ok {
		if details, ok2 := v.([]interface{}); ok2 {
			for _, d := range details {
				if detailMap, ok3 := d.(map[string]interface{}); ok3 {
					dType, _ := detailMap["type"].(string)
					dVal, _ := detailMap["value"].(string)
					dType = strings.TrimSpace(dType)
					dVal = strings.TrimSpace(dVal)
					if dType != "" && dVal != "" {
						if err := s.AddOrUpdateContactDetail(uow, userID, contactID, dType, dVal); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	uow.Commit()
	return nil
}

func (s *ContactService) DeleteContactByID(userID, contactID uint) error {
	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	if err := s.contactRepo.Delete(uow, &contact.Contact{}, repository.Filter("contact_id = ? AND user_id = ?", contactID, userID)); err != nil {
		return err
	}

	if err := s.contactDetailRepo.UpdateWithMap(
		uow,
		&contact_detail.ContactDetail{},
		map[string]interface{}{"is_active": false},
		repository.Filter("contact_id = ? AND user_id = ?", contactID, userID),
	); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *ContactService) CreateContactWithUOW(uow *repository.UnitOfWork, userID uint, fname, lname string) (*contact.Contact, error) {
	newContact := &contact.Contact{
		UserID:   userID,
		FName:    strings.TrimSpace(fname),
		LName:    strings.TrimSpace(lname),
		IsActive: true,
	}
	if err := s.contactRepo.Add(uow, newContact); err != nil {
		return nil, err
	}
	return newContact, nil
}
