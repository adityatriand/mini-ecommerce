package order

import (
	"context"
	"errors"
	"strconv"

	"mini-e-commerce/internal/product"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

const (
	ErrOrderNotFound                    = "order not found"
	ErrProductNotFound                  = "product not found"
	ErrInsufficientStock                = "insufficient stock"
	ErrNotAuthorizedToUpdate            = "not authorized to update this order"
	ErrInvalidStatusValue               = "invalid status value"
	ErrCannotUpdateBothQuantityAndStatus = "cannot update both quantity and status simultaneously"
	ErrCannotChangePaidOrderToPending   = "cannot change paid order back to pending"
	ErrCannotChangeCancelledOrderStatus = "cannot change cancelled order status"
	ErrInsufficientStockForUpdate       = "insufficient stock for quantity update"
)

type Service struct {
	repo        *Repository
	productRepo *product.Repository
	validator   *validator.Validate
}

func NewService(repo *Repository, productRepo *product.Repository) *Service {
	return &Service{
		repo:        repo,
		productRepo: productRepo,
		validator:   validator.New(),
	}
}

func (s *Service) CreateOrder(ctx context.Context, input CreateOrderRequest, userID uint) (*Order, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	product, err := s.productRepo.FindByID(ctx, input.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrProductNotFound)
		}
		return nil, err
	}

	if input.Quantity > product.Stock {
		return nil, errors.New(ErrInsufficientStock)
	}

	order := Order{
		UserID:     userID,
		ProductID:  input.ProductID,
		Quantity:   input.Quantity,
		TotalPrice: input.Quantity * product.Price,
		Status:     StatusPending,
	}

	// Kurangi stok lalu simpan lewat repo (repo sudah handle transaction)
	product.Stock -= input.Quantity
	if err := s.repo.Create(ctx, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *Service) GetAllOrders(ctx context.Context) ([]Order, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetOrderByID(ctx context.Context, id uint) (*Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrOrderNotFound)
		}
		return nil, err
	}
	return &order, nil
}

func (s *Service) UpdateOrder(ctx context.Context, id uint, input UpdateOrderRequest, userID uint) (*Order, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrOrderNotFound)
		}
		return nil, err
	}

	if order.UserID != userID {
		return nil, errors.New(ErrNotAuthorizedToUpdate)
	}

	if err := s.validateStatusTransition(&order, input.Status); err != nil {
		return nil, err
	}

	if input.Quantity != nil && input.Status != nil {
		return nil, errors.New(ErrCannotUpdateBothQuantityAndStatus)
	}

	if input.Status != nil {
		return s.updateOrderStatus(ctx, &order, *input.Status)
	}

	if input.Quantity != nil {
		return s.updateOrderQuantity(ctx, &order, *input.Quantity)
	}

	return &order, nil
}

func (s *Service) DeleteOrder(ctx context.Context, id uint) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrOrderNotFound)
		}
		return err
	}

	product, err := s.productRepo.FindByID(ctx, order.ProductID)
	if err != nil {
		return err
	}

	product.Stock += order.Quantity
	return s.repo.Delete(ctx, id)
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

// Helpers
func (s *Service) validateStatusTransition(order *Order, newStatus *OrderStatus) error {
	if newStatus == nil {
		return nil
	}
	switch *newStatus {
	case StatusPending, StatusPaid, StatusCancelled:
	default:
		return errors.New(ErrInvalidStatusValue)
	}
	if order.Status == StatusPaid && *newStatus == StatusPending {
		return errors.New(ErrCannotChangePaidOrderToPending)
	}
	if order.Status == StatusCancelled && *newStatus != StatusCancelled {
		return errors.New(ErrCannotChangeCancelledOrderStatus)
	}
	return nil
}

func (s *Service) updateOrderStatus(ctx context.Context, order *Order, newStatus OrderStatus) (*Order, error) {
	if newStatus == StatusCancelled && order.Status != StatusCancelled {
		product, err := s.productRepo.FindByID(ctx, order.ProductID)
		if err != nil {
			return nil, err
		}
		product.Stock += order.Quantity
		if err := s.repo.Update(ctx, order, &product, func(o *Order) {
			o.Status = newStatus
		}); err != nil {
			return nil, err
		}
		return order, nil
	}

	if err := s.repo.Update(ctx, order, nil, func(o *Order) {
		o.Status = newStatus
	}); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *Service) updateOrderQuantity(ctx context.Context, order *Order, newQuantity int) (*Order, error) {
	product, err := s.productRepo.FindByID(ctx, order.ProductID)
	if err != nil {
		return nil, err
	}

	diff := newQuantity - order.Quantity
	if diff > 0 && diff > product.Stock {
		return nil, errors.New(ErrInsufficientStockForUpdate)
	}

	product.Stock -= diff
	if err := s.repo.Update(ctx, order, &product, func(o *Order) {
		o.Quantity = newQuantity
		o.TotalPrice = newQuantity * product.Price
	}); err != nil {
		return nil, err
	}

	return order, nil
}

