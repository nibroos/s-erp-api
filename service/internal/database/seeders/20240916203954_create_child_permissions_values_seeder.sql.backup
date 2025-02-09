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
        name = 'customer_types'
      LIMIT
        1
    ), 'create_customer_types', 'Permission to create customer types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'customer_types'
      LIMIT
        1
    ), 'read_customer_types', 'Permission to read customer types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'customer_types'
      LIMIT
        1
    ), 'update_customer_types', 'Permission to update customer types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'customer_types'
      LIMIT
        1
    ), 'delete_customer_types', 'Permission to delete customer types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'customer_types'
      LIMIT
        1
    ), 'restore_customer_types', 'Permission to restore customer types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'currencies'
      LIMIT
        1
    ), 'create_currencies', 'Permission to create currencies', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'currencies'
      LIMIT
        1
    ), 'read_currencies', 'Permission to read currencies', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'currencies'
      LIMIT
        1
    ), 'update_currencies', 'Permission to update currencies', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'currencies'
      LIMIT
        1
    ), 'delete_currencies', 'Permission to delete currencies', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'currencies'
      LIMIT
        1
    ), 'restore_currencies', 'Permission to restore currencies', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'ingoing_types'
      LIMIT
        1
    ), 'create_ingoing_types', 'Permission to create ingoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'ingoing_types'
      LIMIT
        1
    ), 'read_ingoing_types', 'Permission to read ingoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'ingoing_types'
      LIMIT
        1
    ), 'update_ingoing_types', 'Permission to update ingoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'ingoing_types'
      LIMIT
        1
    ), 'delete_ingoing_types', 'Permission to delete ingoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'ingoing_types'
      LIMIT
        1
    ), 'restore_ingoing_types', 'Permission to restore ingoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'outgoing_types'
      LIMIT
        1
    ), 'create_outgoing_types', 'Permission to create outgoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'outgoing_types'
      LIMIT
        1
    ), 'read_outgoing_types', 'Permission to read outgoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'outgoing_types'
      LIMIT
        1
    ), 'update_outgoing_types', 'Permission to update outgoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'outgoing_types'
      LIMIT
        1
    ), 'delete_outgoing_types', 'Permission to delete outgoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'outgoing_types'
      LIMIT
        1
    ), 'restore_outgoing_types', 'Permission to restore outgoing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'order_types'
      LIMIT
        1
    ), 'create_order_types', 'Permission to create order types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'order_types'
      LIMIT
        1
    ), 'read_order_types', 'Permission to read', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'order_types'
      LIMIT
        1
    ), 'update_order_types', 'Permission to update order types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'order_types'
      LIMIT
        1
    ), 'delete_order_types', 'Permission to delete order types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'order_types'
      LIMIT
        1
    ), 'restore_order_types', 'Permission to restore order types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'units'
      LIMIT
        1
    ), 'create_units', 'Permission to create units', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'units'
      LIMIT
        1
    ), 'read_units', 'Permission to read units', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'units'
      LIMIT
        1
    ), 'update_units', 'Permission to update units', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'units'
      LIMIT
        1
    ), 'delete_units', 'Permission to delete units', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'units'
      LIMIT
        1
    ), 'restore_units', 'Permission to restore units', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'shipping_terms'
      LIMIT
        1
    ), 'create_shipping_terms', 'Permission to create shipping terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'shipping_terms'
      LIMIT
        1
    ), 'read_shipping_terms', 'Permission to read shipping terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'shipping_terms'
      LIMIT
        1
    ), 'update_shipping_terms', 'Permission to update shipping terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'shipping_terms'
      LIMIT
        1
    ), 'delete_shipping_terms', 'Permission to delete shipping terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'shipping_terms'
      LIMIT
        1
    ), 'restore_shipping_terms', 'Permission to restore shipping terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'payment_terms'
      LIMIT
        1
    ), 'create_payment_terms', 'Permission to create payment terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'payment_terms'
      LIMIT
        1
    ), 'read_payment_terms', 'Permission to read payment terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'payment_terms'
      LIMIT
        1
    ), 'update_payment_terms', 'Permission to update payment terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'payment_terms'
      LIMIT
        1
    ), 'delete_payment_terms', 'Permission to delete payment terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'payment_terms'
      LIMIT
        1
    ), 'restore_payment_terms', 'Permission to restore payment terms', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'purchase_types'
      LIMIT
        1
    ), 'create_purchase_types', 'Permission to create purchase types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'purchase_types'
      LIMIT
        1
    ), 'read_purchase_types', 'Permission to read purchase types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'purchase_types'
      LIMIT
        1
    ), 'update_purchase_types', 'Permission to update purchase types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'purchase_types'
      LIMIT
        1
    ), 'delete_purchase_types', 'Permission to delete purchase types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'purchase_types'
      LIMIT
        1
    ), 'restore_purchase_types', 'Permission to restore purchase types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'production_types'
      LIMIT
        1
    ), 'create_production_types', 'Permission to create production types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'production_types'
      LIMIT
        1
    ), 'read_production_types', 'Permission to read production types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'production_types'
      LIMIT
        1
    ), 'update_production_types', 'Permission to update production types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'production_types'
      LIMIT
        1
    ), 'delete_production_types', 'Permission to delete production types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'production_types'
      LIMIT
        1
    ), 'restore_production_types', 'Permission to restore production types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'colors'
      LIMIT
        1
    ), 'create_colors', 'Permission to create colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'colors'
      LIMIT
        1
    ), 'read_colors', 'Permission to read colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'colors'
      LIMIT
        1
    ), 'update_colors', 'Permission to update colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'colors'
      LIMIT
        1
    ), 'delete_colors', 'Permission to delete colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'colors'
      LIMIT
        1
    ), 'restore_colors', 'Permission to restore colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_groups'
      LIMIT
        1
    ), 'create_item_groups', 'Permission to create item groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_groups'
      LIMIT
        1
    ), 'read_item_groups', 'Permission to read item groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_groups'
      LIMIT
        1
    ), 'update_item_groups', 'Permission to update item groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_groups'
      LIMIT
        1
    ), 'delete_item_groups', 'Permission to delete item groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_groups'
      LIMIT
        1
    ), 'restore_item_groups', 'Permission to restore item groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_sub_groups'
      LIMIT
        1
    ), 'create_item_sub_groups', 'Permission to create item sub groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_sub_groups'
      LIMIT
        1
    ), 'read_item_sub_groups', 'Permission to read item sub groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_sub_groups'
      LIMIT
        1
    ), 'update_item_sub_groups', 'Permission to update item sub groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_sub_groups'
      LIMIT
        1
    ), 'delete_item_sub_groups', 'Permission to delete item sub groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'item_sub_groups'
      LIMIT
        1
    ), 'restore_item_sub_groups', 'Permission to restore item sub groups', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'vats'
      LIMIT
        1
    ), 'create_vats', 'Permission to create vats', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'vats'
      LIMIT
        1
    ), 'read_vats', 'Permission to read vats', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'vats'
      LIMIT
        1
    ), 'update_vats', 'Permission to update vats', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'vats'
      LIMIT
        1
    ), 'delete_vats', 'Permission to delete vats', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'vats'
      LIMIT
        1
    ), 'restore_vats', 'Permission to restore vats', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pph23s'
      LIMIT
        1
    ), 'create_pph23s', 'Permission to create pph23s', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pph23s'
      LIMIT
        1
    ), 'read_pph23s', 'Permission to read pph23s', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pph23s'
      LIMIT
        1
    ), 'update_pph23s', 'Permission to update pph23s', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pph23s'
      LIMIT
        1
    ), 'delete_pph23s', 'Permission to delete pph23s', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'pph23s'
      LIMIT
        1
    ), 'restore_pph23s', 'Permission to restore pph23s', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_types'
      LIMIT
        1
    ), 'create_cap_types', 'Permission to create cap types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_types'
      LIMIT
        1
    ), 'read_cap_types', 'Permission to read cap types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_types'
      LIMIT
        1
    ), 'update_cap_types', 'Permission to update cap types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_types'
      LIMIT
        1
    ), 'delete_cap_types', 'Permission to delete cap types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_types'
      LIMIT
        1
    ), 'restore_cap_types', 'Permission to restore cap types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_sizes'
      LIMIT
        1
    ), 'create_cap_sizes', 'Permission to create cap sizes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_sizes'
      LIMIT
        1
    ), 'read_cap_sizes', 'Permission to read cap sizes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_sizes'
      LIMIT
        1
    ), 'update_cap_sizes', 'Permission to update cap sizes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_sizes'
      LIMIT
        1
    ), 'delete_cap_sizes', 'Permission to delete cap sizes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_sizes'
      LIMIT
        1
    ), 'restore_cap_sizes', 'Permission to restore cap sizes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_categories'
      LIMIT
        1
    ), 'create_cap_categories', 'Permission to create cap categories', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_categories'
      LIMIT
        1
    ), 'read_cap_categories', 'Permission to read cap categories', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_categories'
      LIMIT
        1
    ), 'update_cap_categories', 'Permission to update cap categories', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_categories'
      LIMIT
        1
    ), 'delete_cap_categories', 'Permission to delete cap categories', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_categories'
      LIMIT
        1
    ), 'restore_cap_categories', 'Permission to restore cap categories', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_colors'
      LIMIT
        1
    ), 'create_cap_colors', 'Permission to create cap colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_colors'
      LIMIT
        1
    ), 'read_cap_colors', 'Permission to read cap colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_colors'
      LIMIT
        1
    ), 'update_cap_colors', 'Permission to update cap colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_colors'
      LIMIT
        1
    ), 'delete_cap_colors', 'Permission to delete cap colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cap_colors'
      LIMIT
        1
    ), 'restore_cap_colors', 'Permission to restore cap colors', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cutting_types'
      LIMIT
        1
    ), 'create_cutting_types', 'Permission to create cutting types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cutting_types'
      LIMIT
        1
    ), 'read_cutting_types', 'Permission to read cutting types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cutting_types'
      LIMIT
        1
    ), 'update_cutting_types', 'Permission to update cutting types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cutting_types'
      LIMIT
        1
    ), 'delete_cutting_types', 'Permission to delete cutting types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'cutting_types'
      LIMIT
        1
    ), 'restore_cutting_types', 'Permission to restore cutting types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'collections'
      LIMIT
        1
    ), 'create_collections', 'Permission to create collections', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'collections'
      LIMIT
        1
    ), 'read_collections', 'Permission to read collections', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'collections'
      LIMIT
        1
    ), 'update_collections', 'Permission to update collections', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'collections'
      LIMIT
        1
    ), 'delete_collections', 'Permission to delete collections', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'collections'
      LIMIT
        1
    ), 'restore_collections', 'Permission to restore collections', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'finishing_types'
      LIMIT
        1
    ), 'create_finishing_types', 'Permission to create finishing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'finishing_types'
      LIMIT
        1
    ), 'read_finishing_types', 'Permission to read finishing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'finishing_types'
      LIMIT
        1
    ), 'update_finishing_types', 'Permission to update finishing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'finishing_types'
      LIMIT
        1
    ), 'delete_finishing_types', 'Permission to delete finishing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'finishing_types'
      LIMIT
        1
    ), 'restore_finishing_types', 'Permission to restore finishing types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'packing_methods'
      LIMIT
        1
    ), 'create_packing_methods', 'Permission to create packing methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'packing_methods'
      LIMIT
        1
    ), 'read_packing_methods', 'Permission to read packing methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'packing_methods'
      LIMIT
        1
    ), 'update_packing_methods', 'Permission to update packing methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'packing_methods'
      LIMIT
        1
    ), 'delete_packing_methods', 'Permission to delete packing methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'packing_methods'
      LIMIT
        1
    ), 'restore_packing_methods', 'Permission to restore packing methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'unit_conversions'
      LIMIT
        1
    ), 'create_unit_conversions', 'Permission to create unit conversions', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'unit_conversions'
      LIMIT
        1
    ), 'read_unit_conversions', 'Permission to read unit conversions', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'unit_conversions'
      LIMIT
        1
    ), 'update_unit_conversions', 'Permission to update unit conversions', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'unit_conversions'
      LIMIT
        1
    ), 'delete_unit_conversions', 'Permission to delete unit conversions', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'unit_conversions'
      LIMIT
        1
    ), 'restore_unit_conversions', 'Permission to restore unit conversions', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_methods'
      LIMIT
        1
    ), 'create_color_methods', 'Permission to create color methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_methods'
      LIMIT
        1
    ), 'read_color_methods', 'Permission to read color methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_methods'
      LIMIT
        1
    ), 'update_color_methods', 'Permission to update color methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_methods'
      LIMIT
        1
    ), 'delete_color_methods', 'Permission to delete color methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_methods'
      LIMIT
        1
    ), 'restore_color_methods', 'Permission to restore color methods', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_processes'
      LIMIT
        1
    ), 'create_color_processes', 'Permission to create color processes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_processes'
      LIMIT
        1
    ), 'read_color_processes', 'Permission to read color processes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_processes'
      LIMIT
        1
    ), 'update_color_processes', 'Permission to update color processes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_processes'
      LIMIT
        1
    ), 'delete_color_processes', 'Permission to delete color processes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'color_processes'
      LIMIT
        1
    ), 'restore_color_processes', 'Permission to restore color processes', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'doc_types'
      LIMIT
        1
    ), 'create_doc_types', 'Permission to create doc types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'doc_types'
      LIMIT
        1
    ), 'read_doc_types', 'Permission to read doc types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'doc_types'
      LIMIT
        1
    ), 'update_doc_types', 'Permission to update doc types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'doc_types'
      LIMIT
        1
    ), 'delete_doc_types', 'Permission to delete doc types', 1, '{}', CURRENT_TIMESTAMP,
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
        name = 'doc_types'
      LIMIT
        1
    ), 'restore_doc_types', 'Permission to restore doc types', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;