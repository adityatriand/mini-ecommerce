package response

const (
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"

	ErrCodeDataNotFound      = "DATA_NOT_FOUND"
	ErrCodeDataAlreadyExists = "DATA_ALREADY_EXISTS"
	ErrCodeDataCreateFail    = "DATA_CREATE_FAILED"
	ErrCodeDataUpdateFail    = "DATA_UPDATE_FAILED"
	ErrCodeDataDeleteFail    = "DATA_DELETE_FAILED"

	ErrCodeValidationError = "VALIDATION_ERROR"
	ErrCodeDatabaseError   = "DATABASE_ERROR"
	ErrCodeInternalServer  = "INTERNAL_SERVER_ERROR"
)
