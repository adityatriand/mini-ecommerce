package order

import (
	"context"
	"errors"

	"mini-e-commerce/internal/product"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

const (
	ErrOrderNotFound                     = "order not found"
	ErrProductNotFound                   = "product not found"
	ErrInsufficientStock                 = "insufficient stock"
	ErrNotAuthorizedToUpdate             = "not authorized to update this order"
	ErrInvalidStatusValue                = "invalid status value"
	ErrCannotUpdateBothQuantityAndStatus = "cannot update both quantity and status simultaneously"
	ErrCannotChangePaidOrderToPending    = "cannot change paid order back to pending"
	ErrCannotChangeCancelledOrderStatus  = "cannot change cancelled order status"
	ErrInsufficientStockForUpdate        = "insufficient stock for quantity update"
)

type Service interface {
	CreateOrder(ctx context.Context, input CreateOrderRequest, userID uint) (*Order, error)
	GetAllOrders(ctx context.Context) ([]Order, error)
	GetOrderByID(ctx context.Context, id uint) (*Order, error)
	UpdateOrder(ctx context.Context, id uint, input UpdateOrderRequest, userID uint) (*Order, error)
	DeleteOrder(ctx context.Context, id uint) error
}

type service struct {
	repo           Repository
	productService product.Service
	validator      *validator.Validate
}

func NewService(repo Repository, productService product.Service) Service {
	return &service{
		repo:           repo,
		productService: productService,
		validator:      validator.New(),
	}
}

func (s *service) CreateOrder(ctx context.Context, input CreateOrderRequest, userID uint) (*Order, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	product, err := s.productService.GetProductByID(ctx, input.ProductID)
	if err != nil {
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

	// Reduce stock using product service
	if err := s.productService.UpdateStock(ctx, input.ProductID, -input.Quantity); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, &order); err != nil {
		// Rollback stock if order creation fails
		s.productService.UpdateStock(ctx, input.ProductID, input.Quantity)
		return nil, err
	}

	return &order, nil
}

func (s *service) GetAllOrders(ctx context.Context) ([]Order, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) GetOrderByID(ctx context.Context, id uint) (*Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrOrderNotFound)
		}
		return nil, err
	}
	return &order, nil
}

func (s *service) UpdateOrder(ctx context.Context, id uint, input UpdateOrderRequest, userID uint) (*Order, error) {
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

func (s *service) DeleteOrder(ctx context.Context, id uint) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrOrderNotFound)
		}
		return err
	}

	// Restore stock when deleting order
	if err := s.productService.UpdateStock(ctx, order.ProductID, order.Quantity); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

// Helpers
func (s *service) validateStatusTransition(order *Order, newStatus *OrderStatus) error {
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

func (s *service) updateOrderStatus(ctx context.Context, order *Order, newStatus OrderStatus) (*Order, error) {
	if newStatus == StatusCancelled && order.Status != StatusCancelled {
		// Restore stock when cancelling order
		if err := s.productService.UpdateStock(ctx, order.ProductID, order.Quantity); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, order, func(o *Order) {
		o.Status = newStatus
	}); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *service) updateOrderQuantity(ctx context.Context, order *Order, newQuantity int) (*Order, error) {
	product, err := s.productService.GetProductByID(ctx, order.ProductID)
	if err != nil {
		return nil, err
	}

	diff := newQuantity - order.Quantity
	if diff > 0 && diff > product.Stock {
		return nil, errors.New(ErrInsufficientStockForUpdate)
	}

	// Update stock based on quantity difference
	if err := s.productService.UpdateStock(ctx, order.ProductID, -diff); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, order, func(o *Order) {
		o.Quantity = newQuantity
		o.TotalPrice = newQuantity * product.Price
	}); err != nil {
		return nil, err
	}

	return order, nil
}
