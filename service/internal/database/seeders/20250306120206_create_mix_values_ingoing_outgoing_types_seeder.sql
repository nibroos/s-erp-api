BEGIN;

INSERT INTO
  mix_values (
    group_id,
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
        name = 'ingoing_types'
      LIMIT
        1
    ),
    'Purchase Order',
    'Ingoing document for purchasing goods',
    1,
    '{"code": "PO", "type": "IN"}',
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
        name = 'ingoing_types'
      LIMIT
        1
    ),
    'Goods Receipt',
    'Ingoing document for receiving goods',
    1,
    '{"code": "GR", "type": "IN"}',
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
        name = 'outgoing_types'
      LIMIT
        1
    ),
    'Sales Order',
    'Outgoing document for sales transaction',
    1,
    '{"code": "SO", "type": "OUT"}',
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
        name = 'outgoing_types'
      LIMIT
        1
    ),
    'Delivery Order',
    'Outgoing document for goods delivery',
    1,
    '{"code": "DO", "type": "OUT"}',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;
