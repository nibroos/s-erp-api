BEGIN;

INSERT INTO 
  mix_values (
    group_id, 
    name, 
    description, 
    remark, 
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
      'IN IMPORT', 
      'Import goods from suppliers', 
      'INVENTORY_IN', 
      1,
      '{"code": "INV-IN-IMPORT", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_IN"}',
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
      'IN REJECT', 
      'Rejected items returned to inventory', 
      'INVENTORY_IN', 
      1,
      '{"code": "INV-IN-REJECT", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_IN"}',
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
      'IN SCRAP', 
      'Scrap materials returned to inventory', 
      'INVENTORY_IN', 
      1,
      '{"code": "INV-IN-SCRAP", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_IN"}',
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
      'IN STOCK AWAL', 
      'Initial stock entry into inventory', 
      'INVENTORY_IN', 
      1,
      '{"code": "INV-IN-BASE", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_IN"}',
      CURRENT_TIMESTAMP, 
      CURRENT_TIMESTAMP
    ),
    (
      (
        SELECT 
        id 
        FROM 
        groups 
        WHERE name = 'ingoing_types' 
        LIMIT 
        1
      ),
      'IN SUPPLIER', 
      'Items received from suppliers', 
      'INVENTORY_IN', 
      1,
      '{"code": "INV-IN-SUPPLIER", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_IN"}',
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
      'IN WAREHOUSE', 
      'Inter-warehouse transfer receipts', 
      'INVENTORY_IN', 
      1,
      '{"code": "INV-WAREHOUSE", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_IN"}',
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
      'IN SUBCON', 
      'Items received from subcontractors', 
      'INVENTORY_IN', 
      1,
      '{"code": "INV-SUBCON", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_IN"}',
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
      'OUT BUYER', 
      'Items shipped to buyers', 
      'INVENTORY_OUT', 
      1,
      '{"code": "INV-OUT-BUYER", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_OUT"}',
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
      'OUT SCRAP', 
      'Scrapped items removed from inventory', 
      'INVENTORY_OUT', 
      1,
      '{"code": "INV-OUT-SCRAP", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_OUT"}',
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
      'OUT SUBCON', 
      'Items sent to subcontractors', 
      'INVENTORY_OUT', 
      1,
      '{"code": "INV-OUT-SUBCON", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_OUT"}',
      CURRENT_TIMESTAMP, 
      CURRENT_TIMESTAMP
    ),
    (
      (SELECT 
        id 
        FROM 
        groups 
        WHERE 
        name = 'outgoing_types' 
        LIMIT 
        1
      ),
      'OUT SUPPLIER', 
      'Items returned to suppliers', 
      'INVENTORY_OUT', 
      1,
      '{"code": "INV-OUT-SUPPLIER", 
      "type": "INVENTORY", 
      "io_type": "INVENTORY_OUT"}',
      CURRENT_TIMESTAMP, 
      CURRENT_TIMESTAMP
    );

COMMIT;
