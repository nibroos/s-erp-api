BEGIN;

-- "id"	"name"	"conversion_unit"	"initial_value"	"conversion_rate"	"description"	"is_active"	"created_at"	"updated_at"	"deleted_at"	"log_json"	"created_by_id"	"updated_by_id"
-- "1"	"KG"	"Gr"	"1 Kg"	"1000.00000000"	"Kilogram"	"1"	"2024-03-27 02:07:22"	"2024-12-19 13:55:17"	\N	\N	\N	\N
-- "2"	"GRAM"	"Gr"	"1 Gr"	"0.00000000"	"Gram"	"1"	"2024-03-27 02:07:22"	"2024-12-19 13:54:53"	\N	\N	\N	\N
-- "3"	"CM"	"mm"	"1 Cm"	"100.00000000"	"Centimeter"	"1"	"2024-03-27 02:07:22"	"2024-12-19 13:54:41"	\N	\N	\N	\N
-- "4"	"METER"	"Cm"	"1 M"	"100.00000000"	"Meter"	"1"	"2024-03-27 02:07:22"	"2024-12-19 13:54:25"	\N	\N	\N	\N
-- "5"	"PIECE"	"Piece"	"1 Piece"	"0.00000000"	"Piece"	"1"	"2024-03-27 02:07:22"	"2024-12-19 13:54:13"	\N	\N	\N	\N
-- "6"	"YARD"	"M"	"1 Yds"	"0.91440000"	"Yard"	"1"	"2024-03-27 02:07:22"	"2024-12-19 13:53:50"	\N	\N	\N	\N
-- "7"	"CONE"	"KG"	"1 CONE"	"0.00000000"	\N	"1"	\N	\N	\N	\N	\N	\N
-- "8"	"ROLL"	"KG"	"1 ROLL"	"0.00000000"	\N	"1"	\N	\N	\N	\N	\N	\N
-- "9"	"EA"	"KG"	"1 EA"	"0.00000000"	\N	"1"	\N	"2024-10-19 08:45:01"	\N	\N	\N	\N
-- "10"	"SET"	"KG"	"1 SET"	"0.00000000"	\N	"1"	\N	\N	\N	\N	\N	\N
-- "11"	"PACK"	"KG"	"1 PACK"	"0.00000000"	\N	"1"	\N	\N	\N	\N	\N	\N
-- "12"	"BOX"	"KG"	"1 BOX"	"0.00000000"	\N	"1"	\N	\N	\N	\N	\N	\N
-- "13"	"DRUM"	"KG"	"1 DRUM"	"0.00000000"	\N	"0"	\N	\N	\N	\N	\N	\N
-- "14"	"CAN"	"KG"	"1 CAN"	"0.00000000"	\N	"1"	\N	"2024-10-30 13:34:38"	\N	\N	\N	\N
-- "15"	"P1"	"KG"	"1 P1"	"0.00000000"	\N	"1"	\N	\N	\N	\N	\N	\N
-- "16"	"LITER"	\N	\N	\N	\N	"1"	"2024-07-22 07:31:37"	"2024-12-19 13:53:34"	\N	\N	\N	\N
-- "17"	"UNIT"	\N	\N	\N	\N	"1"	"2024-07-25 08:54:11"	"2024-12-19 13:53:20"	\N	\N	\N	\N
-- "18"	"DOZEN"	\N	\N	\N	\N	"1"	"2024-10-30 13:35:10"	"2024-12-19 11:28:56"	\N	\N	\N	\N
-- "19"	"DRUM"	\N	\N	\N	\N	"1"	"2024-12-13 15:55:06"	"2024-12-19 13:55:53"	\N	\N	\N	\N
-- "20"	"CAN"	\N	\N	\N	\N	"1"	"2024-12-19 13:55:35"	"2024-12-19 13:55:46"	\N	\N	\N	\N
-- "21"	"ROLL"	\N	\N	\N	\N	"1"	"2024-12-19 13:56:24"	"2024-12-19 13:56:24"	\N	\N	\N	\N
-- "22"	"CONE"	\N	\N	\N	\N	"1"	"2024-12-19 13:56:37"	"2024-12-19 13:56:37"	\N	\N	\N	\N
-- "23"	"EA"	\N	\N	\N	\N	"1"	"2024-12-19 13:57:14"	"2024-12-19 13:57:22"	\N	\N	\N	\N
-- "24"	"PACK"	\N	\N	\N	\N	"1"	"2024-12-19 13:57:49"	"2024-12-19 13:57:49"	\N	\N	\N	\N
-- "25"	"RIM"	\N	\N	\N	\N	"1"	"2024-12-23 15:39:37"	"2025-01-06 15:15:34"	\N	\N	\N	\N
INSERT INTO
  mix_values (
    group_id,
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
        name = 'item_units'
    ),
    'KG',
    'Kilogram',
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
        name = 'item_units'
    ),
    'GRAM',
    'Gram',
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
        name = 'item_units'
    ),
    'CM',
    'Centimeter',
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
        name = 'item_units'
    ),
    'METER',
    'Meter',
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
        name = 'item_units'
    ),
    'PIECE',
    'Piece',
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
        name = 'item_units'
    ),
    'YARD',
    'Yard',
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
        name = 'item_units'
    ),
    'CONE',
    'Cone',
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
        name = 'item_units'
    ),
    'ROLL',
    'Roll',
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
        name = 'item_units'
    ),
    'EA',
    'EA',
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
        name = 'item_units'
    ),
    'SET',
    'Set',
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
        name = 'item_units'
    ),
    'PACK',
    'Pack',
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
        name = 'item_units'
    ),
    'BOX',
    'Box',
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
        name = 'item_units'
    ),
    'DRUM',
    'Drum',
    0,
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
        name = 'item_units'
    ),
    'CAN',
    'Can',
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
        name = 'item_units'
    ),
    'P1',
    'P1',
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
        name = 'item_units'
    ),
    'LITER',
    'Liter',
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
        name = 'item_units'
    ),
    'UNIT',
    'Unit',
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
        name = 'item_units'
    ),
    'DOZEN',
    'Dozen',
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
        name = 'item_units'
    ),
    'RIM',
    'Rim',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;

ROLLBACK;