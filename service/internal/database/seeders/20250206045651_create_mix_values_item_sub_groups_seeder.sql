BEGIN;

-- group_id	name	discount	unit	is_active
-- 2	Packaging	15	%	1
-- 2	Material	\N	\N	1
-- 2	Barang Jadi	\N	\N	1
-- 2	Barang Setengah Jadi	\N	\N	1
-- 2	Bahan Penolong	15	\N	1
-- 1	Peralatan Kantor	0	\N	1
-- 2	Fiber	0	\N	1
-- 2	Mesin	0	\N	1
-- 2	Peralatan Pabrik	0	\N	1
-- 2	Scrap	\N	\N	1
-- 2	Reject	\N	\N	1
INSERT INTO
  mix_values (
    group_id,
    parent_id,
    name,
    description,
    status,
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Packaging',
    'Packaging',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Material',
    'Material',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Barang Jadi',
    'Barang Jadi',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Barang Setengah Jadi',
    'Barang Setengah Jadi',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Bahan Penolong',
    'Bahan Penolong',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Office'
    ),
    'Peralatan Kantor',
    'Peralatan Kantor',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Fiber',
    'Fiber',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Mesin',
    'Mesin',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Peralatan Pabrik',
    'Peralatan Pabrik',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Scrap',
    'Scrap',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Reject',
    'Reject',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;