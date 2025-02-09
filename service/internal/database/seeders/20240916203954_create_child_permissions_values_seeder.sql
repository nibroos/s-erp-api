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
-- child permissions
INSERT INTO
  mix_values (
    group_id,
    parent_id,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'users'
      LIMIT
        1
    ), 'create_users', 'Permission to create users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'users'
      LIMIT
        1
    ), 'read_users', 'Permission to read users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'users'
      LIMIT
        1
    ), 'update_users', 'Permission to update users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'users'
      LIMIT
        1
    ), 'delete_users', 'Permission to delete users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'users'
      LIMIT
        1
    ), 'restore_users', 'Permission to restore users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'role_permissions'
      LIMIT
        1
    ), 'create_role_permissions', 'Permission to create role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'role_permissions'
      LIMIT
        1
    ), 'read_role_permissions', 'Permission to read role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'role_permissions'
      LIMIT
        1
    ), 'update_role_permissions', 'Permission to update role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'role_permissions'
      LIMIT
        1
    ), 'delete_role_permissions', 'Permission to delete role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'role_permissions'
      LIMIT
        1
    ), 'restore_role_permissions', 'Permission to restore role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'company_profiles'
      LIMIT
        1
    ), 'create_company_profiles', 'Permission to create company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'company_profiles'
      LIMIT
        1
    ), 'read_company_profiles', 'Permission to read company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'company_profiles'
      LIMIT
        1
    ), 'update_company_profiles', 'Permission to update company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'company_profiles'
      LIMIT
        1
    ), 'delete_company_profiles', 'Permission to delete company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'company_profiles'
      LIMIT
        1
    ), 'restore_company_profiles', 'Permission to restore company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'customers'
      LIMIT
        1
    ), 'create_customers', 'Permission to create customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'customers'
      LIMIT
        1
    ), 'read_customers', 'Permission to read customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'customers'
      LIMIT
        1
    ), 'update_customers', 'Permission to update customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'customers'
      LIMIT
        1
    ), 'delete_customers', 'Permission to delete customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'customers'
      LIMIT
        1
    ), 'restore_customers', 'Permission to restore customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'masters'
      LIMIT
        1
    ), 'create_masters', 'Permission to create masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'masters'
      LIMIT
        1
    ), 'read_masters', 'Permission to read masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'masters'
      LIMIT
        1
    ), 'update_masters', 'Permission to update masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'masters'
      LIMIT
        1
    ), 'delete_masters', 'Permission to delete masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'masters'
      LIMIT
        1
    ), 'restore_masters', 'Permission to restore masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'items'
      LIMIT
        1
    ), 'create_items', 'Permission to create items', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'items'
      LIMIT
        1
    ), 'read_items', 'Permission to read items', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'items'
      LIMIT
        1
    ), 'update_items', 'Permission to update items', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'items'
      LIMIT
        1
    ), 'delete_items', 'Permission to delete items', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'items'
      LIMIT
        1
    ), 'restore_items', 'Permission to restore items', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'sales_orders'
      LIMIT
        1
    ), 'create_sales_orders', 'Permission to create sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'sales_orders'
      LIMIT
        1
    ), 'read_sales_orders', 'Permission to read sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'sales_orders'
      LIMIT
        1
    ), 'update_sales_orders', 'Permission to update sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'sales_orders'
      LIMIT
        1
    ), 'delete_sales_orders', 'Permission to delete sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'sales_orders'
      LIMIT
        1
    ), 'restore_sales_orders', 'Permission to restore sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'purchase_orders'
      LIMIT
        1
    ), 'create_purchase_orders', 'Permission to create purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'purchase_orders'
      LIMIT
        1
    ), 'read_purchase_orders', 'Permission to read purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'purchase_orders'
      LIMIT
        1
    ), 'update_purchase_orders', 'Permission to update purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'purchase_orders'
      LIMIT
        1
    ), 'delete_purchase_orders', 'Permission to delete purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'purchase_orders'
      LIMIT
        1
    ), 'restore_purchase_orders', 'Permission to restore purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'inventories'
      LIMIT
        1
    ), 'create_inventories', 'Permission to create inventories', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'inventories'
      LIMIT
        1
    ), 'read_inventories', 'Permission to read inventories', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'inventories'
      LIMIT
        1
    ), 'update_inventories', 'Permission to update inventories', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'inventories'
      LIMIT
        1
    ), 'delete_inventories', 'Permission to delete inventories', 1, '{}', CURRENT_TIMESTAMP,
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
    ), (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'inventories'
      LIMIT
        1
    ), 'restore_inventories', 'Permission to restore inventories', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;