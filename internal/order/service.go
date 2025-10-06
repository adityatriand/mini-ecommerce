package order

import (
	"context"
	"errors"

	"mini-e-commerce/internal/dto"
	"mini-e-commerce/internal/logger"
	"mini-e-commerce/internal/product"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	ErrOrderNotFound                    = "order not found"
	ErrProductNotFound                  = "product not found"
	ErrInsufficientStock                = "insufficient stock"
	ErrNotAuthorizedToUpdate            = "not authorized to update this order"
	ErrInvalidStatusValue               = "invalid status value"
	ErrCannotChangePaidOrderToPending   = "cannot change paid order back to pending"
	ErrCannotChangeCancelledOrderStatus = "cannot change cancelled order status"

	DefaultPage      = 1
	DefaultPageSize  = 10
	MaxPageSize      = 100
	MinQuantity      = 1
	DefaultSortOrder = "desc"
	DefaultSortField = "created_at"
)

type Service interface {
	CreateOrder(ctx context.Context, input CreateOrderRequest, userID uint) (*Order, error)
	GetAllOrders(ctx context.Context) ([]Order, error)
	GetAllOrdersWithQuery(ctx context.Context, query OrderQuery) (*OrderListResponse, error)
	GetOrderByID(ctx context.Context, id uint) (*Order, error)
	UpdateOrder(ctx context.Context, id uint, input UpdateOrderRequest, userID uint) (*Order, error)
	DeleteOrder(ctx context.Context, id uint) error
}

type service struct {
	repo           Repository
	productService product.Service
	validator      *validator.Validate
	logger         logger.Logger
}

func NewService(repo Repository, productService product.Service, log logger.Logger) Service {
	return &service{
		repo:           repo,
		productService: productService,
		validator:      validator.New(),
		logger:         log,
	}
}

func (s *service) CreateOrder(ctx context.Context, input CreateOrderRequest, userID uint) (*Order, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var orderItems []OrderItem
	var totalPrice int
	stockUpdates := make(map[uint]int)

	for _, item := range input.Items {
		product, err := s.productService.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}

		subtotal := item.Quantity * product.Price
		orderItem := OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			Subtotal:  subtotal,
		}

		orderItems = append(orderItems, orderItem)
		totalPrice += subtotal
		stockUpdates[item.ProductID] += item.Quantity
	}

	for productID, totalQuantity := range stockUpdates {
		product, err := s.productService.GetProductByID(ctx, productID)
		if err != nil {
			return nil, err
		}
		if totalQuantity > product.Stock {
			return nil, errors.New(ErrInsufficientStock)
		}
	}

	order := Order{
		UserID:     userID,
		TotalPrice: totalPrice,
		Status:     StatusPending,
		OrderItems: orderItems,
	}

	err := s.repo.CreateWithTransaction(ctx, &order, func(tx *gorm.DB) error {
		for productID, quantity := range stockUpdates {
			if err := s.productService.UpdateStockWithTx(tx, productID, -quantity); err != nil {
				s.logger.Error("Failed to update stock in transaction",
					zap.Uint("product_id", productID),
					zap.Int("quantity", -quantity),
					zap.Error(err),
				)
				return err
			}
		}
		return nil
	})

	if err != nil {
		s.logger.Error("Order creation transaction failed",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
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

	if input.Status != nil {
		return s.updateOrderStatus(ctx, &order, *input.Status)
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

	err = s.repo.DeleteWithTransaction(ctx, id, func(tx *gorm.DB) error {
		for _, item := range order.OrderItems {
			if err := s.productService.UpdateStockWithTx(tx, item.ProductID, item.Quantity); err != nil {
				s.logger.Error("Failed to restore stock in transaction",
					zap.Uint("product_id", item.ProductID),
					zap.Int("quantity", item.Quantity),
					zap.Error(err),
				)
				return err
			}
		}
		return nil
	})

	if err != nil {
		s.logger.Error("Order deletion transaction failed",
			zap.Uint("order_id", id),
			zap.Error(err),
		)
		return err
	}

	return nil
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
		err := s.repo.UpdateWithTransaction(ctx, order, func(o *Order) {
			o.Status = newStatus
		}, func(tx *gorm.DB) error {
			for _, item := range order.OrderItems {
				if err := s.productService.UpdateStockWithTx(tx, item.ProductID, item.Quantity); err != nil {
					s.logger.Error("Failed to restore stock on cancellation",
						zap.Uint("product_id", item.ProductID),
						zap.Int("quantity", item.Quantity),
						zap.Error(err),
					)
					return err
				}
			}
			return nil
		})
		if err != nil {
			s.logger.Error("Order cancellation transaction failed",
				zap.Uint("order_id", order.ID),
				zap.Error(err),
			)
			return nil, err
		}
		return order, nil
	}

	if err := s.repo.Update(ctx, order, func(o *Order) {
		o.Status = newStatus
	}); err != nil {
		s.logger.Error("Failed to update order status",
			zap.Uint("order_id", order.ID),
			zap.Error(err),
		)
		return nil, err
	}

	return order, nil
}

func (s *service) GetAllOrdersWithQuery(ctx context.Context, query OrderQuery) (*OrderListResponse, error) {
	page := query.Page
	if page <= 0 {
		page = DefaultPage
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	order := query.Order
	if order != "asc" && order != "desc" {
		order = DefaultSortOrder
	}

	sortBy := query.SortBy
	validSortFields := map[string]bool{
		"id": true, "user_id": true, "product_id": true, "quantity": true, "total_price": true, "status": true, "created_at": true,
	}
	if sortBy != "" && !validSortFields[sortBy] {
		sortBy = DefaultSortField
	}

	offset := (page - 1) * pageSize

	orders, total, err := s.repo.FindAllWithPagination(ctx, offset, pageSize, sortBy, order)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	response := &OrderListResponse{
		Data: orders,
		Pagination: dto.PaginationMetadata{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	return response, nil
}
