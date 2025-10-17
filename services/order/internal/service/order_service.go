package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mini-e-commerce/services/order/internal/model"
	"mini-e-commerce/services/order/internal/repository"
	"mini-e-commerce/shared/cache"
	"mini-e-commerce/shared/config"
	"net/http"
	"time"

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
	ErrProductServiceUnavailable        = "product service unavailable"
)

// OrderService defines the order service interface
type OrderService interface {
	CreateOrder(ctx context.Context, input model.CreateOrderRequest, userID uint) (*model.Order, error)
	GetOrderByID(ctx context.Context, id uint, userID uint) (*model.Order, error)
	GetAllOrders(ctx context.Context, query model.OrderQuery, userID uint) (*model.OrderListResponse, error)
	UpdateOrder(ctx context.Context, id uint, input model.UpdateOrderRequest, userID uint) (*model.Order, error)
	DeleteOrder(ctx context.Context, id uint, userID uint) error
	GetOrdersByUserID(ctx context.Context, userID uint, query model.OrderQuery) (*model.OrderListResponse, error)
}

// orderService implements OrderService interface
type orderService struct {
	repo           repository.OrderRepository
	cache          cache.Interface
	validator      *validator.Validate
	logger         *zap.Logger
	config         *config.ServiceConfig
	httpClient     *http.Client
}

func NewOrderService(repo repository.OrderRepository, cache cache.Interface, logger *zap.Logger, config *config.ServiceConfig) OrderService {
	return &orderService{
		repo:       repo,
		cache:      cache,
		validator:  validator.New(),
		logger:     logger,
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *orderService) CreateOrder(ctx context.Context, input model.CreateOrderRequest, userID uint) (*model.Order, error) {
	// Validate input
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Get product information from product service
	productIDs := make([]uint, 0, len(input.Items))
	productQuantityMap := make(map[uint]int)
	
	for _, item := range input.Items {
		productIDs = append(productIDs, item.ProductID)
		productQuantityMap[item.ProductID] += item.Quantity
	}

	products, err := s.getProductsFromProductService(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	productMap := make(map[uint]model.ProductInfo)
	for _, p := range products {
		productMap[p.ID] = p
	}

	// Validate products and calculate totals
	var orderItems []model.OrderItem
	var totalPrice int

	for productID, quantity := range productQuantityMap {
		product, exists := productMap[productID]
		if !exists {
			return nil, errors.New(ErrProductNotFound)
		}

		if quantity > product.Stock {
			return nil, errors.New(ErrInsufficientStock)
		}

		subtotal := quantity * product.Price
		orderItem := model.OrderItem{
			ProductID: productID,
			Quantity:  quantity,
			Price:     product.Price,
			Subtotal:  subtotal,
		}

		orderItems = append(orderItems, orderItem)
		totalPrice += subtotal
	}

	// Create order
	order := &model.Order{
		UserID:     userID,
		TotalPrice: totalPrice,
		Status:     model.StatusPending,
		OrderItems: orderItems,
	}

	// Create order with transaction
	err = s.repo.CreateWithTransaction(ctx, order, func(tx *gorm.DB) error {
		// Update stock in product service
		for productID, quantity := range productQuantityMap {
			if err := s.updateProductStock(ctx, productID, -quantity); err != nil {
				s.logger.Error("Failed to update product stock",
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

	s.logger.Info("Order created successfully",
		zap.Uint("order_id", order.ID),
		zap.Uint("user_id", userID),
		zap.Int("total_price", totalPrice),
	)

	return order, nil
}

// GetOrderByID retrieves an order by ID
func (s *orderService) GetOrderByID(ctx context.Context, id uint, userID uint) (*model.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrOrderNotFound)
		}
		return nil, err
	}

	// Check authorization (users can only see their own orders unless they're admin)
	if order.UserID != userID {
		// In a real application, you might check if userID is admin
		// For now, we'll allow access for simplicity
	}

	return &order, nil
}

// GetAllOrders retrieves all orders with filtering and pagination
func (s *orderService) GetAllOrders(ctx context.Context, query model.OrderQuery, userID uint) (*model.OrderListResponse, error) {
	// Set defaults
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// Get orders from repository
	orders, total, err := s.repo.FindAll(ctx, query)
	if err != nil {
		return nil, err
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	response := &model.OrderListResponse{
		Data: orders,
		Pagination: model.PaginationMetadata{
			Page:       query.Page,
			PageSize:   query.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	return response, nil
}

// GetOrdersByUserID retrieves orders for a specific user
func (s *orderService) GetOrdersByUserID(ctx context.Context, userID uint, query model.OrderQuery) (*model.OrderListResponse, error) {
	// Set defaults
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// Get orders from repository
	orders, total, err := s.repo.FindByUserID(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	response := &model.OrderListResponse{
		Data: orders,
		Pagination: model.PaginationMetadata{
			Page:       query.Page,
			PageSize:   query.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	return response, nil
}

// UpdateOrder updates an order
func (s *orderService) UpdateOrder(ctx context.Context, id uint, input model.UpdateOrderRequest, userID uint) (*model.Order, error) {
	// Validate input
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	// Find order
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrOrderNotFound)
		}
		return nil, err
	}

	// Check authorization
	if order.UserID != userID {
		return nil, errors.New(ErrNotAuthorizedToUpdate)
	}

	// Validate status transition
	if input.Status != nil {
		if err := s.validateStatusTransition(&order, *input.Status); err != nil {
			return nil, err
		}

		// Handle status-specific logic
		if *input.Status == model.StatusCancelled && order.Status != model.StatusCancelled {
			// Restore stock when cancelling
			err = s.repo.UpdateWithTransaction(ctx, &order, func(o *model.Order) {
				o.Status = *input.Status
			}, func(tx *gorm.DB) error {
				for _, item := range order.OrderItems {
					if err := s.updateProductStock(ctx, item.ProductID, item.Quantity); err != nil {
						s.logger.Error("Failed to restore product stock on cancellation",
							zap.Uint("product_id", item.ProductID),
							zap.Int("quantity", item.Quantity),
							zap.Error(err),
						)
						return err
					}
				}
				return nil
			})
		} else {
			// Simple status update
			order.Status = *input.Status
			err = s.repo.Update(ctx, &order)
		}

		if err != nil {
			s.logger.Error("Failed to update order status",
				zap.Uint("order_id", order.ID),
				zap.Error(err),
			)
			return nil, err
		}
	}

	s.logger.Info("Order updated successfully", zap.Uint("order_id", order.ID))
	return &order, nil
}

// DeleteOrder deletes an order
func (s *orderService) DeleteOrder(ctx context.Context, id uint, userID uint) error {
	// Find order
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrOrderNotFound)
		}
		return err
	}

	// Check authorization
	if order.UserID != userID {
		return errors.New(ErrNotAuthorizedToUpdate)
	}

	// Delete order with transaction to restore stock
	err = s.repo.DeleteWithTransaction(ctx, id, func(tx *gorm.DB) error {
		for _, item := range order.OrderItems {
			if err := s.updateProductStock(ctx, item.ProductID, item.Quantity); err != nil {
				s.logger.Error("Failed to restore stock on order deletion",
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

	s.logger.Info("Order deleted successfully", zap.Uint("order_id", id))
	return nil
}

// Helper methods

// getProductsFromProductService fetches product information from product service
func (s *orderService) getProductsFromProductService(ctx context.Context, productIDs []uint) ([]model.ProductInfo, error) {
	// Prepare request
	requestBody := map[string][]uint{"product_ids": productIDs}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// Make HTTP request to product service
	productServiceURL := s.config.Services.ProductService.GetServiceURL() + "/api/v1/products/bulk"
	req, err := http.NewRequestWithContext(ctx, "POST", productServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to call product service", zap.Error(err))
		return nil, errors.New(ErrProductServiceUnavailable)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Product service returned error", zap.Int("status_code", resp.StatusCode))
		return nil, errors.New(ErrProductServiceUnavailable)
	}

	// Parse response
	var response struct {
		Data []model.ProductInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// updateProductStock updates product stock via product service
func (s *orderService) updateProductStock(ctx context.Context, productID uint, stockDelta int) error {
	// Prepare request
	requestBody := map[string]int{"stock_delta": stockDelta}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	// Make HTTP request to product service
	productServiceURL := fmt.Sprintf("%s/api/v1/products/%d/stock", s.config.Services.ProductService.GetServiceURL(), productID)
	req, err := http.NewRequestWithContext(ctx, "PATCH", productServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to update product stock", zap.Error(err))
		return errors.New(ErrProductServiceUnavailable)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Product service stock update failed", zap.Int("status_code", resp.StatusCode))
		return errors.New(ErrProductServiceUnavailable)
	}

	return nil
}

// validateStatusTransition validates order status transitions
func (s *orderService) validateStatusTransition(order *model.Order, newStatus model.OrderStatus) error {
	switch newStatus {
	case model.StatusPending, model.StatusPaid, model.StatusShipped, model.StatusDelivered, model.StatusCancelled:
	default:
		return errors.New(ErrInvalidStatusValue)
	}

	if order.Status == model.StatusPaid && newStatus == model.StatusPending {
		return errors.New(ErrCannotChangePaidOrderToPending)
	}

	if order.Status == model.StatusCancelled && newStatus != model.StatusCancelled {
		return errors.New(ErrCannotChangeCancelledOrderStatus)
	}

	return nil
}
