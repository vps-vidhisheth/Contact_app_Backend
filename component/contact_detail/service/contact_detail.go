// package service

// import (
// 	"Contact_App/apperror"
// 	"Contact_App/models/contact_detail"
// 	"Contact_App/repository"
// 	"strings"

// 	"gorm.io/gorm"
// )

// var contactDetailRepo = repository.NewGormRepository()

// type UpdateDetailInput struct {
// 	Email string `json:"email,omitempty"`
// 	Phone string `json:"phone,omitempty"`
// }

// func AddDetailToContact(db *gorm.DB, userID, contactID int, detailType, value string) (*contact_detail.ContactDetail, error) {
// 	if strings.TrimSpace(detailType) == "" || strings.TrimSpace(value) == "" {
// 		return nil, apperror.NewValidationError("type/value", "type and value cannot be empty")
// 	}

// 	detail := &contact_detail.ContactDetail{
// 		UserID:    userID,
// 		ContactID: contactID,
// 		Type:      strings.ToLower(detailType),
// 		Value:     value,
// 	}

// 	if err := db.Create(detail).Error; err != nil {
// 		return nil, apperror.NewInternalError("failed to save detail to database")
// 	}

// 	return detail, nil
// }

// func UpdateDetailByID(db *gorm.DB, userID, contactID, detailID int, input UpdateDetailInput) error {
// 	uow := repository.NewUnitOfWork(db, false)
// 	defer uow.Rollback()

// 	var details []*contact_detail.ContactDetail
// 	err := contactDetailRepo.GetAll(uow, &details,
// 		repository.Filter("contact_details_id = ? AND contact_id = ?", detailID, contactID),
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	if len(details) == 0 {
// 		return apperror.NewNotFoundError("contact_detail", detailID)
// 	}

// 	detail := details[0]

// 	if strings.TrimSpace(input.Email) != "" {
// 		detail.Type = "email"
// 		detail.Value = strings.TrimSpace(input.Email)
// 	} else if strings.TrimSpace(input.Phone) != "" {
// 		detail.Type = "phone"
// 		detail.Value = strings.TrimSpace(input.Phone)
// 	} else {
// 		return apperror.NewValidationError("body", "must contain either 'email' or 'phone'")
// 	}

// 	if detail.Value == "" {
// 		return apperror.NewValidationError("value", "cannot be empty")
// 	}

// 	err = contactDetailRepo.Update(uow, detail)
// 	if err != nil {
// 		return err
// 	}

// 	uow.Commit()
// 	return nil
// }

// func DeleteDetailByID(db *gorm.DB, userID, contactID, detailID int) error {
// 	uow := repository.NewUnitOfWork(db, false)
// 	defer uow.Rollback()

// 	var details []*contact_detail.ContactDetail
// 	err := contactDetailRepo.GetAll(uow, &details,
// 		repository.Filter("contact_details_id = ? AND contact_id = ?", detailID, contactID),
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	if len(details) == 0 {
// 		return apperror.NewNotFoundError("contact_detail", detailID)
// 	}

// 	err = uow.DB.Delete(details[0]).Error
// 	if err != nil {
// 		return apperror.NewInternalError("failed to delete contact detail")
// 	}

// 	uow.Commit()
// 	return nil
// }

package service

import (
	"Contact_App/apperror"
	"Contact_App/models/contact_detail"
	"Contact_App/repository"
	"strings"

	"gorm.io/gorm"
)

var contactDetailRepo = repository.NewGormRepository()

type UpdateDetailInput struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// Add a new contact detail
func AddDetailToContact(db *gorm.DB, userID, contactID int, detailType, value string) (*contact_detail.ContactDetail, error) {
	if strings.TrimSpace(detailType) == "" || strings.TrimSpace(value) == "" {
		return nil, apperror.NewValidationError("type/value", "type and value cannot be empty")
	}

	detail := &contact_detail.ContactDetail{
		UserID:    userID,
		ContactID: contactID,
		Type:      strings.ToLower(detailType),
		Value:     value,
	}

	if err := db.Create(detail).Error; err != nil {
		return nil, apperror.NewInternalError("failed to save detail to database")
	}

	return detail, nil
}

// Update a contact detail by ID
func UpdateDetailByID(db *gorm.DB, userID, contactID, detailID int, input UpdateDetailInput) error {
	uow := repository.NewUnitOfWork(db, false)
	defer uow.Rollback()

	var details []*contact_detail.ContactDetail
	err := contactDetailRepo.GetAll(uow, &details,
		repository.Filter("contact_details_id = ? AND contact_id = ? AND user_id = ?", detailID, contactID, userID),
	)
	if err != nil {
		return err
	}
	if len(details) == 0 {
		return apperror.NewNotFoundError("contact_detail", detailID)
	}

	updates := make(map[string]interface{})

	if strings.TrimSpace(input.Email) != "" {
		updates["type"] = "email"
		updates["value"] = strings.TrimSpace(input.Email)
	} else if strings.TrimSpace(input.Phone) != "" {
		updates["type"] = "phone"
		updates["value"] = strings.TrimSpace(input.Phone)
	} else {
		return apperror.NewValidationError("body", "must contain either 'email' or 'phone'")
	}

	if updates["value"] == "" {
		return apperror.NewValidationError("value", "cannot be empty")
	}

	err = contactDetailRepo.UpdateWithMap(uow, &contact_detail.ContactDetail{}, updates,
		repository.Filter("contact_details_id = ? AND contact_id = ? AND user_id = ?", detailID, contactID, userID),
	)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

// Delete a contact detail by ID
func DeleteDetailByID(db *gorm.DB, userID, contactID, detailID int) error {
	uow := repository.NewUnitOfWork(db, false)
	defer uow.Rollback()

	var details []*contact_detail.ContactDetail
	err := contactDetailRepo.GetAll(uow, &details,
		repository.Filter("contact_detail_id = ? AND contact_id = ? AND user_id = ?", detailID, contactID, userID),
	)
	if err != nil {
		return err
	}
	if len(details) == 0 {
		return apperror.NewNotFoundError("contact_detail", detailID)
	}

	err = contactDetailRepo.Delete(uow, details[0])
	if err != nil {
		return apperror.NewInternalError("failed to delete contact detail")
	}

	uow.Commit()
	return nil
}

// Get all contact details with optional filters
func GetContactDetails(db *gorm.DB, userID, contactID int, detailType, value string) ([]*contact_detail.ContactDetail, error) {
	uow := repository.NewUnitOfWork(db, true)
	defer uow.Rollback()

	filters := []repository.QueryProcessor{
		repository.Filter("user_id = ?", userID),
		repository.Filter("contact_id = ?", contactID),
	}

	if strings.TrimSpace(detailType) != "" {
		filters = append(filters, repository.Filter("type LIKE ?", "%"+detailType+"%"))
	}
	if strings.TrimSpace(value) != "" {
		filters = append(filters, repository.Filter("value LIKE ?", "%"+value+"%"))
	}

	var details []*contact_detail.ContactDetail
	err := contactDetailRepo.GetAll(uow, &details, filters...)
	if err != nil {
		return nil, err
	}

	return details, nil
}

// Get a single contact detail by ID
func GetContactDetailByID(db *gorm.DB, userID, contactID, detailID int) (*contact_detail.ContactDetail, error) {
	uow := repository.NewUnitOfWork(db, true)
	defer uow.Rollback()

	var details []*contact_detail.ContactDetail
	err := contactDetailRepo.GetAll(uow, &details,
		repository.Filter("contact_detail_id = ? AND contact_id = ? AND user_id = ?", detailID, contactID, userID),
	)
	if err != nil {
		return nil, err
	}
	if len(details) == 0 {
		return nil, apperror.NewNotFoundError("contact_detail", detailID)
	}

	return details[0], nil
}
