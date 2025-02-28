BEGIN;

INSERT INTO
  branch_items (
    branch_id,
    item_unit_id,
    name,
    specification,
    description,
    tpb_code,
    minimum_stock,
    price_sell,
    price_buy,
    margin,
    status,
    created_by_id,
    created_at
  )
VALUES
  (
    1,
    1,
    'Item A',
    'Item A Specification',
    'Item A Description',
    '0001',
    2,
    2000,
    800,
    60,
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    1,
    'Item A1',
    'Item A1 Specification',
    'Item A1 Description',
    '0001-1',
    5,
    2300,
    1000,
    56.52,
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    4,
    'Item B1',
    'Item B1 Specification',
    'Item B1 Description',
    '0002',
    3,
    300,
    100,
    66.67,
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    6,
    'Item C1',
    'Item C1 Specification',
    'Item C1 Description',
    '0003',
    4,
    4000,
    2000,
    100,
    1,
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;