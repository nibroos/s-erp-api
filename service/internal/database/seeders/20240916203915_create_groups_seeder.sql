BEGIN;

INSERT INTO
  groups (
    id,
    name,
    description,
    remark,
    status,
    options_json,
    created_by_id,
    updated_by_id
  )
VALUES
  (
    1,
    'roles',
    'Roles',
    'Roles table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    2,
    'permissions',
    'Permissions',
    'Permissions table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    3,
    'users',
    'Users',
    'Users table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    4,
    'identifiers',
    'Identifier',
    'Identifier (ID)',
    1,
    '{}',
    1,
    1
  ),
  (
    5,
    'addresses',
    'Addresses',
    'Alamat',
    1,
    '{}',
    1,
    1
  ),
  (
    6,
    'contacts',
    'Contacts',
    'Contacts table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    7,
    'customer_types',
    'Customer types',
    'Customer types table for storing user data. Ex: Buyer, Agent, Vendor, Supplier, etc',
    1,
    '{}',
    1,
    1
  ),
  (
    8,
    'currencies',
    'Currencies',
    'Currencies table for storing user data. Ex: USD, EUR, etc',
    1,
    '{}',
    1,
    1
  ),
  (
    9,
    'item_groups',
    'Item groups',
    'Item groups table for storing user data. Ex: Office, Production, etc',
    1,
    '{}',
    1,
    1
  ),
  (
    10,
    'units',
    'Item units',
    'Item units table for storing user data. Ex: Kg, Ltr, Pcs etc',
    1,
    '{}',
    1,
    1
  ),
  (
    11,
    'item_sub_groups',
    'Item sub groups',
    'Item sub groups table for storing user data. Ex: Finished goods, Semi-finished goods, Raw materials etc',
    1,
    '{}',
    1,
    1
  ),
  (
    12,
    'vats',
    'VATs',
    'VATs table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    13,
    'pph23s',
    'PPH23s',
    'PPH23s table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    14,
    'cap_types',
    'Cap types',
    'Cap types table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    15,
    'cap_sizes',
    'Cap sizes',
    'Cap sizes table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    16,
    'cap_categories',
    'Cap categories',
    'Cap categories table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    17,
    'cap_colors',
    'Cap colors',
    'Cap colors table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    18,
    'cap_statuses',
    'Cap statuses',
    'Cap statuses table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    19,
    'cutting_types',
    'Cutting types',
    'Cutting types table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    20,
    'collections',
    'Collections',
    'Collections table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    21,
    'process_statuses',
    'Process statuses',
    'Process statuses table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    22,
    'finishing_types',
    'Finishing types',
    'Finishing types table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    23,
    'packing_methods',
    'Packing methods',
    'Packing methods table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    24,
    'unit_conversions',
    'Unit conversions',
    'Unit conversions table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    25,
    'shipping_terms',
    'Shipping terms',
    'Shipping terms table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    26,
    'payment_terms',
    'Payment terms',
    'Payment terms table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    27,
    'purchase_types',
    'Purchase types',
    'Purchase types table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    28,
    'order_types',
    'Order types',
    'Order types table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    29,
    'production_types',
    'Production types',
    'Production types table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    30,
    'colors',
    'Colors',
    'Colors table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    31,
    'color_processes',
    'Color processes',
    'Color processes table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    32,
    'lines',
    'Lines',
    'Lines rack',
    1,
    '{}',
    1,
    1
  ),
  (
    33,
    'identifiers',
    'Identifier (ID)',
    'KTP, NPWP, SIM, Passport, etc',
    1,
    '{}',
    1,
    1
  ),
  (
    34,
    'contacts',
    'Contacts',
    'Rumah, Kantor, Saudara, etc',
    1,
    '{}',
    1,
    1
  ),
  (
    35,
    'addresses',
    'Addresses',
    'Rumah, Kantor, Saudara, etc',
    1,
    '{}',
    1,
    1
  ),
  (
    36,
    'ingoing_types',
    'Ingoing types',
    'Ingoing types table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    37,
    'outgoing_types',
    'Outgoing types',
    'Outgoing types table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    38,
    'warehouses',
    'Warehouses',
    'Warehouses table for storing user data',
    1,
    '{}',
    1,
    1
  ),
  (
    39,
    'tasks',
    'Tasks',
    'Tasks table for storing tasks master data',
    1,
    '{}',
    1,
    1
  ),
  (
    40,
    'steps',
    'Steps',
    'Steps table for storing steps master data',
    1,
    '{}',
    1,
    1
  ),
  (
    41,
    'category_types',
    'Category types',
    'Category types table for storing category types master data',
    1,
    '{}',
    1,
    1
  ),
  (
    42,
    'payment_types',
    'Payment types',
    'Payment types table for storing payment types master data',
    1,
    '{}',
    1,
    1
  );

-- vats, pph23s, cap_types, cap_sizes, cap_categories, cap_colors, cap_statuses
-- cutting_types, collections, Process Status, finishing_types, packing_methode
-- units, unit_conversions
-- shipping_terms, payment_terms, purchase_types, order_types, production_types
-- colors, color_processes, lines
COMMIT;

ROLLBACK;