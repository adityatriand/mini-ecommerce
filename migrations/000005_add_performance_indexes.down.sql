-- Drop performance optimization indexes

-- Drop order items indexes
DROP INDEX IF EXISTS idx_order_items_product_quantity;
DROP INDEX IF EXISTS idx_order_items_created_at;
DROP INDEX IF EXISTS idx_order_items_price;
DROP INDEX IF EXISTS idx_order_items_quantity;
DROP INDEX IF EXISTS idx_order_items_product_id;

-- Drop orders indexes
DROP INDEX IF EXISTS idx_orders_status_created_at;
DROP INDEX IF EXISTS idx_orders_user_created_at;
DROP INDEX IF EXISTS idx_orders_user_status;
DROP INDEX IF EXISTS idx_orders_updated_at;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_total_price;
DROP INDEX IF EXISTS idx_orders_status;

-- Drop products indexes
DROP INDEX IF EXISTS idx_products_created_at_desc;
DROP INDEX IF EXISTS idx_products_price_stock;
DROP INDEX IF EXISTS idx_products_updated_at;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_stock;
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_name;
