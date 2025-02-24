BEGIN;

-- users, role_permissions, branches, customers, customer_types, warehouses, currencies,
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
    ), 'c_users', 'Permission to create users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'r_users', 'Permission to read users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'u_users', 'Permission to update users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'd_users', 'Permission to delete users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'rs_users', 'Permission to restore users', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'c_roles', 'Permission to create role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'r_roles', 'Permission to read role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'u_roles', 'Permission to update role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'd_roles', 'Permission to delete role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'rs_roles', 'Permission to restore role permissions', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'branches'
      LIMIT
        1
    ), 'c_branches', 'Permission to create company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'branches'
      LIMIT
        1
    ), 'r_branches', 'Permission to read company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'branches'
      LIMIT
        1
    ), 'u_branches', 'Permission to update company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'branches'
      LIMIT
        1
    ), 'd_branches', 'Permission to delete company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'branches'
      LIMIT
        1
    ), 'rs_branches', 'Permission to restore company profiles', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'c_customers', 'Permission to create customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'r_customers', 'Permission to read customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'u_customers', 'Permission to update customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'd_customers', 'Permission to delete customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'rs_customers', 'Permission to restore customers', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'c_masters', 'Permission to create masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'r_masters', 'Permission to read masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'u_masters', 'Permission to update masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'd_masters', 'Permission to delete masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'rs_masters', 'Permission to restore masters', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'c_items', 'Permission to create items', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'r_items', 'Permission to read items', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'u_items', 'Permission to update items', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'd_items', 'Permission to delete items', 1, '{}', CURRENT_TIMESTAMP,
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
    ), 'rs_items', 'Permission to restore items', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'sos'
      LIMIT
        1
    ), 'c_sos', 'Permission to create sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'sos'
      LIMIT
        1
    ), 'r_sos', 'Permission to read sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'sos'
      LIMIT
        1
    ), 'u_sos', 'Permission to update sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'sos'
      LIMIT
        1
    ), 'd_sos', 'Permission to delete sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'sos'
      LIMIT
        1
    ), 'rs_sos', 'Permission to restore sales orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pos'
      LIMIT
        1
    ), 'c_pos', 'Permission to create purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pos'
      LIMIT
        1
    ), 'r_pos', 'Permission to read purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pos'
      LIMIT
        1
    ), 'u_pos', 'Permission to update purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pos'
      LIMIT
        1
    ), 'd_pos', 'Permission to delete purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pos'
      LIMIT
        1
    ), 'rs_pos', 'Permission to restore purchase orders', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'invs'
      LIMIT
        1
    ), 'c_invs', 'Permission to create inventories', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'invs'
      LIMIT
        1
    ), 'r_invs', 'Permission to read inventories', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'invs'
      LIMIT
        1
    ), 'u_invs', 'Permission to update inventories', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'invs'
      LIMIT
        1
    ), 'd_invs', 'Permission to delete inventories', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'invs'
      LIMIT
        1
    ), 'rs_invs', 'Permission to restore inventories', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;