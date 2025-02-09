BEGIN;

-- users, role_permissions, company_profiles, customers, customer_types, warehouses, currencies,
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
    ), 'role_permissions', 'Role Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'company_profiles', 'Company Profile Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'customer_types', 'Customer Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'currencies', 'Currency Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'ingoing_types', 'Ingoing Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'outgoing_types', 'Outgoing Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'order_types', 'Order Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'units', 'Unit Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'shipping_terms', 'Shipping Term Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'payment_terms', 'Payment Term Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'purchase_types', 'Purchase Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'production_types', 'Production Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'colors', 'Color Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'item_groups', 'Item Group Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'item_sub_groups', 'Item Sub Group Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'vats', 'Vat Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'pph23s', 'Pph23 Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'cap_types', 'Cap Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'cap_sizes', 'Cap Size Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'cap_categories', 'Cap Category Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'cap_colors', 'Cap Color Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'cutting_types', 'Cutting Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'collections', 'Collection Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'finishing_types', 'Finishing Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'packing_methode', 'Packing Methode Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'unit_conversions', 'Unit Conversion Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'color_methods', 'Color Method Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'color_processes', 'Color Process Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'sales_orders', 'Sales Order Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'purchase_orders', 'Purchase Order Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'inventories', 'Inventory Permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'doc_types', 'Doc Type Permissions', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;