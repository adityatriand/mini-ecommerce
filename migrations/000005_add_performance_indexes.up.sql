-- Performance optimization indexes
-- Add indexes for frequently queried columns and sorting operations

-- Products table indexes
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);
CREATE INDEX IF NOT EXISTS idx_products_updated_at ON products(updated_at);

-- Composite index for common product filtering scenarios
CREATE INDEX IF NOT EXISTS idx_products_price_stock ON products(price, stock);
CREATE INDEX IF NOT EXISTS idx_products_created_at_desc ON products(created_at DESC);

-- Orders table indexes
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_total_price ON orders(total_price);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_updated_at ON orders(updated_at);

-- Composite indexes for common order queries
CREATE INDEX IF NOT EXISTS idx_orders_user_status ON orders(user_id, status);
CREATE INDEX IF NOT EXISTS idx_orders_user_created_at ON orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status_created_at ON orders(status, created_at DESC);

-- Order items table indexes
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_order_items_quantity ON order_items(quantity);
CREATE INDEX IF NOT EXISTS idx_order_items_price ON order_items(price);
CREATE INDEX IF NOT EXISTS idx_order_items_created_at ON order_items(created_at);

-- Composite index for order item analysis
CREATE INDEX IF NOT EXISTS idx_order_items_product_quantity ON order_items(product_id, quantity);
