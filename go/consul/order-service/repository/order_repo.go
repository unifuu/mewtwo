package repository

import (
    "gorm.io/gorm"
    "order-service/model"
)

type OrderRepository interface {
    Create(order *model.Order) error
    FindByID(id uint) (*model.Order, error)
    FindByUserID(userID uint) ([]model.Order, error)
}

type orderRepository struct {
    db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
    return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order) error {
    return r.db.Create(order).Error
}

func (r *orderRepository) FindByID(id uint) (*model.Order, error) {
    var order model.Order
    err := r.db.First(&order, id).Error
    return &order, err
}

func (r *orderRepository) FindByUserID(userID uint) ([]model.Order, error) {
    var orders []model.Order
    err := r.db.Where("user_id = ?", userID).Find(&orders).Error
    return orders, err
}