package order

import (
	"context"
	"errors"
	"strconv"

	"mini-e-commerce/internal/product"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
	db   *gorm.DB
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

func (s *Service) Create(ctx context.Context, p *Order) error {
	return s.repo.Create(ctx, p)
}

func (s *Service) GetAll(ctx context.Context) ([]Order, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id uint) (Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, p *Order) error {
	return s.repo.Update(ctx, p)
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) CreateOrder(ctx context.Context, input CreateOrderInput, userID uint) (*Order, error) {
	var prod product.Product
	if err := s.db.First(&prod, input.ProductID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	if input.Quantity > prod.Stock {
		return nil, errors.New("insufficient stock")
	}

	order := Order{
		UserID:     userID,
		ProductID:  input.ProductID,
		Quantity:   input.Quantity,
		TotalPrice: input.Quantity * prod.Price,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		prod.Stock -= input.Quantity
		if err := tx.Save(&prod).Error; err != nil {
			return err
		}
		
		return tx.Create(&order).Error
	})

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *Service) UpdateOrder(ctx context.Context, id uint, input UpdateOrderInput, userID uint) (*Order, error) {
	var order Order
	if err := s.db.First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	if order.UserID != userID {
		return nil, errors.New("not authorized to update this order")
	}

	if input.Status != nil {
		switch *input.Status {
		case StatusPending, StatusPaid, StatusCancelled:
		default:
			return nil, errors.New("invalid status value")
		}
	}

	if input.Quantity != nil {
		var prod product.Product
		if err := s.db.First(&prod, order.ProductID).Error; err != nil {
			return nil, err
		}

		diff := *input.Quantity - order.Quantity
		if diff > 0 && diff > prod.Stock {
			return nil, errors.New("insufficient stock for quantity update")
		}

		err := s.db.Transaction(func(tx *gorm.DB) error {
				prod.Stock -= diff
			if err := tx.Save(&prod).Error; err != nil {
				return err
			}

			order.Quantity = *input.Quantity
			order.TotalPrice = *input.Quantity * prod.Price
			if input.Status != nil {
				order.Status = *input.Status
			}
			return tx.Save(&order).Error
		})

		if err != nil {
			return nil, err
		}
	} else if input.Status != nil {
		order.Status = *input.Status
		if err := s.db.Save(&order).Error; err != nil {
			return nil, err
		}
	}

	return &order, nil
}

func (s *Service) DeleteOrder(ctx context.Context, id uint) error {
	var order Order
	if err := s.db.First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("order not found")
		}
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var prod product.Product
		if err := tx.First(&prod, order.ProductID).Error; err != nil {
			return err
		}
		prod.Stock += order.Quantity
		if err := tx.Save(&prod).Error; err != nil {
			return err
		}

		return tx.Delete(&order).Error
	})
}

func (s *Service) ParseIDFromString(idStr string) (uint, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (s *Service) ParseUserIDFromString(userIDStr string) (uint, error) {
	uid, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(uid), nil
}
