BEGIN;

-- users, roles, branches, customers, customer_types, warehouses, currencies,
-- ingoing_types, outgoing_types, order_types, items,
-- units, shipping_terms, payment_terms, purchase_types 
-- production_types, colors, item_groups, item_sub_groups,
-- vats, pph23s, cap_types, cap_sizes, cap_categories, cap_colors,
-- cutting_types, collections, finishing_types, packing_methode, unit_conversions,
-- color_methods, color_processes, 
-- sales_orders, purchase_orders, 
-- inventories,
-- doc_types
-- parent permissions
INSERT INTO
  mix_values (
    group_id,
    name,
    description,
    status,
    options_json,
    created_at,
    updated_at
  )
VALUES
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'permissions'
      LIMIT
        1
    ), 'users', 'User Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'permissions'
      LIMIT
        1
    ), 'roles', 'Role Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'permissions'
      LIMIT
        1
    ), 'branches', 'Company Profile Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'permissions'
      LIMIT
        1
    ), 'customers', 'Customer Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'permissions'
      LIMIT
        1
    ), 'masters', 'Master Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'permissions'
      LIMIT
        1
    ), 'items', 'Item Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'permissions'
      LIMIT
        1
    ), 'sos', 'Sales Order Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'permissions'
      LIMIT
        1
    ), 'pos', 'Purchase Order Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'permissions'
      LIMIT
        1
    ), 'invs', 'Inventory Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;