package swagger

// Standard Swagger annotation templates for consistent documentation

// Template for POST endpoints (Create operations)
const CreateTemplate = `
// {FunctionName} godoc
// @Summary {Summary}
// @Description {Description}
// @Tags {Tag}
// @Accept  json
// @Produce  json
// @Param   request body {RequestStruct} true "{RequestDescription}"
// @Success 201 {object} response.SuccessResponse{data={ResponseStruct}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router {Route} [post]
`

// Template for GET endpoints (Read operations - single item)
const GetSingleTemplate = `
// {FunctionName} godoc
// @Summary {Summary}
// @Description {Description}
// @Tags {Tag}
// @Accept  json
// @Produce  json
// @Param   id path string true "{IdDescription}"
// @Success 200 {object} response.SuccessResponse{data={ResponseStruct}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router {Route} [get]
`

// Template for GET endpoints (Read operations - list)
const GetListTemplate = `
// {FunctionName} godoc
// @Summary {Summary}
// @Description {Description}
// @Tags {Tag}
// @Accept  json
// @Produce  json
// @Success 200 {object} response.SuccessResponse{data=[]${ResponseStruct}}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router {Route} [get]
`

// Template for PUT endpoints (Update operations)
const UpdateTemplate = `
// {FunctionName} godoc
// @Summary {Summary}
// @Description {Description}
// @Tags {Tag}
// @Accept  json
// @Produce  json
// @Param   id path string true "{IdDescription}"
// @Param   request body {RequestStruct} true "{RequestDescription}"
// @Success 200 {object} response.SuccessResponse{data={ResponseStruct}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router {Route} [put]
`

// Template for PATCH endpoints (Partial update operations)
const PatchTemplate = `
// {FunctionName} godoc
// @Summary {Summary}
// @Description {Description}
// @Tags {Tag}
// @Accept  json
// @Produce  json
// @Param   id path string true "{IdDescription}"
// @Param   request body {RequestStruct} true "{RequestDescription}"
// @Success 200 {object} response.SuccessResponse{data={ResponseStruct}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router {Route} [patch]
`

// Template for DELETE endpoints
const DeleteTemplate = `
// {FunctionName} godoc
// @Summary {Summary}
// @Description {Description}
// @Tags {Tag}
// @Accept  json
// @Produce  json
// @Param   id path string true "{IdDescription}"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router {Route} [delete]
`

// Template for POST endpoints without authentication
const CreatePublicTemplate = `
// {FunctionName} godoc
// @Summary {Summary}
// @Description {Description}
// @Tags {Tag}
// @Accept  json
// @Produce  json
// @Param   request body {RequestStruct} true "{RequestDescription}"
// @Success 201 {object} response.SuccessResponse{data={ResponseStruct}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router {Route} [post]
`

/*
Usage Examples:

1. For POST /products (authenticated):
// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with name, price, and stock
// @Tags Products
// @Accept  json
// @Produce  json
// @Param   request body CreateProductRequest true "Product request body"
// @Success 201 {object} response.SuccessResponse{data=Product}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products [post]

2. For GET /products/{id}:
// GetProductByID godoc
// @Summary Get product by ID
// @Description Retrieve a specific product by its ID
// @Tags Products
// @Accept  json
// @Produce  json
// @Param   id path string true "Product ID"
// @Success 200 {object} response.SuccessResponse{data=Product}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products/{id} [get]

3. For GET /products (list):
// GetAllProducts godoc
// @Summary Get all products
// @Description Get a list of all products
// @Tags Products
// @Accept  json
// @Produce  json
// @Success 200 {object} response.SuccessResponse{data=[]Product}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products [get]

Standard Tags to use:
- Authentication (Auth, Users)
- Products
- Orders
- Categories (if you add them later)

Standard HTTP Status Codes:
- 200: OK (GET, PUT, PATCH, DELETE)
- 201: Created (POST)
- 400: Bad Request (validation errors)
- 401: Unauthorized (authentication required)
- 404: Not Found
- 500: Internal Server Error
*/