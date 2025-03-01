BEGIN;

INSERT INTO
  boms (
    product_id,
    product_item_id,
    item_unit_id,
    qty,
    remark,
    created_by_id,
    created_at
  )
VALUES
  (
    12,
    6,
    8,
    1,
    'Re - SKYHAWK SEAGATE 4TB',
    1,
    CURRENT_TIMESTAMP
  ),
  (
    12,
    7,
    9,
    1,
    'Re - HIKVISION DS - 7208HQHI - K1 / E',
    1,
    CURRENT_TIMESTAMP
  ),
  (
    13,
    8,
    10,
    1,
    'Re - Intel Xeon E - 2224 Gen 10',
    1,
    CURRENT_TIMESTAMP
  ),
  (
    13,
    9,
    11,
    2,
    'Re - NVME 1TB ABC',
    1,
    CURRENT_TIMESTAMP
  ),
  (
    13,
    10,
    12,
    1,
    'Re - RAM 32GB DDR4',
    1,
    CURRENT_TIMESTAMP
  ),
  (
    13,
    11,
    13,
    1,
    'Re - PSU 500W KKK',
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;