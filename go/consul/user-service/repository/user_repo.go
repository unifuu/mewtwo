package repository

import (
    "gorm.io/gorm"
    "user-service/model"
)

type UserRepository interface {
    Create(user *model.User) error
    FindByID(id uint) (*model.User, error)
    FindAll() ([]model.User, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
    return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
    var user model.User
    err := r.db.First(&user, id).Error
    return &user, err
}

func (r *userRepository) FindAll() ([]model.User, error) {
    var users []model.User
    err := r.db.Find(&users).Error
    return users, err
}