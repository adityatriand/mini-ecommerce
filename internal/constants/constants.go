package constants

import "net/http"

const (
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusNotFound            = http.StatusNotFound
	StatusInternalServerError = http.StatusInternalServerError
)

const (
	// Product messages
	ProductCreatedMessage    = "Product created successfully"
	ProductsRetrievedMessage = "Products retrieved successfully"
	ProductUpdatedMessage    = "Product updated successfully"
	ProductDeletedMessage    = "Product deleted successfully"
	ProductNotFoundMessage   = "Product not found"

	// Auth messages
	UserRegisteredMessage = "User registered successfully"
	LoginSuccessfulMessage = "Login successful"
	LogoutSuccessfulMessage = "Logout successful"

	// Order messages
	OrderCreatedMessage = "Order created successfully"
	OrdersRetrievedMessage = "Orders fetched successfully"
	OrderRetrievedMessage = "Order fetched successfully"
	OrderUpdatedMessage = "Order updated successfully"
	OrderDeletedMessage = "Order deleted successfully"

	// General messages
	ValidationErrorMessage     = "Validation error"
	InternalServerErrorMessage = "Internal server error"

	// Common Error Messages (repeated across handlers)
	InvalidInputMessage      = "Invalid input"
	UnauthorizedMessage      = "Unauthorized"
	InvalidIDMessage         = "Invalid ID"
	NotFoundMessage          = "Not found"
	FailedToCreateMessage    = "Failed to create"
	FailedToFetchMessage     = "Failed to fetch"
	FailedToUpdateMessage    = "Failed to update"
	FailedToDeleteMessage    = "Failed to delete"
	FailedToProcessMessage   = "Failed to process"

	// HTTP Status Messages
	BadRequestMessage        = "Bad request"
	ForbiddenMessage         = "Forbidden"

	// Error codes
	ErrorCodeValidation     = "VALIDATION_ERROR"
	ErrorCodeNotFound       = "NOT_FOUND"
	ErrorCodeInternalServer = "INTERNAL_SERVER_ERROR"
)