BEGIN;

INSERT INTO
  products (
    item_sub_group_id,
    item_unit_id,
    code,
    name,
    prod_type,
    specification,
    description,
    tpb_code,
    minimum_stock,
    is_all_branch,
    status,
    created_by_id,
    created_at,
    is_vat,
    is_pph23
  )
VALUES
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Barang Setengah Jadi'
      LIMIT
        1
    ), 1, '0001', 'Item A', 'single', 'Item A Specification', 'Item A Description', '0001', 2, 0, 1, 1, CURRENT_TIMESTAMP,
    1,
    0
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Produk'
      LIMIT
        1
    ), 4, '0002', 'Item B', 'single', 'Item B Specification', 'Item B Description', '0002', 5, 0, 1, 1, CURRENT_TIMESTAMP,
    1,
    0
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Produk'
      LIMIT
        1
    ), 6, '0003', 'Item C', 'single', 'Item C Specification', 'Item C Description', '0003', 3, 1, 1, 1, CURRENT_TIMESTAMP,
    1,
    0
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Peralatan Kantor'
      LIMIT
        1
    ), NULL,
    '0004',
    'Item D',
    'single',
    'Item D Specification',
    'Item D Description',
    '0004',
    1,
    0,
    1,
    1,
    CURRENT_TIMESTAMP,
    1,
    0
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Material'
      LIMIT
        1
    ), 7, '0005', 'Item E', 'single', 'Item E Specification', 'Item E Description', '0005', 7, 1, 1, 1, CURRENT_TIMESTAMP,
    1,
    0
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Produk'
      LIMIT
        1
    ), 8, '0006', 'SKYHAWK SEAGATE 4TB', 'single', 'SKYHAWK SEAGATE 4TB Specification', 'SKYHAWK SEAGATE 4TB Description', '0006', 1, 1, 1, 1, CURRENT_TIMESTAMP,
    1,
    0
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Produk'
      LIMIT
        1
    ), 9, '0007', 'HIKVISION DS-7208HQHI-K1/E', 'single', 'HIKVISION DS-7208HQHI-K1/E Specification', 'HIKVISION DS-7208HQHI-K1/E Description', '0007', 1, 1, 1, 1, CURRENT_TIMESTAMP,
    1,
    0
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Produk'
      LIMIT
        1
    ), 10, '0008', 'Intel Xeon E-2224 Gen 10', 'single', 'Intel Xeon E-2224 Gen 10 Specification', 'Intel Xeon E-2224 Gen 10 Description', '0008', 1, 1, 1, 1, CURRENT_TIMESTAMP,
    1,
    0
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Produk'
      LIMIT
        1
    ), 11, '0009', 'NVME 1TB ABC', 'single', 'NVME 1TB ABC Specification', 'NVME 1TB ABC Description', '0009', 1, 1, 1, 1, CURRENT_TIMESTAMP,
    1,
    0
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Produk'
      LIMIT
        1
    ), 12, '0010', 'RAM 32GB DDR4', 'single', 'RAM 32GB DDR4 Specification', 'RAM 32GB DDR4 Description', '0010', 1, 1, 1, 1, CURRENT_TIMESTAMP,
    0,
    1
  ),
  (
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Produk'
      LIMIT
        1
    ), 13, '0011', 'PSU 500W KKK', 'single', 'PSU 500W KKK Specification', 'PSU 500W KKK Desc', '0011', 1, 1, 1, 1, CURRENT_TIMESTAMP,
    1,
    1
  ),
  (
    (
      SELECT
        mix_values.id
      FROM
        mix_values
        JOIN groups ON mix_values.group_id = groups.id
      WHERE
        mix_values.name = 'Produk'
      LIMIT
        1
    ), 14, 'CD001', 'DVR 8CH + HDD 4TB', 'product', 'Spec DVR 8CH + HDD 4TB', 'Desc DVR 8CH + HDD 4TB', 'TP0012', 1, 1, 1, 1, CURRENT_TIMESTAMP,
    0,
    1
  ),
  (
    (
      SELECT
        mix_values.id
      FROM
        mix_values
        JOIN groups ON mix_values.group_id = groups.id
      WHERE
        mix_values.name = 'Produk'
      LIMIT
        1
    ), 15, 'CPC002', 'PC Server A', 'product', 'Specification PC Server A', 'Description PC Server A', 'TP0013', 1, 1, 1, 1, CURRENT_TIMESTAMP,
    1,
    1
  ),
  (
    (
      SELECT
        mix_values.id
      FROM
        mix_values
        JOIN groups ON mix_values.group_id = groups.id
      WHERE
        mix_values.name = 'Produk'
      LIMIT
        1
    ), 16, 'CP003', 'Product C', 'single', 'Product C Specification', 'Product C Description', 'TP0014', 1, 1, 1, 1, CURRENT_TIMESTAMP,
    1,
    0
  );

COMMIT;