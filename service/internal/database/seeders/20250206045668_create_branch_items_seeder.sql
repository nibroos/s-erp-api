BEGIN;

INSERT INTO
  branch_items (
    branch_id,
    ms_item_id,
    name,
    specification,
    description,
    tpb_code,
    minimum_stock,
    price_sell,
    price_buy,
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
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    2,
    'Item B',
    'Item B Specification',
    'Item B Description',
    '0002',
    3,
    300,
    100,
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    3,
    'Item C',
    'Item C Specification',
    'Item C Description',
    '0003',
    4,
    4000,
    2000,
    1,
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;