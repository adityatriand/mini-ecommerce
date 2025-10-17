package service

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"mini-e-commerce/shared/config"
	"mini-e-commerce/shared/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProxyService handles request proxying to backend services
type ProxyService struct {
	config     *config.ServiceConfig
	httpClient *http.Client
	logger     *zap.Logger
}

func NewProxyService(config *config.ServiceConfig, logger *zap.Logger) *ProxyService {
	return &ProxyService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// ProxyToUserService proxies requests to the user service
func (p *ProxyService) ProxyToUserService(c *gin.Context) {
	p.proxyRequest(c, p.config.Services.UserService.GetServiceURL())
}

// ProxyToProductService proxies requests to the product service
func (p *ProxyService) ProxyToProductService(c *gin.Context) {
	p.proxyRequest(c, p.config.Services.ProductService.GetServiceURL())
}

// ProxyToOrderService proxies requests to the order service
func (p *ProxyService) ProxyToOrderService(c *gin.Context) {
	p.proxyRequest(c, p.config.Services.OrderService.GetServiceURL())
}

// proxyRequest handles the actual request proxying
func (p *ProxyService) proxyRequest(c *gin.Context, targetURL string) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if exists {
		// Add user ID to headers for backend services
		c.Header("X-User-ID", fmt.Sprintf("%v", userID))
	}

	// Create target URL
	targetPath := c.Request.URL.Path
	targetQuery := c.Request.URL.RawQuery
	fullURL := targetURL + targetPath
	if targetQuery != "" {
		fullURL += "?" + targetQuery
	}

	// Create request
	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, fullURL, c.Request.Body)
	if err != nil {
		p.logger.Error("Failed to create proxy request", zap.Error(err))
		response.NewResponseHelper(p.logger).InternalError(c, "Failed to create request", err.Error())
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.logger.Error("Failed to proxy request", 
			zap.String("target", fullURL),
			zap.Error(err))
		response.NewResponseHelper(p.logger).ServiceUnavailable(c, "Service temporarily unavailable")
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Error("Failed to read response body", zap.Error(err))
		response.NewResponseHelper(p.logger).InternalError(c, "Failed to read response", err.Error())
		return
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set status code and body
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)

	p.logger.Info("Request proxied successfully",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("target", fullURL),
		zap.Int("status", resp.StatusCode))
}
