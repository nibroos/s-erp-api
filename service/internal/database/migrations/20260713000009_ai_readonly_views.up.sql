-- Read-only data access for the chat AI assistant.
--
-- The assistant writes its own SELECT statements, so the safety boundary must
-- be enforced by PostgreSQL rather than by inspecting the generated SQL. Two
-- pieces do that:
--
--   1. schema `ai` — a curated set of views. Each one pre-filters soft-deleted
--      rows and joins human-readable names in, so the model never has to guess
--      at `deleted_at` or at lookup ids.
--   2. role `ai_readonly` — NOLOGIN, granted SELECT on those views and nothing
--      else. The application switches to it with SET LOCAL ROLE for the single
--      read-only transaction that runs the generated query. Because views are
--      executed with their owner's privileges, the role can read the views
--      without holding any privilege on `public`, so tables holding password
--      hashes, refresh tokens, HR records and private chat messages stay
--      unreachable even if the query text passes every application-side check.
--
-- The role deliberately has no password: there is no second credential to leak
-- or rotate, and SET LOCAL ROLE cannot outlive its transaction.

CREATE SCHEMA IF NOT EXISTS ai;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'ai_readonly') THEN
    CREATE ROLE ai_readonly NOLOGIN;
  END IF;
END
$$;

-- ---------------------------------------------------------------- master data

-- The trailing *_id / *_name aliases are deliberate redundancy. Item and
-- movement views expose product_id/product_name and customer_id/customer_name,
-- and a model reliably carries those names over to the master tables no matter
-- how firmly the prompt says not to. Accepting both spellings removes a whole
-- class of failed query instead of arguing with the model about it. (They are
-- appended rather than placed next to the originals so CREATE OR REPLACE keeps
-- working, which only permits adding columns at the end.)
CREATE OR REPLACE VIEW ai.customers AS
SELECT id, code, name, shortname, email, phone, address, pic_name, status, created_at,
       id AS customer_id, name AS customer_name
FROM public.customers
WHERE deleted_at IS NULL;

CREATE OR REPLACE VIEW ai.products AS
SELECT id, code, name, sku, barcode, prod_type, specification,
       minimum_stock, qty_stock, status, created_at,
       id AS product_id, name AS product_name, code AS product_code
FROM public.products
WHERE deleted_at IS NULL;

CREATE OR REPLACE VIEW ai.branches AS
SELECT id, name, address, phone, email, status
FROM public.branches
WHERE deleted_at IS NULL;

-- --------------------------------------------------------------------- sales

-- grand_total is the order value including tax; order_at is the business date
-- to group by for "sales per month" style questions (created_at is the row's
-- insert time and will mislead).
CREATE OR REPLACE VIEW ai.sales_orders AS
SELECT so.id,
       so.sales_order_no,
       so.customer_id,
       c.name AS customer_name,
       so.branch_id,
       so.status,
       so.order_at,
       so.shipping_at,
       so.due_at,
       so.total_qty,
       so.subtotal,
       so.total_discount,
       so.total_vat,
       so.grand_total,
       so.remark
FROM public.sales_orders so
LEFT JOIN public.customers c ON c.id = so.customer_id AND c.deleted_at IS NULL
WHERE so.deleted_at IS NULL;

CREATE OR REPLACE VIEW ai.sales_order_items AS
SELECT d.id,
       d.sales_order_id,
       so.sales_order_no,
       so.order_at,
       so.customer_id,
       c.name AS customer_name,
       d.item_id AS product_id,
       p.code AS product_code,
       p.name AS product_name,
       d.qty,
       d.price_sell,
       d.total_am AS line_total
FROM public.so_dts d
JOIN public.sales_orders so ON so.id = d.sales_order_id AND so.deleted_at IS NULL
LEFT JOIN public.products p ON p.id = d.item_id AND p.deleted_at IS NULL
LEFT JOIN public.customers c ON c.id = so.customer_id AND c.deleted_at IS NULL
WHERE d.deleted_at IS NULL;

CREATE OR REPLACE VIEW ai.quotations AS
SELECT q.id, q.quo_no, q.customer_id, c.name AS customer_name, q.title,
       q.status, q.due_at, q.expired_at, q.total_qty, q.subtotal, q.grand_total,
       q.created_at
FROM public.quotations q
LEFT JOIN public.customers c ON c.id = q.customer_id AND c.deleted_at IS NULL
WHERE q.deleted_at IS NULL;

-- ----------------------------------------------------------------- invoicing

CREATE OR REPLACE VIEW ai.sales_invoices AS
SELECT si.id, si.invoice_no, si.customer_id, c.name AS customer_name,
       si.invoice_date, si.due_date, si.status, si.total_qty, si.subtotal,
       si.total_vat, si.grand_total, si.title
FROM public.sales_invoices si
LEFT JOIN public.customers c ON c.id = si.customer_id AND c.deleted_at IS NULL
WHERE si.deleted_at IS NULL;

CREATE OR REPLACE VIEW ai.invoice_dps AS
SELECT dp.id, dp.invoice_no, dp.customer_id, c.name AS customer_name,
       dp.invoice_date, dp.due_date, dp.status, dp.dp_percentage,
       dp.subtotal, dp.total_vat, dp.grand_total, dp.title
FROM public.invoice_dps dp
LEFT JOIN public.customers c ON c.id = dp.customer_id AND c.deleted_at IS NULL
WHERE dp.deleted_at IS NULL;

-- ---------------------------------------------------------------- purchasing

-- purchase_orders.customer_id points at the supplier.
CREATE OR REPLACE VIEW ai.purchase_orders AS
SELECT po.id, po.po_no, po.customer_id AS supplier_id, c.name AS supplier_name,
       po.branch_id, po.status, po.po_date, po.delivery_date,
       po.total_qty, po.subtotal, po.total_vat, po.grand_total, po.remark
FROM public.purchase_orders po
LEFT JOIN public.customers c ON c.id = po.customer_id AND c.deleted_at IS NULL
WHERE po.deleted_at IS NULL;

CREATE OR REPLACE VIEW ai.purchase_order_items AS
SELECT d.id, d.po_id AS purchase_order_id, po.po_no, po.po_date,
       d.product_id, p.code AS product_code, p.name AS product_name,
       d.qty, d.price, d.total_amount AS line_total
FROM public.purchase_order_dts d
JOIN public.purchase_orders po ON po.id = d.po_id AND po.deleted_at IS NULL
LEFT JOIN public.products p ON p.id = d.product_id AND p.deleted_at IS NULL
WHERE d.deleted_at IS NULL;

CREATE OR REPLACE VIEW ai.request_orders AS
SELECT id, request_no, request_date, branch_id, warehouse_id, status,
       grand_total_req_qty, grand_total_wh_qty, remark, created_at
FROM public.request_orders
WHERE deleted_at IS NULL;

-- ----------------------------------------------------------------- inventory

-- io_type resolves the mix_values lookup; direction is IN or OUT so the model
-- does not have to know that group 36 is inbound and 37 outbound.
CREATE OR REPLACE VIEW ai.inventories AS
SELECT i.id,
       i.inventory_no,
       i.customer_id,
       c.name AS customer_name,
       mv.name AS io_type,
       CASE mv.group_id WHEN 36 THEN 'IN' WHEN 37 THEN 'OUT' END AS direction,
       i.warehouse_id,
       i.branch_id,
       i.status,
       i.do_at,
       i.ingoing_at,
       i.invoice_at,
       i.total_qty,
       i.subtotal,
       i.grand_total
FROM public.inventories i
LEFT JOIN public.mix_values mv ON mv.id = i.io_type_id
LEFT JOIN public.customers c ON c.id = i.customer_id AND c.deleted_at IS NULL
WHERE i.deleted_at IS NULL;

CREATE OR REPLACE VIEW ai.inventory_items AS
SELECT d.id, d.inventory_id, i.inventory_no, i.do_at,
       CASE mv.group_id WHEN 36 THEN 'IN' WHEN 37 THEN 'OUT' END AS direction,
       d.item_id AS product_id, p.code AS product_code, p.name AS product_name,
       d.qty, d.price_sell, d.price_buy, d.total_am AS line_total
FROM public.inv_dts d
JOIN public.inventories i ON i.id = d.inventory_id AND i.deleted_at IS NULL
LEFT JOIN public.mix_values mv ON mv.id = i.io_type_id
LEFT JOIN public.products p ON p.id = d.item_id AND p.deleted_at IS NULL
WHERE d.deleted_at IS NULL;

CREATE OR REPLACE VIEW ai.stocks AS
SELECT s.id, s.warehouse_id, s.branch_id, s.item_id AS product_id,
       p.code AS product_code, p.name AS product_name, s.qty
FROM public.stocks s
LEFT JOIN public.products p ON p.id = s.item_id AND p.deleted_at IS NULL
WHERE s.deleted_at IS NULL;

-- ------------------------------------------------------------------- support

CREATE OR REPLACE VIEW ai.tickets AS
SELECT t.id, t.ticket_no, t.customer_id, c.name AS customer_name,
       t.product_id, p.name AS product_name,
       t.priority_type, t.status, t.title, t.issue_desc, t.issue_solution,
       t.reported_at, t.created_at
FROM public.tickets t
LEFT JOIN public.customers c ON c.id = t.customer_id AND c.deleted_at IS NULL
LEFT JOIN public.products p ON p.id = t.product_id AND p.deleted_at IS NULL
WHERE t.deleted_at IS NULL;

-- ------------------------------------------------------------------- grants

GRANT USAGE ON SCHEMA ai TO ai_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA ai TO ai_readonly;
-- Any view added to `ai` later is readable too, without another grant step.
ALTER DEFAULT PRIVILEGES IN SCHEMA ai GRANT SELECT ON TABLES TO ai_readonly;

-- Belt and braces: the role must never be able to reach base tables directly.
-- (PostgreSQL grants no table privileges by default, so this only removes what
-- an earlier blanket GRANT may have handed to PUBLIC on these specific tables.)
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM ai_readonly;
REVOKE ALL ON SCHEMA public FROM ai_readonly;
