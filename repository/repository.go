package repository

import (
	"Contact_App/apperror"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	GetAll(uow *UnitOfWork, out interface{}, queryProcessors ...QueryProcessor) error
	Add(uow *UnitOfWork, model interface{}) error
	Update(uow *UnitOfWork, model interface{}) error
	UpdateWithMap(uow *UnitOfWork, model interface{}, value map[string]interface{}, queryProcessors ...QueryProcessor) error
	GetByID(uow *UnitOfWork, id uint, out interface{}, queryProcessors ...QueryProcessor) error
	Save(uow *UnitOfWork, value interface{}) error
	Delete(uow *UnitOfWork, model interface{}, queryProcessors ...QueryProcessor) error
}

type GormRepository struct{}

func NewGormRepository() *GormRepository {
	return &GormRepository{}
}

type UnitOfWork struct {
	DB        *gorm.DB
	Committed bool
	Readonly  bool
}

func NewUnitOfWork(db *gorm.DB, readonly bool) *UnitOfWork {
	if readonly {
		return &UnitOfWork{
			DB:        db.Session(&gorm.Session{}),
			Committed: false,
			Readonly:  true,
		}
	}
	return &UnitOfWork{
		DB:        db.Begin(),
		Committed: false,
		Readonly:  false,
	}
}

func (uow *UnitOfWork) Commit() {
	if !uow.Readonly && !uow.Committed {
		uow.Committed = true
		uow.DB.Commit()
	}
}

func (uow *UnitOfWork) Rollback() {
	if !uow.Committed && !uow.Readonly {
		uow.DB.Rollback()
	}
}

func (r *GormRepository) Save(uow *UnitOfWork, value interface{}) error {
	if err := uow.DB.Save(value).Error; err != nil {
		return apperror.NewInternalError("Failed to save data: " + err.Error())
	}
	return nil
}

func (r *GormRepository) Add(uow *UnitOfWork, model interface{}) error {
	if err := uow.DB.Create(model).Error; err != nil {
		return apperror.NewInternalError("Failed to add record: " + err.Error())
	}
	return nil
}

func (r *GormRepository) Update(uow *UnitOfWork, model interface{}) error {
	if err := uow.DB.Save(model).Error; err != nil {
		return apperror.NewInternalError("Failed to update record: " + err.Error())
	}
	return nil
}

func (r *GormRepository) UpdateWithMap(uow *UnitOfWork, model interface{}, value map[string]interface{}, queryProcessors ...QueryProcessor) error {
	db := uow.DB.Model(model)
	var err error
	db, err = applyQueryProcessors(db, model, queryProcessors...)
	if err != nil {
		return err
	}
	if err := db.Updates(value).Error; err != nil {
		return apperror.NewInternalError("Failed to update with map: " + err.Error())
	}
	return nil
}

type QueryProcessor func(db *gorm.DB, out interface{}) (*gorm.DB, error)

func Filter(condition string, args ...interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		return db.Where(condition, args...), nil
	}
}

func Select(fields string) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		return db.Select(fields), nil
	}
}

func applyQueryProcessors(db *gorm.DB, out interface{}, processors ...QueryProcessor) (*gorm.DB, error) {
	var err error
	for _, process := range processors {
		if process != nil {
			db, err = process(db, out)
			if err != nil {
				return db, apperror.NewValidationError("query", err.Error())
			}
		}
	}
	return db, nil
}

//	func (r *GormRepository) Delete(uow *UnitOfWork, model interface{}, queryProcessors ...QueryProcessor) error {
//		db := uow.DB.Model(model)
//		var err error
//		db, err = applyQueryProcessors(db, model, queryProcessors...)
//		if err != nil {
//			return err
//		}
//		if err := db.Update("is_active", false).Error; err != nil {
//			return apperror.NewInternalError(fmt.Sprintf("soft delete failed: %v", err))
//		}
//		return nil
//	}
func (r *GormRepository) Delete(uow *UnitOfWork, model interface{}, queryProcessors ...QueryProcessor) error {
	db := uow.DB.Model(model)
	var err error
	db, err = applyQueryProcessors(db, model, queryProcessors...)
	if err != nil {
		return err
	}
	if err := db.Delete(model).Error; err != nil {
		return apperror.NewInternalError(fmt.Sprintf("soft delete failed: %v", err))
	}
	return nil
}

func (r *GormRepository) GetAll(uow *UnitOfWork, out interface{}, queryProcessors ...QueryProcessor) error {
	db := uow.DB.Model(out).Where("is_active = ?", true)
	var err error
	db, err = applyQueryProcessors(db, out, queryProcessors...)
	if err != nil {
		return err
	}
	if err := db.Find(out).Error; err != nil {
		return apperror.NewInternalError(fmt.Sprintf("failed to fetch records: %v", err))
	}
	return nil
}

func (r *GormRepository) GetByID(uow *UnitOfWork, id uint, out interface{}, queryProcessors ...QueryProcessor) error {
	db := uow.DB.Model(out).Where("user_id = ? AND is_active = ?", id, true) // FIXED
	var err error
	db, err = applyQueryProcessors(db, out, queryProcessors...)
	if err != nil {
		return err
	}
	if err := db.First(out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFoundError("record", int(id))
		}
		return apperror.NewInternalError(fmt.Sprintf("failed to fetch record: %v", err))
	}
	return nil
}
