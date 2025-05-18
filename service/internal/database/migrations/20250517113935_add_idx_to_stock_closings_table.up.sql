CREATE INDEX idx_stock_closings_warehouse_id ON stock_closings(warehouse_id);

CREATE INDEX idx_stock_closings_item_id ON stock_closings(item_id);

CREATE INDEX idx_stock_closings_closing_at ON stock_closings(closing_at);