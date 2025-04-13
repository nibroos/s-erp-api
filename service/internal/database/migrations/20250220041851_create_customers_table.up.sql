CREATE TABLE IF NOT EXISTS customers (
  id SERIAL PRIMARY KEY,
  customer_type_id INT NOT NULL REFERENCES mix_values(id) ON DELETE RESTRICT,
  agent_id INT REFERENCES customers(id) ON DELETE RESTRICT,
  currency_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  shortname TEXT,
  code TEXT,
  name VARCHAR(255) NOT NULL,
  address TEXT,
  phone TEXT,
  email TEXT,
  pic TEXT,
  status INT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_customers_customer_type_id ON customers(customer_type_id);

CREATE INDEX idx_customers_agent_id ON customers(agent_id);

CREATE INDEX idx_customers_currency_id ON customers(currency_id);