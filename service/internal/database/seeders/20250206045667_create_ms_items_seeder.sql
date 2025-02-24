BEGIN;

INSERT INTO
  ms_items (
    item_sub_group_id,
    item_unit_id,
    code,
    name,
    specification,
    description,
    tpb_code,
    minimum_stock,
    is_all_branch,
    status,
    created_by_id,
    created_at
  )
VALUES
  (
    85,
    1,
    '0001',
    'Item A',
    'Item A Specification',
    'Item A Description',
    '0001',
    2,
    0,
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    86,
    4,
    '0002',
    'Item B',
    'Item B Specification',
    'Item B Description',
    '0002',
    5,
    0,
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    87,
    6,
    '0003',
    'Item C',
    'Item C Specification',
    'Item C Description',
    '0003',
    3,
    1,
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    82,
    NULL,
    '0004',
    'Item D',
    'Item D Specification',
    'Item D Description',
    '0004',
    1,
    0,
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    83,
    7,
    '0005',
    'Item E',
    'Item E Specification',
    'Item E Description',
    '0005',
    7,
    1,
    1,
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;